package stream

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

func TestStream(t *testing.T) {

	gfh := "1"
	gdata := "hello world"
	g := &getter{
		datas: map[string]string{
			gfh: gdata,
		},
	}

	xl := xlog.NewDummy()
	st := New(xl, g, []byte(gfh), int64(len(gdata)))
	if fh, _ := st.QueryFhandle(); !bytes.Equal(fh, []byte(gfh)) {
		t.Fatal("TestStream: unexpected fh", fh)
	}
	if size, _ := st.UploadedSize(); size != int64(len(gdata)) {
		t.Fatal("TestStream: unexpected size", size)
	}

	w := bytes.NewBuffer(nil)
	n, err := io.Copy(w, st)
	if err != nil || string(w.Bytes()) != gdata {
		t.Fatal("TestStream: unexpected io.Copy", err, n)
	}
	st.Seek(-int64(len(gdata)), 2)

	b := make([]byte, 5)
	n1, err := st.Read(b)
	if err != nil || string(b[:n1]) != gdata[0:5] {
		t.Fatal("TestStream: unexpected Read", err, n)
	}
	n1, err = st.Read(b)
	if err != nil || string(b[:n1]) != gdata[5:10] {
		t.Fatal("TestStream: unexpected Read", err, n)
	}
	n1, err = st.Read(b)
	if err != nil || string(b[:n1]) != gdata[10:] {
		t.Fatal("TestStream: unexpected Read", err, n1)
	}
	n1, err = st.Read(b)
	if err != io.EOF || n1 != 0 {
		t.Fatal("TestStream: unexpected Read", err, n1)
	}

	w.Reset()
	st.Seek(1, 0)
	n, err = io.Copy(w, st)
	if err != nil || string(w.Bytes()) != gdata[1:] {
		t.Fatal("TestStream: unexpected io.Copy", err, n)
	}

	w.Reset()
	st.Seek(1, 0)
	st.Seek(1, 1)
	n, err = io.Copy(w, st)
	if err != nil || string(w.Bytes()) != gdata[2:] {
		t.Fatal("TestStream: unexpected io.Copy", err, n)
	}

	w.Reset()
	err = st.RangeRead(w, 0, int64(len(gdata)))
	if err != nil || string(w.Bytes()) != gdata {
		t.Fatal("TestStream: unexpected RangeRead", err, w.Bytes())
	}

	w.Reset()
	err = st.RangeRead(w, 1, 10)
	if err != nil || string(w.Bytes()) != gdata[1:10] {
		t.Fatal("TestStream: unexpected RangeRead", err, w.Bytes())
	}

	w.Reset()
	err = st.RangeRead(w, 1, int64(len(gdata))+1)
	if err != nil || string(w.Bytes()) != gdata[1:] {
		t.Fatal("TestStream: unexpected RangeRead", err, w.Bytes())
	}

	w.Reset()
	err = st.RangeRead(w, int64(len(gdata)), int64(len(gdata)+2))
	if err != syscall.EINVAL {
		t.Fatal("TestStream: unexpected RangeRead", err, w.Bytes())
	}
}

// -----------------------------------------------------------------------------

type getter struct {
	datas map[string]string
}

func (g *getter) Get(xl rpc.Logger, key []byte, w io.Writer, from, to int64) (int64, error) {

	data, ok := g.datas[string(key)]
	if !ok {
		return 0, errors.New("not found")
	}

	length := int64(len(data))
	if from >= length {
		return 0, errors.New("EOF")
	}

	var err error
	if to > length {
		to = length
		err = errors.New("UnexpectedEOF")
	}

	n, err1 := io.Copy(w, strings.NewReader(data[from:to]))
	if err1 != nil {
		err = err1
	}
	return n, err
}
