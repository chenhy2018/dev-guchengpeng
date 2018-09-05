package directio

import (
	"fmt"
	"io"
	"os"
	"github.com/qiniu/ts"
	"testing"
)

func doTestIO(data []byte, f *os.File, off, l int64, bufBits uint, t *testing.T) {

	mask := int64((1 << bufBits) - 1)

	r := NewReaderSize(f, off, l, bufBits)
	fmt.Println(len(r.buf), r.off, r.limit, r.pos, r.lastErr)
	if (r.off & mask) != 0 {
		ts.Fatal(t, "round fail")
	}

	b := make([]byte, 4096)
	n, err := io.ReadFull(r, b[:3])
	if n != 3 || err != nil {
		ts.Fatal(t, n, err)
	}
	for i := 0; i < n; i++ {
		if b[i] != byte(off+int64(i)) {
			ts.Fatal(t, "read fail")
		}
	}

	fmt.Println(len(r.buf), r.off, r.limit, r.pos, r.lastErr)
	if (r.off & mask) != 0 {
		ts.Fatal(t, "round fail")
	}
	if r.pos != 3 {
		ts.Fatal(t, "read fail")
	}

	n, err = io.ReadFull(r, b)
	if n != len(b) || err != nil {
		ts.Fatal(t, "read fail:", n, err)
	}

	fmt.Println(len(r.buf), r.off, r.limit, r.pos, r.lastErr)
	if (r.off & mask) != 0 {
		ts.Fatal(t, "round fail")
	}

	n, err = io.ReadFull(r, b)
	if n != 4093 || err != io.ErrUnexpectedEOF {
		ts.Fatal(t, "read fail:", n, err)
	}
	fmt.Println(len(r.buf), r.off, r.limit, r.pos, r.lastErr)
	if r.off != r.limit {
		ts.Fatal(t, "round fail")
	}
}

func TestIO(t *testing.T) {

	fname := os.Getenv("HOME") + "/directTest.data"
	f, err := os.Create(fname)
	if err != nil {
		ts.Fatal(t, err)
	}

	b := make([]byte, 8192*2)
	for i := 0; i < len(b); i++ {
		b[i] = byte(i)
	}
	_, err = f.Write(b)
	if err != nil {
		ts.Fatal(t, err)
	}
	f.Close()

	f, err = os.Open(fname)
	if err != nil {
		ts.Fatal(t, err)
	}
	defer f.Close()

	doTestIO(b, f, 3, 8192, 12, t)
}
