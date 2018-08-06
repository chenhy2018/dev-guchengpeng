package mockmq

import (
	"bytes"
	. "qbox.us/api/conf"
	mqapi "qbox.us/qmq/qmqapi/v1/mq"
	"testing"
	"time"
)

func TestMockMQ(t *testing.T) {

	svr := New()
	go svr.Run(":12345")
	time.Sleep(1e9)

	doTestMockMQ(t)
}

func doTestMockMQ(t *testing.T) {

	buf1 := []byte("hello.world")
	buf2 := []byte("hello.china")

	MQ_HOST = "http://localhost:12345"
	mq := mqapi.New()

	err := mq.Make(nil, "nfop", 3)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = mq.Get(nil, "nfop")
	if err == nil {
		t.Fatal("should err")
	}

	msgid, err := mq.Put(nil, "nfop", buf1)
	if err != nil {
		t.Fatal(err)
	}

	msgid, err = mq.Put(nil, "nfop", buf2)
	if err != nil {
		t.Fatal(err)
	}

	buf, msgid, err := mq.Get(nil, "nfop")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, buf1) {
		t.Fatal("should equal", string(buf), string(buf1))
	}

	buf, msgid, err = mq.Get(nil, "nfop")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, buf2) {
		t.Fatal("should equal", buf, buf2)
	}

	_, _, err = mq.Get(nil, "nfop")
	if err == nil {
		t.Fatal("should err")
	}

	time.Sleep(3e9)

	buf, msgid, err = mq.Get(nil, "nfop")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, buf1) {
		t.Fatal("should equal", string(buf), string(buf1))
	}

	err = mq.Delete(nil, "nfop", msgid)
	if err != nil {
		t.Fatal(err)
	}

	buf, msgid, err = mq.Get(nil, "nfop")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, buf2) {
		t.Fatal("should equal", buf, buf2)
	}
}
