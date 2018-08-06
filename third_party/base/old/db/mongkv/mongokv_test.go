package mongokv

import (
	"fmt"
	"launchpad.net/mgo"
	"qbox.us/ts"
	"testing"
)

func TestMongoKV(t *testing.T) {

	session, err := mgo.Mongo("localhost")
	if err != nil {
		ts.Fatal(t, err)
	}
	defer session.Close()

	session.SetSafe(&mgo.Safe{})

	c := session.DB("qbox_mongokv").C("config")
	c.RemoveAll(M{})
	cdb, err := OpenConfig(c)
	if err != nil {
		ts.Fatal(t, "OpenConfig:", err)
	}

	err = cdb.Put("bv", 1234)
	if err != nil {
		ts.Fatal(t, "Put:", err)
	}

	var v int64
	err = cdb.Get("bv", &v)
	fmt.Println("bv:", v, err)
	if err != nil || v != 1234 {
		ts.Fatal(t, "Get:", v, err)
	}

	err = cdb.Put("bv", "abcd")
	if err != nil {
		ts.Fatal(t, "Put:", err)
	}

	var v2 string
	err = cdb.Get("bv", &v2)
	fmt.Println("bv:", v2, err)
	if err != nil || v2 != "abcd" {
		ts.Fatal(t, "Get:", v2, err)
	}
}
