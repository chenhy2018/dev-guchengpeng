package mqutil

// ------------------------------------------------------------------

type MessageQ struct {
	head *Message
	tail **Message
}

func NewMessageQ() *MessageQ {

	p := new(MessageQ)
	p.tail = &p.head
	return p
}

func (p *MessageQ) Put(m *Message) {

	*p.tail = m
	p.tail = &m.prev
}

func (p *MessageQ) TryGet() (m *Message, ok bool) {

	if p.head != nil {
		m, ok = p.head, true
		p.head = m.prev
		if p.head == nil {
			p.tail = &p.head
		}
		m.prev = nil
	}
	return
}

// ------------------------------------------------------------------
