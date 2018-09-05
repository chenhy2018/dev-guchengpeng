package state

import "errors"
import "testing"

func testFunc1() {
	unit := state.Enter("qbox.us/state/testFunc1")
	defer unit.Leave(nil)
}

func testFunc2() (err error) {
	unit := state.Enter("qbox.us/state/testFunc2")
	defer unit.Leave(&err)

	err = errors.New("haha")
	return
}

func TestState(t *testing.T) {

	testFunc1()
	if info, ok := state.Dump()["qbox.us/state/testFunc1"]; !ok || info.Count != 1 {
		t.Fatal("failed")
	}

	testFunc2()
	if info, ok := state.Dump()["qbox.us/state/testFunc2"]; !ok || info.FailCount != 1 {
		t.Fatal("failed")
	}
}
