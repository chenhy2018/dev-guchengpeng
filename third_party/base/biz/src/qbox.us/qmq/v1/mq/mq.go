package mq

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qiniu/bytes"
	"github.com/qiniu/log.v1"
	"qbox.us/bufio"
	"qbox.us/cc"
	"qbox.us/errors"
	"qbox.us/largefile"
	"qbox.us/net/httputil"
	"qbox.us/qmq/qmqapi/v1/mq"
)

var (
	ENoSuchEntry = httputil.NewError(mq.NoSuchEntry, "no such file or directory")
)

// ---------------------------------------------------------------

/* MQ 系统设计
 *
 * 写指针(offPut): put 操作当前位置
 * 读指针(offGet): get 操作当前位置
 * 超时读指针（offTimeout)：第二 get 操作位置，当超时成立时可用
 */

type Instance struct {
	fp         cc.Writer // put
	fg         cc.Reader // get
	mutexp     sync.Mutex
	mutexg     sync.Mutex
	index      *os.File
	mutexi     sync.Mutex // for rw index
	mqf        *largefile.Instance
	offTimeout int64
	done       chan bool
	expires    uint32
	loop       bool
	doing      int64
	todo       int64
}

type indexFile struct {
	OffPut     int64
	OffGet     int64
	OffTimeout int64
	Expires    uint32
	Reserved   uint32
}

type recordHeader struct {
	MsgLen   uint16
	Tag      uint8
	State    uint8
	LockTime uint32 // in s
}

const (
	indexSize     = 32
	headerSize    = 8
	recordTag     = 0x0a
	offsetOfState = 3
)

const (
	stateNormal     = 0x01
	stateProcessing = 0x02
	stateDone       = 0x04

	stateTravelPut     = stateNormal | stateProcessing | stateDone
	stateTravelGet     = stateProcessing | stateDone
	stateTravelTimeout = stateDone
)

var (
	ErrInvalidRecord = errors.New("invalid record")
)

func chgState(mqf *largefile.Instance, msgId int64, state uint8) (err error) {

	_, err = mqf.WriteAt([]byte{state}, msgId+offsetOfState)
	if err != nil {
		err = errors.Info(err, "mq.chgState failed", msgId, state).Detail(err).Warn()
	}
	return
}

func lockMsg(mqf *largefile.Instance, msgId int64, expires uint32) (err error) {

	var b [5]byte
	b[0] = stateProcessing
	binary.LittleEndian.PutUint32(b[1:], uint32(time.Now().Unix())+expires)
	_, err = mqf.WriteAt(b[:], msgId+offsetOfState)
	if err != nil {
		err = errors.Info(err, "mq.lockMsg failed", msgId).Detail(err).Warn()
	}
	return
}

func (r *Instance) loopGetTimeout() {

	for r.loop {
		offGet := r.getOffGet()
		offTimeout, err := r.loopProcessingMsg(offGet)

		log.Debug("loopGetTimeout:", offTimeout, offGet)

		if err != nil {
			if err != errors.ENOENT {
				errors.Info(err, "mq.loopGetTimeout: loopProcessingMsg failed").Detail(err).Warn()
			}
			time.Sleep(time.Second)
		} else {
			atomic.StoreInt64(&r.offTimeout, offTimeout)
			offPut := r.getOffPut()

			log.Info("mq.Commit Index:", offPut, offGet, offTimeout)

			err = r.updateIndex(offPut, offGet, offTimeout)
			if err != nil {
				errors.Info(err, "mq.Commit Index failed").Detail(err).Warn()
			}
		}
	}
	r.done <- true
}

//
// TODO: if err != nil: how to fix this error? (通知人工介入)
//
func (r *Instance) loopProcessingMsg(offGet int64) (offTimeout int64, err error) {

	offTimeout = atomic.LoadInt64(&r.offTimeout)
	if offTimeout >= offGet {
		err = errors.ENOENT
		return
	}

	for offTimeout < offGet {

		log.Debug("loopProcessingMsg:", offTimeout, offGet)

		fb := &cc.Reader{ReaderAt: r.mqf, Offset: offTimeout}

		var h recordHeader
		err = binary.Read(fb, binary.LittleEndian, &h)
		if err != nil {
			err = errors.Info(err, "mq.loopProcessingMsg: read record header failed").Detail(err)
			return
		}

		if h.State == stateProcessing {
			currTime := uint32(time.Now().Unix())
			log.Debug("lockTime:", h.LockTime, currTime)
			if currTime < h.LockTime {
				time.Sleep(time.Duration(h.LockTime-currTime) * time.Second)
				continue
			}
			msg := make([]byte, h.MsgLen)
			_, err = io.ReadFull(fb, msg)
			if err != nil {
				err = errors.Info(err, "mq.loopProcessingMsg: read message failed").Detail(err)
				return
			}
			_, err = r.Put(msg)
			if err != nil {
				err = errors.Info(err, "mq.loopProcessingMsg: reput message failed").Detail(err)
				return
			}
			chgState(r.mqf, offTimeout, stateDone)
			atomic.AddInt64(&r.doing, -1)
			offTimeout += (headerSize + int64(h.MsgLen))
			continue
		}
		if h.State == stateDone {
			offTimeout += (headerSize + int64(h.MsgLen))
			continue
		}
		if h.State != stateNormal {
			err = errors.Info(ErrInvalidRecord, "mq.nextProcessingMsg: invalid record state")
			return
		}
	}
	return
}

