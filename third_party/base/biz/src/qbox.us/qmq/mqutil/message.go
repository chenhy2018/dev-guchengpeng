package mqutil

import (
	"sync"
	"time"
)

// ------------------------------------------------------------------

const (
	stateInvalid    = 0
	stateProcessing = 0x02

	StateDone = 0x04
)

type Message struct {
	prev *Message

	Id       string
	Msg      []byte
	TryTimes uint32
	State    uint32

	lockTime int64
	mutex    sync.Mutex
}

func (p *Message) WaitAndGetState() (state uint32) {

	p.mutex.Lock()
	lockTime, state := p.lockTime, p.State
	p.mutex.Unlock()

	if state == StateDone {
		return
	}

	currTime := time.Now().UnixNano()
	if currTime < lockTime {
		time.Sleep(time.Duration(lockTime - currTime))
	}

	p.mutex.Lock()
	state = p.State
	p.mutex.Unlock()
	return
}

func (p *Message) SetProcessing(expires uint32) bool {

	p.mutex.Lock()
	if p.State == StateDone {
		p.mutex.Unlock()
		return false
	}
	p.State = stateProcessing
	p.lockTime = time.Now().UnixNano() + int64(expires)*1e9
	p.mutex.Unlock()
	return true
}

func (p *Message) Done() {

	p.mutex.Lock()
	p.State = StateDone
	p.mutex.Unlock()
}

// ------------------------------------------------------------------
