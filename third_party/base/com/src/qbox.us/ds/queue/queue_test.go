package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {

	var que Queue

	que.Push(1)
	que.Push(2)

	v0 := que.Top()
	if v0 == nil || v0.(int) != 1 {
		t.Fatal("que.Top() != 1")
	}

	v1, ok := que.Pop()
	if !ok || v1.(int) != 1 {
		t.Fatal("que.Pop failed")
	}

	if que.Empty() {
		t.Fatal("que.Empty failed")
	}

	v2, ok := que.Pop()
	if !ok || v2.(int) != 2 {
		t.Fatal("que.Pop failed")
	}

	_, ok = que.Pop()
	if ok {
		t.Fatal("que.Pop empty stack failed")
	}

	v := que.Top()
	if v != nil {
		t.Fatal("que.Top() != nil")
	}

	if !que.Empty() {
		t.Fatal("que.Empty failed")
	}

	que.Push(3)
	que.Push(4)

	v3 := que.Top()
	if v3 == nil || v3.(int) != 3 {
		t.Fatal("que.Top() != 3")
	}

	v31, ok := que.Pop()
	if !ok || v31.(int) != 3 {
		t.Fatal("que.Pop failed")
	}

	v4, ok := que.Pop()
	if !ok || v4.(int) != 4 {
		t.Fatal("que.Pop failed")
	}
}
