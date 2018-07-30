package dlog

import (
	"fmt"
	"io"
	"os"
	"qbox.us/cc"
	"qbox.us/encoding/binary"
	"qbox.us/errors"
	"github.com/qiniu/ts"
	"testing"
)

type fooRecord struct {
	Cmd  uint16
	Len  uint16
	Time uint32 // 精确到秒的时间
	Data []byte
}

const (
	fooRec = 234
)

func TestDlog(t *testing.T) {

	home := os.Getenv("HOME")
	dlogFile := home + "/dlogTest.log"

	f, err := OpenWriter(dlogFile)
	if err != nil {
		ts.Fatal(t, "OpenWriter failed:", err)
	}
	defer f.Close()

	var rec fooRecord
	data := "foo Record"
	rec.Cmd = fooRec
	rec.Len = uint16(len(data))
	rec.Data = []byte(data)
	{
		n, err := f.Put(&rec)
		if err != nil || n != 22 {
			ts.Fatal(t, "Put failed:", n, errors.Detail(err))
		}
		fmt.Println("n:", n)
	}
	{
		n, err := f.Put(&rec)
		if err != nil || n != 22 {
			ts.Fatal(t, "Put failed:", n, errors.Detail(err))
		}
		fmt.Println("n:", n)
	}
	f2, err := os.Open(dlogFile)
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}
	defer f2.Close()

	r, err := NewReader(f2, 0, int64(rec.Len+12)*2)
	if err != nil {
		ts.Fatal(t, "NewReader failed:", err)
	}
	{
		cmd, msg, err := Next(r)
		if err != nil || cmd != fooRec {
			ts.Fatal(t, "Read failed:", cmd, err)
		}

		var rec2 fooRecord
		rec2.Data = make([]byte, len(msg)-8)
		err = binary.Read(cc.NewBytesReader(msg), binary.LittleEndian, &rec2)
		if err != nil || string(rec2.Data) != data || rec.Cmd != fooRec {
			ts.Fatal(t, "binary.Read failed:", err)
		}
		fmt.Println(rec2)
	}
	{
		cmd, msg, err := Next(r)
		if err != nil || cmd != fooRec {
			ts.Fatal(t, "Read failed:", cmd, err)
		}

		var rec2 fooRecord
		rec2.Data = make([]byte, len(msg)-8)
		err = binary.Read(cc.NewBytesReader(msg), binary.LittleEndian, &rec2)
		if err != nil || string(rec2.Data) != data || rec.Cmd != fooRec {
			ts.Fatal(t, "binary.Read failed:", err)
		}
		fmt.Println(rec2)
	}
	{
		_, _, err := Next(r)
		if err == nil || err != io.EOF {
			ts.Fatal(t, "not eof?", err)
		}
	}
}
