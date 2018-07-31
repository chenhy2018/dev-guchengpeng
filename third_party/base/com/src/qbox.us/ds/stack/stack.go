package stack

// --------------------------------------------------------------------

type elem struct {
	prev *elem
	val  interface{}
}

type Stack struct {
	top *elem
}

func (r *Stack) Push(val interface{}) {
	r.top = &elem{r.top, val}
}

func (r *Stack) Pop() (val interface{}, ok bool) {
	if top := r.top; top != nil {
		r.top = top.prev
		return top.val, true
	}
	return
}

func (r *Stack) Top() (val interface{}) {
	if r.top != nil {
		return r.top.val
	}
	return nil
}

func (r *Stack) Empty() (yes bool) {
	return r.top == nil
}

// --------------------------------------------------------------------
