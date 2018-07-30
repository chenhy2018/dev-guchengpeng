package cf

import (
	"fmt"
	"launchpad.net/mgo"
	cfc "qbox.us/api/cf"
	"testing"
	"time"
)

type S struct {
	A int
	B string
}

func TestCf(t *testing.T) {

	session, err := mgo.Mongo("localhost")
	if err != nil {
		fmt.Println("\nConnect MongoDB Error:", err)
		return
	}
	defer session.Close()
	db := session.DB("qbox_test")
	c := db.C("cf")
	defer c.RemoveAll(M{})

	go Run("localhost:8111", &Config{c})
	time.Sleep(3E9)

	c1 := cfc.New("http://localhost:8111", "a")
	c2 := cfc.New("http://localhost:8111", "b")

	s := S{1, "a"}
	var s1 S

	err = c1.Set([]byte("key1"), s)
	if err != nil {
		fmt.Println(err)
	}
	err = c1.Get(&s1, []byte("key1"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(s1)

	err = c2.Set([]byte("key1"), s)
	if err != nil {
		fmt.Println(err)
	}
	err = c2.Get(&s1, []byte("key1"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(s1)

	err = c1.Del([]byte("key1"))
	if err != nil {
		fmt.Println(err)
	}
	err = c1.Get(&s1, []byte("key1"))
	if err != nil {
		fmt.Println(err)
	}

	err = c2.Get(&s1, []byte("key1"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(s1)
}
