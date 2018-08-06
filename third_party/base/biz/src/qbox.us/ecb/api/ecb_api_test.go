package api

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"

	. "qbox.us/ebd/api/types"
)

func TestStripeRgetInfo(t *testing.T) {
	srgi := &StripeRgetInfo{
		Soff:    4321,
		Bsize:   32222,
		BadSuid: ^uint64(0),
		Psects:  [N + M]uint64{},
	}
	for i := range srgi.Psects {
		srgi.Psects[i] = uint64(rand.Int63())
	}

	b, err := EncodeStripeRgetInfo(srgi)
	if err != nil {
		t.Fatal(err)
	}
	srgi2, err := ReadStripeRgetInfo(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(srgi, srgi2) {
		t.Fatal("srgi, srgi2 not equal")
	}
}
