package ts

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func doTestRandReader(t *testing.T) {

	fmt.Println("Begin doTestRandReader")
	rbs, err := NewBigRandBytes(512 * 1024)
	if err != nil {
		t.Fatal(err)
	}

	{
		r := NewRandReader(rbs, 20)
		buf := new(bytes.Buffer)
		_, err = io.CopyN(buf, r, 10)
		if err != nil {
			t.Fatal(err)
		}
		if buf.Len() != 10 {
			t.Fatal("RandReader read failed", err)
		}
		fmt.Println("buf:", buf.Bytes())
	}

	{
		r := NewRandReader(rbs, 20)
		buf := new(bytes.Buffer)
		_, err = io.CopyN(buf, r, 40)
		if err != io.EOF {
			t.Fatal("Should EOF")
		}
	}

	{
		r := NewRandReader(rbs, 200)
		buf := new(bytes.Buffer)
		_, err = io.CopyN(buf, r, 40)
		if err != nil {
			t.Fatal(err)
		}
		if buf.Len() != 40 {
			t.Fatal("RandReader read failed", buf.Len())
		}
		fmt.Println("buf:", buf.Bytes())
		buf = new(bytes.Buffer)
		_, err = io.Copy(buf, r)
		if err != nil {
			t.Fatal(err)
		}
		if buf.Len() != 160 {
			t.Fatal("RandReader read failed", buf.Len())
		}
	}
}

func doTestWReader(t *testing.T) {

	fmt.Println("Begin doTestWReader")
	rbs, err := NewBigRandBytes(512 * 1024)
	if err != nil {
		t.Fatal(err)
	}

	r := NewRandReader(rbs, 20)
	wr := NewWReader(r, new(bytes.Buffer))
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, wr)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 20 {
		t.Fatal("RandReader read failed", err)
	}
	wrBuf := wr.w.(*bytes.Buffer)
	if fmt.Sprint(buf.Bytes()) != fmt.Sprint(wrBuf.Bytes()) {
		t.Fatal("WReader write failed", buf.Bytes(), wrBuf.Bytes())
	}
}

func Test(t *testing.T) {
	doTestRandReader(t)
	doTestWReader(t)
}
