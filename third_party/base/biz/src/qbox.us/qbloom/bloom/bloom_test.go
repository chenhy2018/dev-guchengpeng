package bloom_test

import (
	"os"
	"qbox.us/qbloom/bloom"
	"testing"
)

func Test(t *testing.T) {

	file := os.Getenv("HOME") + "/qbloomTest.bloom"
	os.Remove(file)

	const (
		m = 239626459
		k = 17
	)

	keys := []string{
		"L", "Lo", "Lov", "Love", "Love!", "Love!!", "Love!!!",
	}
	{
		filter, err := bloom.Open(file, m, k)
		if err != nil {
			t.Fatal("bloom.Open failed:", err)
		}

		for _, key1 := range keys {
			println(key1)
			key := []byte(key1)
			old, err := filter.TestAndAdd(key)
			if err != nil {
				t.Fatal("filter.TestAndAdd failed:", err)
			}
			println(key, old)
			fnew, err := filter.Test(key)
			if err != nil {
				t.Fatal("filter.Test failed:", err)
			}
			if old || !fnew {
				t.Fatal("bloom Add/Test failed:", key1)
			}
		}
		filter.Close()
	}
	{
		filter, err := bloom.Open(file, m, k)
		if err != nil {
			t.Fatal("bloom.Open failed:", err)
		}

		for _, key1 := range keys {
			println(key1)
			key := []byte(key1)
			old, err := filter.TestAndAdd(key)
			if err != nil {
				t.Fatal("filter.TestAndAdd failed:", err)
			}
			println(key, old)
			fnew, err := filter.Test(key)
			if err != nil {
				t.Fatal("filter.Test failed:", err)
			}
			if !old || !fnew {
				t.Fatal("bloom Add/Test failed:", key1)
			}
		}
		filter.Close()
	}
}
