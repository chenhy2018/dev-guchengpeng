package acc

import (
	"fmt"
	"testing"
)

// ------------------------------------------------------------------------

func Test(t *testing.T) {

	id := MakeId(123)
	fmt.Println(id)

	uid, err := ParseId(id)
	if err != nil || uid != 123 {
		t.Fatal("ParseId failed:", uid, err)
	}
}

// ------------------------------------------------------------------------
