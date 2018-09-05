package dlog

import (
	"bytes"
	"hash/crc32"
	"io"
	"os"
	"qbox.us/bufio"
	"qbox.us/encoding/binary"
	"qbox.us/errors"
	"github.com/qiniu/log.v1"
	"sync"
)

/*
日志文件每条记录的格式如下：

	Crc32 uint32

	Cmd uint16
	Len uint16
	Time uint32		// 精确到秒的时间
	Data [Len]byte
*/

const (
	headerBytes = 12
	cmdOffset   = 4
	lenOffset   = 6
)

var ErrInvalidLogItem = errors.Register("invalid logitem")

// --------------------------------------------------------------------

type Writer struct {
	file  *os.File
	off   int64
	mutex sync.Mutex
}

func OpenWriter(path string) (dlog *Writer, err error) {

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		err = errors.Info(err, "dlog.OpenWriter", path).Detail(err)
		return
	}

	off, err := f.Seek(0, 2)
	if err != nil {
		err = errors.Info(err, "dlog.OpenWriter").Detail(err)
		return
	}
	log.Debug("dlog.OpenWriter:", path, off)

	dlog = &Writer{file: f, off: off}
	return
}

func (r *Writer) Close() (err error) {

	oldfile := r.file
	r.file = nil
	r.off = 0
	if oldfile != nil {
		oldfile.Close()
	}
	return nil
}

func (r *Writer) FilePath() (path string) {

	return r.file.Name()
}

func (r *Writer) Reader() (f io.ReaderAt) {

	return r.file
}

func (r *Writer) Len() (length int64) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.off
}

func (r *Writer) syncPut(rec []byte) (err error) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	off := r.off
	r.off += int64(len(rec))
	_, err = r.file.WriteAt(rec, off)
	return
}

/*
func (r *Writer) alloc(size int) (off int64) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	off = r.off
	r.off += int64(size)
	return
}

func (r *Writer) asyncPut(rec []byte) (err error) {

	off := r.alloc(len(rec))
	_, err = r.file.WriteAt(rec, off)
	return
}
*/

func (r *Writer) Position() (off int64) {

	r.mutex.Lock()
	off = r.off
	r.mutex.Unlock()

	return
}

func (r *Writer) Put(v interface{}) (n int, err error) {

	log.Debug("dlog.Put:", v)

	var crc [4]byte

	b := bytes.NewBuffer(nil)
	b.Write(crc[:])
	err = binary.Write(b, binary.LittleEndian, v)
	if err != nil {
		err = errors.Info(err, "dlog.Writer.Put", v).Detail(err)
		return
	}

	rec := b.Bytes()
	n = len(rec)

	l := binary.LittleEndian.Uint16(rec[lenOffset:])
	if n != int(l)+headerBytes {
		err = errors.Info(ErrInvalidLogItem, "dlog.Writer.Put: invalid length -", l)
		return
	}

	binary.LittleEndian.PutUint32(rec, crc32.ChecksumIEEE(rec[cmdOffset:]))

	err = r.syncPut(rec)
	if err != nil {
		err = errors.Info(err, "dlog.Writer.Put").Detail(err)
	}
	return
}

// --------------------------------------------------------------------

func NewReader(f io.ReadSeeker, from, n int64) (r *bufio.Reader, err error) {

	return NewReaderSize(f, from, n, 4096)
}

func NewReaderSize(f io.ReadSeeker, from, n int64, bufsize int) (r *bufio.Reader, err error) {

	_, err = f.Seek(from, 0)
	if err != nil {
		err = errors.Info(err, "dlog.NewReaderSize").Detail(err)
		return
	}

	lr := io.LimitReader(f, n)
	r = bufio.NewReaderSize(lr, bufsize)
	return
}

func Next(r *bufio.Reader) (cmd uint16, msg []byte, err error) {

	hdr, err := r.Peek(headerBytes)
	if err != nil {
		if err == io.EOF {
			return
		}
		err = errors.Info(err, "dlog.Next", "Read record header failed").Detail(err)
		return
	}

	l := binary.LittleEndian.Uint16(hdr[lenOffset:])
	rec, err := r.Next(int(l) + headerBytes)
	if err != nil {
		err = errors.Info(err, "dlog.Next", "Read record body failed").Detail(err)
		return
	}

	if binary.LittleEndian.Uint32(rec) != crc32.ChecksumIEEE(rec[cmdOffset:]) {
		err = errors.Info(errors.ErrUnmatchedChecksum, "dlog.Next")
		return
	}
	return binary.LittleEndian.Uint16(rec[cmdOffset:]), rec[cmdOffset:], nil
}

// --------------------------------------------------------------------
