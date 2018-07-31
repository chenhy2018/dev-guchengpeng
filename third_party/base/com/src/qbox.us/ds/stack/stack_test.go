package stack

import (
	"testing"
)

func TestStack(t *testing.T) {

	var stk Stack

	stk.Push(1)
	stk.Push(2)

	v0 := stk.Top()
	if v0 == nil || v0.(int) != 2 {
		t.Fatal("stk.Top() != 2")
	}

	v1, ok := stk.Pop()
	if !ok || v1.(int) != 2 {
		t.Fatal("stk.Pop failed")
	}

	if stk.Empty() {
		t.Fatal("stk.Empty failed")
	}

	v2, ok := stk.Pop()
	if !ok || v2.(int) != 1 {
		t.Fatal("stk.Pop failed")
	}

	_, ok = stk.Pop()
	if ok {
		t.Fatal("stk.Pop empty stack failed")
	}

	v := stk.Top()
	if v != nil {
		t.Fatal("stk.Top() != nil")
	}

	if !stk.Empty() {
		t.Fatal("stk.Empty failed")
	}
}
