package safefile

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func readNoneTest(safePath string, t *testing.T) {
	_, err := Open(safePath)
	if err == nil {
		t.Fatal("read none should fail")
	}
}

var (
	s1 = "123456"
	s2 = "654321"
)

func writeNoneTest(safePath string, t *testing.T) {
	safe, err := Create(safePath)
	if err != nil {
		t.Fatal("write should OK ", err)
	}
	safe.Write([]byte(s1))
	if f1, err := Open(safePath); err == nil {
		f1.Close()
		safe.Close()
		t.Fatal("no write complete, read should fail")
	}
	safe.Close()
	readCheck(safePath, s1, t)
}

func writeExistTest(safePath string, t *testing.T) {
	safe, err := Create(safePath)
	if err != nil {
		t.Fatal("write should OK ", err)
	}
	safe.Write([]byte(s2))
	readCheck(safePath, s1, t)
	safe.Close()
	readCheck(safePath, s2, t)
}

func writeExistFailTest(safePath string, t *testing.T) {
	_, err := Create(safePath)
	if err == nil {
		t.Fatal("write should fail")
	}
}

func readExistFailTest(safePath string, t *testing.T) {
	_, err := Open(safePath)
	if err == nil {
		t.Fatal("read none should fail")
	}
}

func readCheck(safePath string, s string, t *testing.T) {
	f, err := Open(safePath)
	defer f.Close()
	if err != nil {
		t.Fatal("read none should fail")
	}
	x, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("write complete, read should OK")
	}
	if s != string(x) {
		t.Fatal("read write Not equal")
	}
}

func TestAll(t *testing.T) {
	safePath, err := ioutil.TempDir("", "testsafefile")
	if err != nil {
		t.Log("open temp dir fail")
		return
	}
	safePath = path.Join(safePath, "test")
	readNoneTest(safePath, t)
	writeNoneTest(safePath, t)
	writeExistTest(safePath, t)

	writeExistFailTest(path.Join(safePath, "index"), t)
	readExistFailTest(path.Join(safePath, "index"), t)
	os.RemoveAll(safePath)
}

func writeIndexFailTest(safePath string, cur int, t *testing.T) {
	err := writeIndex(safePath, cur)
	if err == nil {
		t.Fatal("read should fail")
	}
}

func readIndexFailTest(safePath string, cur int, t *testing.T) {
	_, err := readIndex(safePath)
	if err == nil {
		t.Fatal("read should fail")
	}
}

func writeIndexTest(safePath string, cur int, t *testing.T) {
	err := writeIndex(safePath, cur)
	if err != nil {
		t.Fatal("write should not fail", err)
	}
	readIndexTest(safePath, cur, t)
}

func readIndexTest(safePath string, cur int, t *testing.T) {
	c, err := readIndex(safePath)
	if err != nil {
		t.Fatal("read should not fail", err)
	}
	if c != cur {
		t.Fatal("index number not equal")
	}
}

func TestIndex(t *testing.T) {
	safePath := "/a/b/c/d"
	readIndexFailTest(safePath, 0, t)
	writeIndexFailTest(safePath, 0, t)
	safePath, err := ioutil.TempDir("", "testsafefile")
	if err != nil {
		t.Log("open temp dir fail")
		return
	}
	writeIndexTest(safePath, 0, t)
	os.RemoveAll(safePath)
}
