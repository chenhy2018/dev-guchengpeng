package mmq

import (
	"sync"
	"syscall"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
)

var (
	ErrMqFull      = httputil.NewError(801, "message queue full")
	ErrCancelledOp = httputil.NewError(499, "cancelled op")
	ErrNoSuchEntry = syscall.ENOENT
)

// ------------------------------------------------------------------

type StatInfo struct {
	TodoLen       int `json:"todo"`
	ProcessingLen int `json:"doing"`
}

type Instance struct {
	mq     chan *message
	lockMq chan *message // TODO: lockMq implementation
	done   chan bool

	allMsgs map[string]*message
	mutex   sync.Mutex

	expires  uint32
	tryTimes uint32
}

func NewInstance(mqcap int, expires, tryTimes uint32) *Instance {

	mq := make(chan *message, mqcap)
	lockMq := make(chan *message, mqcap)
	done := make(chan bool, 1)
	allMsgs := make(map[string]*message)
	r := &Instance{
		expires:  expires,
		tryTimes: tryTimes,
		mq:       mq,
		lockMq:   lockMq,
		done:     done,
		allMsgs:  allMsgs,
	}
	go r.loopGetTimeout()
	return r
}

func (r *Instance) Close() (err error) {

	close(r.lockMq)

	<-r.done
	return nil
}

func (r *Instance) Len() int {

	r.mutex.Lock()
	allLen := len(r.allMsgs)
	r.mutex.Unlock()

	return allLen
}

func (r *Instance) Stat() StatInfo {

	todoLen := len(r.mq)

	r.mutex.Lock()
	allLen := len(r.allMsgs)
	r.mutex.Unlock()

	processingLen := allLen - todoLen
	if processingLen < 0 {
		processingLen = 0
	}

	return StatInfo{TodoLen: todoLen, ProcessingLen: processingLen}
}

func (r *Instance) loopGetTimeout() {

	for {
		m, ok := <-r.lockMq
		if !ok {
			break
		}

		state := m.waitAndGetState()
		if state == stateDone {
			continue
		}

		m.tryTimes++
		if m.tryTimes >= r.tryTimes {
			r.Delete(m.id, true)
			continue
		}

		m.state = 0
		r.mq <- m
	}
	r.done <- true
}

func (r *Instance) TryPut(msgId string, msg []byte) (err error) {

	m := &message{id: msgId, msg: msg}

	select {
	case r.mq <- m:
	default:
		return ErrMqFull
	}

	r.mutex.Lock()
	r.allMsgs[msgId] = m
	r.mutex.Unlock()

	return nil
}

func (r *Instance) Put(msgId string, msg []byte) {

	m := &message{id: msgId, msg: msg}

	r.mq <- m

	r.mutex.Lock()
	r.allMsgs[msgId] = m
	r.mutex.Unlock()
}

var defaultCancelChan = make(chan bool)

func (r *Instance) Get(expires int64, cancel <-chan bool) (msgId string, msg []byte, err error) {

	now := time.Now().UnixNano()
	deadline := now + expires
	if cancel == nil {
		cancel = defaultCancelChan
	}
	for {
		delta := deadline - now
		if delta < 0 {
			break
		}
		select {
		case m := <-r.mq:
			if m.setProcessing(r.expires) {
				r.lockMq <- m
				return m.id, m.msg, nil
			}
			now = time.Now().UnixNano()
		case <-time.After(time.Duration(delta)):
			return "", nil, ErrNoSuchEntry
		case <-cancel:
			return "", nil, ErrCancelledOp
		}
	}
	return "", nil, ErrNoSuchEntry
}

func (r *Instance) GetInf(cancel <-chan bool) (msgId string, msg []byte, err error) {

	if cancel != nil {
		for {
			select {
			case m := <-r.mq:
				if m.setProcessing(r.expires) {
					r.lockMq <- m
					return m.id, m.msg, nil
				}
			case <-cancel:
				return "", nil, ErrCancelledOp
			}
		}
	} else {
		for {
			m := <-r.mq
			if m.setProcessing(r.expires) {
				r.lockMq <- m
				return m.id, m.msg, nil
			}
		}
	}
}

func (r *Instance) GetAtOnce() (msgId string, msg []byte, err error) {

	select {
	case m := <-r.mq:
		if m.setProcessing(r.expires) {
			r.lockMq <- m
			return m.id, m.msg, nil
		}
	default:
		break
	}
	return "", nil, ErrNoSuchEntry
}

func (r *Instance) Delete(msgId string, cancel bool) (err error) {

	r.mutex.Lock()
	m, ok := r.allMsgs[msgId]
	if !ok {
		r.mutex.Unlock()
		return ErrNoSuchEntry
	}

	delete(r.allMsgs, msgId)
	r.mutex.Unlock()

	m.done()
	if cancel {
		log.Warn("Message skipped:", msgId)
	}
	return
}

// ------------------------------------------------------------------
