package mmq

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
)

var msgBaseId int64

func init() {
	log.SetOutputLevel(0)
}

func (r *Instance) PutString(msg string) (msgId string, err error) {
	msgBaseId++
	msgId = strconv.FormatInt(msgBaseId, 36)
	r.Put(msgId, []byte(msg))
	return
}

func (r *Instance) GetString(expires int64, cancel <-chan bool) (msgId string, msg string, err error) {
	msgId, msgb, err := r.Get(expires, cancel)
	return msgId, string(msgb), err
}

func (r *Instance) GetStringInf(cancel <-chan bool) (msgId string, msg string, err error) {
	msgId, msgb, err := r.GetInf(cancel)
	return msgId, string(msgb), err
}

func (r *Instance) GetStringAtOnce() (msgId string, msg string, err error) {
	msgId, msgb, err := r.GetAtOnce()
	return msgId, string(msgb), err
}

func TestMQ(t *testing.T) {

	r := NewInstance(100, 1, 4)

	msgId1, err := r.PutString("Hello, world!")
	if err != nil {
		ts.Fatal(t, "PutString failed:", err)
	}

	msgId2, err := r.PutString("Hello, golang!!!")
	if err != nil {
		ts.Fatal(t, "PutString2 failed:", err)
	}

	msgId11, s1, err := r.GetStringInf(nil)
	if err != nil || msgId1 != msgId11 || s1 != "Hello, world!" {
		ts.Fatal(t, "GetString failed:", msgId1, msgId11, s1)
	}

	stat := r.Stat()
	fmt.Println("Stat:", stat)
	if !(stat.TodoLen == 1 && stat.ProcessingLen == 1) {
		ts.Fatal(t, "Stat failed:", stat)
	}

	msgId12, s2, err := r.GetStringAtOnce()
	if err != nil {
		ts.Fatal(t, "GetString2 failed:", err)
	}
	if msgId2 != msgId12 || s2 != "Hello, golang!!!" {
		ts.Fatal(t, "GetString2 failed:", msgId2, msgId12, s2)
	}

	stat = r.Stat()
	fmt.Println("Stat:", stat)
	if !(stat.TodoLen == 0 && stat.ProcessingLen == 2) {
		ts.Fatal(t, "Stat failed:", stat)
	}

	cn := make(chan bool, 1)
	cn <- true

	_, _, err = r.GetStringInf(cn)
	if err != ErrCancelledOp {
		ts.Fatal(t, "GetStringInf failed:", err)
	}

	_, _, err = r.GetString(0, nil)
	if err != ErrNoSuchEntry {
		ts.Fatal(t, "GetString:", err)
	}

	msgId21, s21, err := r.GetString(15e8, nil)
	if err != nil {
		ts.Fatal(t, "GetString failed:", err)
	}
	if s21 != "Hello, world!" {
		ts.Fatal(t, "GetString21 failed:", s21)
	}
	fmt.Println("GetString21:", msgId21, s21)

	_, s22, err := r.GetStringInf(nil)
	if err != nil || s22 != "Hello, golang!!!" {
		ts.Fatal(t, "GetString22 failed:", s22)
	}

	_, _, err = r.GetStringAtOnce()
	if err != ErrNoSuchEntry {
		ts.Fatal(t, "GetString:", err)
	}

	err = r.Delete(msgId21, false)
	if err != nil {
		ts.Fatal(t, "Delete:", err)
	}

	stat = r.Stat()
	fmt.Println("Stat:", stat)
	if !(stat.TodoLen == 0 && stat.ProcessingLen == 1) {
		ts.Fatal(t, "Stat failed:", stat)
	}

	msgId32, s32, err := r.GetString(11e8, nil)
	if err != nil {
		ts.Fatal(t, "GetString failed:", err)
	}
	if s32 != "Hello, golang!!!" {
		ts.Fatal(t, "GetString32 failed:", msgId32, s32)
	}

	err = r.Delete(msgId32, false)
	if err != nil {
		ts.Fatal(t, "Delete:", err)
	}

	_, _, err = r.GetStringAtOnce()
	if err != ErrNoSuchEntry {
		ts.Fatal(t, "GetString:", err)
	}

	stat = r.Stat()
	fmt.Println("Stat:", stat)
	if !(stat.TodoLen == 0 && stat.ProcessingLen == 0) {
		ts.Fatal(t, "Stat failed:", stat)
	}

	r.Close()

	if len(r.allMsgs) != 0 {
		ts.Fatal(t, "len(r.allMsgs) != 0")
	}
}
