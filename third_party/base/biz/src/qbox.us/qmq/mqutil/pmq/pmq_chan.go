package pmq

import (
	"qbox.us/qmq/mqutil"
	"sync"
)

// ------------------------------------------------------------------

type MessageChan struct {
	mq     *MessageQ
	cond   sync.Cond
	mutex  sync.Mutex
	closed bool
}

func NewMessageChan() *MessageChan {

	p := &MessageChan{
		mq: NewMessageQ(),
	}
	p.cond.L = &p.mutex
	return p
}

func (p *MessageChan) Close() (err error) {

	p.mutex.Lock()
	p.closed = true
	p.mutex.Unlock()

	p.cond.Broadcast()
	return nil
}

func (p *MessageChan) Put(uid uint32, m *mqutil.Message) {

	p.mutex.Lock()
	p.mq.Put(uid, m)
	p.mutex.Unlock()

	p.cond.Signal()
}

func (p *MessageChan) Get() (uid uint32, m *mqutil.Message, ok bool) {

	p.mutex.Lock()
	for {
		if uid, m, ok = p.mq.TryGet(); ok {
			break
		}
		if p.closed {
			break
		}
		p.cond.Wait()
	}
	p.mutex.Unlock()

	return
}

func (p *MessageChan) TryGet() (uid uint32, m *mqutil.Message, ok bool) {

	p.mutex.Lock()
	uid, m, ok = p.mq.TryGet()
	p.mutex.Unlock()

	return
}

// ------------------------------------------------------------------
