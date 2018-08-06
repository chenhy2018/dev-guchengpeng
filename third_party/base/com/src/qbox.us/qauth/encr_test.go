package qauth

import (
	"crypto/aes"
	"github.com/qiniu/ts"
	"syscall"
	"testing"
)

var e = newEncryptor("fwafewagheheyegfejf583454271563266")

func TestEncr(t *testing.T) {

	var decoded map[string]interface{}

	orig := map[string]interface{}{"a": 1, "b": "c", "d": 1.2}

	encoded1, err := e.encode(orig, 0x53263523)
	if err != nil {
		t.Errorf("encoded is: %s", err)
		return
	}
	{
		encoded := make([]byte, len(encoded1))
		copy(encoded, encoded1)
		err = e.decode(&decoded, encoded[1:])
		if err == nil || err != syscall.EINVAL {
			ts.Fatal(t, "decoded ok?")
		}
	}
	{
		encoded := make([]byte, len(encoded1))
		copy(encoded, encoded1)
		err = e.decode(&decoded, encoded)
		if err != nil {
			ts.Fatal(t, "decoded is:", err)
		}
	}
	{
		encoded := make([]byte, len(encoded1))
		copy(encoded, encoded1)
		err = e.decode(&decoded, encoded[:len(encoded)-aes.BlockSize])
		if err == nil || err != errDecode {
			ts.Fatal(t, "decoded:", err)
		}
	}
	if decoded == nil {
		t.Errorf("decoded map is null")
		return
	}

	if len(decoded) != 3 {
		t.Errorf("len was %d, expected 3", len(decoded))
		return
	}

	for k, v := range orig {
		if decoded[k] != v {
			t.Errorf("expected decoded[%s] (%#v) == %#v", k,
				decoded[k], v)
		}
	}
}
