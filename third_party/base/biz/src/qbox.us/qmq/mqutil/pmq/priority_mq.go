package pmq

import (
	"qbox.us/qmq/mqutil"
)

// ------------------------------------------------------------------

type MessageQ struct {
}

func NewMessageQ() *MessageQ {

	p := new(MessageQ)
	return p
}

func (p *MessageQ) Put(uid uint32, m *mqutil.Message) {

}

func (p *MessageQ) TryGet() (uid uint32, m *mqutil.Message, ok bool) {

	return
}

// ------------------------------------------------------------------
