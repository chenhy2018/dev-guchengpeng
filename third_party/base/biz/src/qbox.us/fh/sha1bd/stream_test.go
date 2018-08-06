package sha1bd

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"qbox.us/fh/proto"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {

	cs := int64(4 * 1024 * 1024)
	size := 2*cs + 3*123
	data := make([]byte, 0, int(size))
	datas := make([][]byte, 3)
	datas[0] = bytes.Repeat([]byte("12345678"), 512*1024)
	datas[1] = bytes.Repeat([]byte("2345"), 1024*1024)
	datas[2] = bytes.Repeat([]byte("678"), 123)

	g := &getter{
		datas: make(map[string][]byte),
	}
	keys := make([]byte, 0, 60)
	for i := 0; i < 3; i++ {
		key := bytes.Repeat([]byte{byte(i)}, 20)
		g.datas[string(key)] = datas[i]
		keys = append(keys, key...)
		data = append(data, datas[i]...)
	}

	xl := xlog.NewDummy()
	fh := append([]byte{0, 1, 255, 255, 255, 255, 0, 1, 0, 0x16}, keys...)
	_, err := NewStream(xl, g, fh[1:], size)
	assert.Equal(t, proto.ErrInvalidFh, err)
	fh[9] = 0x96
	_, err = NewStream(xl, g, fh, size)
	assert.Equal(t, proto.ErrInvalidFh, err)
	fh[9] = 0x16

	bds := [4]uint16{256, 65535, 65535, 1}
	for i := 0; i < 2; i++ {
		st, _ := NewStream(xl, g, fh, size)
		if fh1, _ := st.QueryFhandle(); !bytes.Equal(fh1, fh) {
			t.Fatal(i, "TestStream: unexpected fh", fh1)
		}
		if size1, _ := st.UploadedSize(); size1 != size {
			t.Fatal(i, "TestStream: unexpected size", size1)
		}
		assert.Equal(t, bds, st.bds)

		b := make([]byte, 20)
		n, err := st.Read(b)
		if err != nil || !bytes.Equal(b, datas[0][:20]) {
			t.Fatal(i, "TestStream: unexpected Read", n, err)
		}
		n, err = st.Read(b)
		if err != nil || !bytes.Equal(b, datas[0][20:40]) {
			t.Fatal(i, "TestStream: unexpected Read", n, err)
		}
		n1, err := st.Seek(size-5, 0)
		if err != nil || n1 != size-5 {
			t.Fatal(i, "TestStream: unexpected Seek", n1, err)
		}
		n, err = st.Read(b)
		if err != nil || !bytes.Equal(b[:n], datas[2][123*3-5:]) {
			t.Fatal(i, "TestStream: unexpected Read", n, err, string(b[:n]))
		}
		n, err = st.Read(b)
		if err != io.EOF || n != 0 {
			t.Fatal(i, "TestStream: unexpected Read", n, err)
		}

		w := bytes.NewBuffer(nil)
		n1, err = st.Seek(-size, 2)
		if err != nil || n1 != 0 {
			t.Fatal(i, "TestStream: unexpected Seek", n1, err)
		}
		n1, err = io.Copy(w, st)
		if err != nil || n1 != size || !bytes.Equal(w.Bytes(), data) {
			t.Fatal(i, "TestStream: unexpected Copy", n1, err)
		}

		w.Reset()
		st.Seek(1, 0)
		n1, err = st.Seek(1, 1)
		if err != nil || n1 != 2 {
			t.Fatal(i, "TestStream: unexpected Seek", n1, err)
		}
		n1, err = io.Copy(w, st)
		if err != nil || n1 != size-2 || !bytes.Equal(w.Bytes(), data[2:]) {
			t.Fatal(i, "TestStream: unexpected Copy", n1, err)
		}

		w.Reset()
		err = st.RangeRead(w, 0, size)
		if err != nil || !bytes.Equal(w.Bytes(), data) {
			t.Fatal(i, "TestStream: unexpected RangeRead", err, w.Len())
		}

		w.Reset()
		err = st.RangeRead(w, 1, 12)
		if err != nil || !bytes.Equal(w.Bytes(), data[1:12]) {
			t.Fatal(i, "TestStream: unexpected RangeRead", err, w.Len())
		}

		w.Reset()
		err = st.RangeRead(w, 123, cs+15)
		if err != nil || !bytes.Equal(w.Bytes(), data[123:cs+15]) {
			t.Fatal(i, "TestStream: unexpected RangeRead", err, w.Len())
		}

		w.Reset()
		err = st.RangeRead(w, 234, size+12)
		if err != nil || !bytes.Equal(w.Bytes(), data[234:]) {
			t.Fatal(i, "TestStream: unexpected RangeRead", err, w.Len())
		}

		// For large key test.
		key := bytes.Repeat([]byte{4}, 20)
		g.datas[string(key)] = keys
		fh = append([]byte{2, 0, 0x96}, key...)
		bds = [4]uint16{0, 65535, 65535, 2}
	}
}

// -----------------------------------------------------------------------------

type getter struct {
	datas map[string][]byte
}

func (g *getter) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {

	xl.Info("getter.Get:", key, from, to)
	data, ok := g.datas[string(key)]
	if !ok {
		return errors.New("not found")
	}

	if from >= len(data) {
		return io.EOF
	}

	var err error
	if to > len(data) {
		to = len(data)
	}

	_, err1 := io.Copy(w, bytes.NewReader(data[from:to]))
	if err1 != nil {
		err = err1
	}
	return err
}
