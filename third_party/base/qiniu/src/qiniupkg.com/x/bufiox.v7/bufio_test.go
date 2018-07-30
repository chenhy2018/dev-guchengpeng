package bufiox

import (
	"bufio"
	"reflect"
	"testing"
	"unsafe"
)

// ---------------------------------------------------

func TestSizeOf(t *testing.T) {

	var r reader
	var b bufio.Reader
	if unsafe.Sizeof(r) != unsafe.Sizeof(b) {
		t.Fatal("unsafe.Sizeof(r) != unsafe.Sizeof(b)")
	}
}

func TestReader(t *testing.T) {

	text := []byte("Hello, buf")
	br := NewReaderBuffer(text)
	if br.Buffered() != len(text) {
		t.Fatal("br.Buffered() != len(text)")
	}
	if c, err := br.ReadByte(); err != nil || c != 'H' {
		t.Fatal("ReadByte:", c, err)
	}
	if b, err := br.Peek(2); err != nil || b[0] != 'e' || b[1] != 'l' {
		t.Fatal("Peek:", b, err)
	}
	if s, err := br.ReadString(','); err != nil || s != "ello," {
		t.Fatal("ReadString:", s, err)
	}
	if br.Buffered() != 4 {
		t.Fatal("br.Buffered() != 0")
	}
}

// ---------------------------------------------------

func TestReflectEqual(t *testing.T) {

	var r reader
	var b bufio.Reader
	ty1 := reflect.TypeOf(r)
	ty2 := reflect.TypeOf(b)
	n1 := ty1.NumField()
	n2 := ty2.NumField()
	if n1 != n2 {
		t.Fatal("numField(reader) != numField(bufio.Reader)")
	}
	for i := 0; i < n1; i++ {
		sf1 := ty1.Field(i)
		sf2 := ty2.Field(i)
		if sf1.Name != sf2.Name {
			t.Fatal("sf1.Name != sf2.Name")
		}
		if sf1.Type != sf2.Type {
			t.Fatal("sf1.Type != sf2.Type")
		}
		if sf1.Offset != sf2.Offset {
			t.Fatal("sf1.Offset != sf2.Offset")
		}
	}
}

// ---------------------------------------------------
