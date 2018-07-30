package mongokv

import (
	"fmt"
	"launchpad.net/mgo"
	"os"
)

type Config struct {
	mgo.Collection
}

type M map[string]interface{}

func OpenConfig(c mgo.Collection) (r *Config, err os.Error) {

	err = c.EnsureIndex(mgo.Index{Key: []string{"k"}, Unique: true})
	if err != nil {
		fmt.Println("mongokv.EnsureIndex:", err)
		return
	}

	return &Config{c}, nil
}

type entryInt64 struct {
	Val int64 "v"
}

type entryInt struct {
	Val int "v"
}

type entryString struct {
	Val string "v"
}

type entryBytes struct {
	Val []byte "v"
}

var entrySelector = M{"v": 1}

func (r *Config) Get(k string, v1 interface{}) (err os.Error) {
	sel := r.Find(M{"k": k}).Select(entrySelector)
	switch v := v1.(type) {
	case *int:
		var e entryInt
		err = sel.One(&e)
		if err == nil {
			*v = e.Val
		}
	case *string:
		var e entryString
		err = sel.One(&e)
		if err == nil {
			*v = e.Val
		}
	case *int64:
		var e entryInt64
		err = sel.One(&e)
		if err == nil {
			*v = e.Val
		}
	case *int32:
		var e entryInt
		err = sel.One(&e)
		if err == nil {
			*v = int32(e.Val)
		}
	case *[]byte:
		var e entryBytes
		err = sel.One(&e)
		if err == nil {
			*v = e.Val
		}
	default:
		err = os.EINVAL
	}
	return
}

func (r *Config) Put(k string, v interface{}) (err os.Error) {
	_, err = r.Upsert(M{"k": k}, M{"k": k, "v": v})
	return
}