func travel(mqf *largefile.Instance, off int64, allowStates uint8) (offNew int64, err error) {

	log.Debug("travel:", off, int(allowStates))

	const BufSize = 1 << 16

	f := &cc.Reader{ReaderAt: mqf, Offset: off}
	bufr := bufio.NewReaderSize(f, BufSize)

	for {
		var h recordHeader
		err = binary.Read(bufr, binary.LittleEndian, &h)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if (h.State & allowStates) == 0 {
			break
		}
		_, err = bufr.Next(int(h.MsgLen))
		if err != nil {
			break
		}
		off += (headerSize + int64(h.MsgLen))
	}
	return off, err
}

func count(mqf *largefile.Instance, start, end int64, allowStates uint8) (cnt int64, err error) {

	log.Debug("count:", start, end)

	const BufSize = 1 << 16

	f := &cc.Reader{ReaderAt: mqf, Offset: start}
	bufr := bufio.NewReaderSize(f, BufSize)

	for {
		if start >= end {
			break
		}
		var h recordHeader
		err = binary.Read(bufr, binary.LittleEndian, &h)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if h.State == allowStates {
			cnt++
		}
		_, err = bufr.Next(int(h.MsgLen))
		if err != nil {
			break
		}

		start += (headerSize + int64(h.MsgLen))
	}
	return
}

func OpenInstance(name string, chunkBits uint, expires uint32) (r *Instance, err error) {

	mqf, err := largefile.Open(name, chunkBits)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: Open failed -", name).Detail(err)
		return
	}

	_, err = os.Stat(name + "/INDEX")
	if err != nil && os.IsNotExist(err) {
		_, err = os.Stat(name + "/index")
		if err == nil {
			log.Warn("rename ", name+"/index to", name+"/INDEX")
			err = os.Rename(name+"/index", name+"/INDEX")
			if err != nil {
				err = errors.Info(err, "mq index file rename failed -", name).Detail(err)
				return
			}
		}
	}
	findex, err := os.OpenFile(name+"/INDEX", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: ReadIndexFile failed -", name+".idx").Detail(err)
		return
	}
	defer func() {
		if err != nil {
			findex.Close()
		}
	}()

	var index indexFile
	err = binary.Read(findex, binary.LittleEndian, &index)
	if err != nil {
		if err != io.EOF {
			err = errors.Info(err, "mq.NewInstance: ReadIndexFile failed").Detail(err)
			return
		} else {
			index.Expires = expires
		}
	}
	log.Info("Index:", index.OffPut, index.OffGet, index.OffTimeout)

	index.OffPut, err = travel(mqf, index.OffPut, stateTravelPut)
	if err != nil {
		errors.Info(err, "mq.NewInstance: travelPut failed").Detail(err).Warn()
		//TODO: truncate mqf to index.OffPut
	}

	index.OffTimeout, err = travel(mqf, index.OffTimeout, stateTravelTimeout)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: travelTimeout failed").Detail(err)
		return
	}

	index.OffGet, err = travel(mqf, index.OffGet, stateTravelGet)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: travelGet failed").Detail(err)
		return
	}

	cntDoing, err := count(mqf, index.OffTimeout, index.OffGet, stateProcessing)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: count doing failed").Detail(err)
		return
	}
	cntTodo, err := count(mqf, index.OffGet, index.OffPut, stateNormal)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: count todo failed").Detail(err)
		return
	}

	r = &Instance{
		fp:         cc.Writer{WriterAt: mqf, Offset: index.OffPut},
		fg:         cc.Reader{ReaderAt: mqf, Offset: index.OffGet},
		mqf:        mqf,
		index:      findex,
		offTimeout: index.OffTimeout,
		done:       make(chan bool, 1),
		expires:    index.Expires,
		loop:       true,
		doing:      cntDoing,
		todo:       cntTodo,
	}
	go r.loopGetTimeout()
	return
}

func (r *Instance) Close() (err error) {

	r.loop = false
	<-r.done

	r.mutexi.Lock()
	r.index.Close()
	r.mutexi.Unlock()

	r.mqf.Close()

	return nil
}

func (r *Instance) PutString(msg string) (msgId int64, err error) {

	return r.Put([]byte(msg))
}

