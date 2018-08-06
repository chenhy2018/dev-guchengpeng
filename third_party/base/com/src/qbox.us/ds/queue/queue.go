package queue

// --------------------------------------------------------------------

type elem struct {
	prev *elem
	val  interface{}
}

type Queue struct {
	top  *elem
	tail **elem
}

func (r *Queue) Push(val interface{}) {
	e := &elem{nil, val}
	if r.top != nil {
		*r.tail = e
	} else {
		r.top = e
	}
	r.tail = &e.prev
}

func (r *Queue) Pop() (val interface{}, ok bool) {
	e := r.top
	if e != nil {
		r.top = e.prev
		return e.val, true
	}
	return
}

func (r *Queue) Top() (val interface{}) {
	if r.top != nil {
		return r.top.val
	}
	return
}

func (r *Queue) Empty() (yes bool) {
	return r.top == nil
}

// --------------------------------------------------------------------
