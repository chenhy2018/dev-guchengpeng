package mmq

import (
	"sync"
	"time"
)

// ------------------------------------------------------------------

const (
	stateProcessing = 0x02
	stateDone       = 0x04
)

type message struct {
	msg      []byte
	id       string
	lockTime int64

	tryTimes uint32
	state    uint32
	mutex    sync.Mutex
}

func (p *message) waitAndGetState() (state uint32) {

	p.mutex.Lock()
	lockTime, state := p.lockTime, p.state
	p.mutex.Unlock()

	if state == stateDone {
		return
	}

	currTime := time.Now().UnixNano()
	if currTime < lockTime {
		time.Sleep(time.Duration(lockTime - currTime))
	}

	p.mutex.Lock()
	state = p.state
	p.mutex.Unlock()
	return
}

func (p *message) setProcessing(expires uint32) bool {

	p.mutex.Lock()
	if p.state == stateDone {
		p.mutex.Unlock()
		return false
	}
	p.state = stateProcessing
	p.lockTime = time.Now().UnixNano() + int64(expires)*1e9
	p.mutex.Unlock()
	return true
}

func (p *message) done() {

	p.mutex.Lock()
	p.state = stateDone
	p.mutex.Unlock()
}

// ------------------------------------------------------------------
