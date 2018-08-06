package mq

import (
	"os"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
	"qbox.us/net/httputil"
	"qbox.us/qmq/qmqapi/v1/mq"
)

func init() {
	log.SetOutputLevel(1)
}

func checkNoSuchEntry(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*httputil.ErrorInfo); ok {
		return e.Code == mq.NoSuchEntry
	}
	return false
}

func TestMQ(t *testing.T) {

	home := os.Getenv("HOME")
	name := home + "/msgQTest"
	os.RemoveAll(name)

	r, err := OpenInstance(name, 4, 1)
	if err != nil {
		ts.Fatal(t, "OpenInstance failed:", err)
	}

	msgId1, err := r.PutString("Hello, world!")
	if err != nil {
		ts.Fatal(t, "PutString failed:", err)
	}

	msgId2, err := r.PutString("Hello, golang!!!")
	if err != nil {
		ts.Fatal(t, "PutString2 failed:", err)
	}

	msgId11, s1, err := r.GetString()
	if err != nil {
		ts.Fatal(t, "GetString failed:", err)
	}
	if msgId1 != msgId11 || s1 != "Hello, world!" {
		ts.Fatal(t, "GetString failed:", msgId1, msgId11, s1)
	}

	stat := r.Stat()
	if !(stat.TodoLen == 1 && stat.ProcessingLen == 1) {
		ts.Fatal(t, "Stat failed:", stat)
	}
	msgId12, s2, err := r.GetString()
	if err != nil {
		ts.Fatal(t, "GetString2 failed:", err)
	}
	if msgId2 != msgId12 || s2 != "Hello, golang!!!" {
		ts.Fatal(t, "GetString2 failed:", msgId2, msgId12, s2)
	}

	_, _, err = r.GetString()
	if !checkNoSuchEntry(err) {
		ts.Fatal(t, "GetString3 failed")
	}

	time.Sleep(2 * time.Second)

	_, s21, err := r.GetString()
	if err != nil {
		ts.Fatal(t, "GetString failed:", err)
	}
	if s21 != "Hello, world!" {
		ts.Fatal(t, "GetString21 failed:", s21)
	}

	_, s22, err := r.GetString()
	if err != nil {
		ts.Fatal(t, "GetString failed:", err)
	}
	if s22 != "Hello, golang!!!" {
		ts.Fatal(t, "GetString22 failed:", s22)
	}

	_, _, err = r.GetString()
	if !checkNoSuchEntry(err) {
		ts.Fatal(t, "GetString23 failed")
	}

	r.Close()

	{
		r, err := OpenInstance(name, 4, 1)
		if err != nil {
			ts.Fatal(t, "OpenInstance failed:", err)
		}
		defer r.Close()

		time.Sleep(2 * time.Second)

		stat := r.Stat()
		if !(stat.TodoLen == 2 && stat.ProcessingLen == 0) {
			ts.Fatal(t, "Stat failed:", stat)
		}

		_, s21, err := r.GetString()
		if err != nil {
			ts.Fatal(t, "GetString failed:", err)
		}
		if s21 != "Hello, world!" {
			ts.Fatal(t, "GetString21 failed:", s21)
		}

		msgId22, s22, err := r.GetString()
		if err != nil {
			ts.Fatal(t, "GetString failed:", err)
		}
		if s22 != "Hello, golang!!!" {
			ts.Fatal(t, "GetString22 failed:", s22)
		}

		_, _, err = r.GetString()
		if !checkNoSuchEntry(err) {
			ts.Fatal(t, "GetString23 failed")
		}

		err = r.Delete(msgId22)
		if err != nil {
			ts.Fatal(t, "Delete failed")
		}

		time.Sleep(2 * time.Second)

		{
			_, s21, err := r.GetString()
			if err != nil {
				ts.Fatal(t, "GetString failed:", err)
			}
			if s21 != "Hello, world!" {
				ts.Fatal(t, "GetString21 failed:", s21)
			}

			_, _, err = r.GetString()
			if !checkNoSuchEntry(err) {
				ts.Fatal(t, "GetString23 failed")
			}
		}
	}
	os.RemoveAll(name)
}
