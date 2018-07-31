package processmgr

import (
	"os/exec"
	"testing"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/xlog.v1"
)

func TestProcessMgr(t *testing.T) {
	if testNormal(t) {
		t.Log("success")
	} else {
		t.Error("failed")
	}
	if testWithInterrupt(t) {
		t.Log("success")
	} else {
		t.Error("failed")
	}
}

func testNormal(t *testing.T) bool {
	cmd := exec.Command("sleep", "3")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	} else {
		t.Log("cmd started")
	}
	var xl = xlog.NewDummy()
	ctx := context.Background()
	Add(xl, cmd.Process, ctx)
	t.Log("process added to processmgr")
	defer func() {
		Del(xl)
	}()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if result := cmd.ProcessState.Success(); result == true {
		t.Log("process fully executed")
		return true
	} else {
		return false
	}
}

func testWithInterrupt(t *testing.T) bool {
	cmd := exec.Command("sleep", "3")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	} else {
		t.Log("cmd started")
	}
	var xl = xlog.NewDummy()
	ctx, cancel := context.WithCancel(context.Background())
	Add(xl, cmd.Process, ctx)
	t.Log("process added to processmgr")
	defer func() {
		Del(xl)
	}()
	cancel()
	if err := cmd.Wait(); err != nil {
		t.Log("the program is not fully executed as expected")
	}
	if result := cmd.ProcessState.Success(); result == false {
		return true
	} else {
		return false
	}

}
