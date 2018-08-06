package dcutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestMultiReaderAt(t *testing.T) {
	head := []byte("01234")
	r := bytes.NewReader([]byte("56789"))
	mr := MultiReaderAt(head, r)
	tests := []struct {
		off     int64
		n       int
		want    string
		wanterr interface{}
	}{
		{0, 5, "01234", nil},
		{0, 6, "012345", nil},
		{0, 10, "0123456789", nil},
		{5, 5, "56789", nil},
		{1, 10, "123456789", io.EOF},
		{1, 9, "123456789", nil},
		{11, 10, "", io.EOF},
		{0, 0, "", nil},
		{-1, 0, "", "multiReaderAt.ReadAt: negative off"},
	}
	for i, tt := range tests {
		b := make([]byte, tt.n)
		rn, err := mr.ReadAt(b, tt.off)
		got := string(b[:rn])
		if got != tt.want {
			t.Errorf("%d. got %q; want %q", i, got, tt.want)
		}
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.wanterr) {
			t.Errorf("%d. got error = %v; want %v", i, err, tt.wanterr)
		}
	}
}

func TestMultiReader(t *testing.T) {
	var mr *multiReaderAt
	nread := 0
	withFooBar := func(tests func()) {
		head := []byte("foo ")
		r := strings.NewReader("bar")
		mr = MultiReaderAt(head, r)
		tests()
	}
	expectRead := func(size int, expected string, eerr error) {
		nread++
		buf := make([]byte, size)
		n, gerr := mr.Read(buf)
		if n != len(expected) {
			t.Errorf("#%d, expected %d bytes; got %d",
				nread, len(expected), n)
		}
		got := string(buf[:n])
		if got != expected {
			t.Errorf("#%d, expected %q; got %q",
				nread, expected, got)
		}
		if gerr != eerr {
			t.Errorf("#%d, expected error %v; got %v",
				nread, eerr, gerr)
		}
	}
	withFooBar(func() {
		expectRead(2, "fo", nil)
		expectRead(5, "o bar", nil)
		expectRead(1, "", io.EOF)
	})
}

func TestMultiReaderCopy(t *testing.T) {
	r := MultiReaderAt([]byte("hello"), strings.NewReader(" world"))
	data, err := ioutil.ReadAll(r)
	if err != nil || string(data) != "hello world" {
		t.Errorf("ReadAll() = %q, %v, want %q, nil", data, err, "hello world")
	}
}
