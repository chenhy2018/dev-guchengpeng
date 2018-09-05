package bloom_test

import (
	"qbox.us/qmset/bloom"
	"testing"
)

func Test(t *testing.T) {

	filter := bloom.New(29, 5)

	keys := []string{
		"L", "Lo", "Lov", "Love", "Love!", "Love!!", "Love!!!",
	}
	for _, key1 := range keys {
		println(key1)
		key := []byte(key1)
		old := filter.TestAndAdd(key)
		println(key, old)
		if old || !filter.Test(key) {
			t.Fatal("bloom Add/Test failed:", key1)
		}
	}
}