func (r *Instance) Put(msg []byte) (msgId int64, err error) {

	if len(msg) >= 0x10000 {
		return 0, errors.EINVAL
	}

	r.mutexp.Lock()
	defer r.mutexp.Unlock()

	b := make([]byte, headerSize+len(msg))
	bw := bytes.NewWriter(b)

	binary.Write(bw, binary.LittleEndian, &recordHeader{
		Tag:    recordTag,
		State:  stateNormal,
		MsgLen: uint16(len(msg)),
	})
	bw.Write(msg)

	msgId = r.fp.Offset
	_, err = r.fp.Write(b)
	if err != nil {
		err = errors.Info(err, "mq.Put failed").Detail(err).Warn()
		r.fp.Offset = msgId
	}
	if err == nil {
		atomic.AddInt64(&r.todo, 1)
	}
	return
}

func (r *Instance) getOffPut() int64 {

	r.mutexp.Lock()
	defer r.mutexp.Unlock()

	return r.fp.Offset
}

func (r *Instance) GetString() (msgId int64, msg string, err error) {

	msgId, msgVal, err := r.Get()
	if err == nil {
		msg = string(msgVal)
	}
	return
}

func (r *Instance) Get() (msgId int64, msg []byte, err error) {

	r.mutexg.Lock()
	defer r.mutexg.Unlock()

retry:
	msgId, msg, state, err := getMsg(&r.fg)
	if err != nil {
		return
	}
	if state == stateDone {
		goto retry
	}

	err = lockMsg(r.mqf, msgId, atomic.LoadUint32(&r.expires))
	if err == nil {
		atomic.AddInt64(&r.todo, -1)
		atomic.AddInt64(&r.doing, 1)
	}
	return
}

func (r *Instance) FilterMsgs(filter func(msg []byte) bool) error {

	r.mutexg.Lock()
	defer r.mutexg.Unlock()

	fg := r.fg
	for {
		msgId, msg, state, err := getMsg(&fg)
		if err != nil {
			if err == ENoSuchEntry {
				return nil
			}
			return err
		}
		if state == stateDone {
			continue
		}
		if filter(msg) {
			chgState(r.mqf, msgId, stateDone)
		}
	}
}

func getMsg(fg *cc.Reader) (msgId int64, msg []byte, state uint8, err error) {

	msgId = fg.Offset

	var h recordHeader
	err = binary.Read(fg, binary.LittleEndian, &h)
	if err != nil {
		//
		// TODO: if err != io.EOF: how to fix this error? (通知人工介入)
		if err == io.EOF {
			err = ENoSuchEntry
			return
		}
		err = errors.Info(err, "mq.getMsg failed", msgId, h).Detail(err).Warn()
		fg.Offset = msgId
		return
	}

	if h.Tag != recordTag {
		err = errors.Info(ErrInvalidRecord, "mq.getMsg failed", msgId, h).Warn()
		fg.Offset = msgId
		return
	}

	msg = make([]byte, h.MsgLen)
	_, err = io.ReadFull(fg, msg)
	if err != nil {
		err = errors.Info(err, "mq.getMsg failed", msgId, h).Detail(err).Warn()
		fg.Offset = msgId
		return
	}

	state = h.State
	return
}

func (r *Instance) getOffGet() int64 {

	r.mutexg.Lock()
	defer r.mutexg.Unlock()

	return r.fg.Offset
}

func (r *Instance) Delete(msgId int64) (err error) {

	err = chgState(r.mqf, msgId, stateDone)
	if err == nil {
		atomic.AddInt64(&r.doing, -1)
	}
	return
}

func (r *Instance) updateIndex(offPut, offGet, offTimeout int64) (err error) {
	r.mutexi.Lock()
	defer r.mutexi.Unlock()

	expires := atomic.LoadUint32(&r.expires)

	b := make([]byte, indexSize)
	err = binary.Write(bytes.NewWriter(b), binary.LittleEndian, &indexFile{offPut, offGet, offTimeout, expires, 0})
	if err != nil {
		return
	}

	_, err = r.index.WriteAt(b, 0)
	return
}

func (r *Instance) UpdateIndexExpires(expires uint32) (err error) {
	r.mutexi.Lock()
	defer r.mutexi.Unlock()

	fi, err := r.index.Stat()
	if err != nil {
		return
	}

	var index indexFile
	b := make([]byte, indexSize)

	// 1. read index from disk if is not empty
	if fi.Size() != 0 {
		n, err1 := r.index.ReadAt(b, 0)
		if n < indexSize || (err1 != nil && err1 != io.EOF) {
			return err1
		}

		err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &index)
		if err != nil {
			return
		}
	}

	// update
	index.Expires = expires

	err = binary.Write(bytes.NewWriter(b), binary.LittleEndian, &index)
	if err != nil {
		return
	}

	// 2. save back to disk
	_, err = r.index.WriteAt(b, 0)
	if err != nil {
		return
	}

	// 3. update expires in mem
	atomic.StoreUint32(&r.expires, expires)
	return
}

type StatInfo struct {
	TodoLen       int64 `json:"todo"`
	ProcessingLen int64 `json:"doing"`
}

func (r *Instance) Stat() StatInfo {

	doing := atomic.LoadInt64(&r.doing)
	todo := atomic.LoadInt64(&r.todo)

	return StatInfo{TodoLen: todo, ProcessingLen: doing}
}

// ---------------------------------------------------------------
