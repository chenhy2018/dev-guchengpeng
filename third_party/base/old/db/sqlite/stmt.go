// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sqlite provides access to the SQLite library, version 3.
package sqlite

/*
#include <sqlite3.h>
#include <stdlib.h>

// These wrappers are necessary because SQLITE_STATIC
// is a pointer constant, and cgo doesn't translate them correctly.
// The definition in sqlite3.h is:
//
// typedef void (*sqlite3_destructor_type)(void*);
// #define SQLITE_STATIC      ((sqlite3_destructor_type)0)
// #define SQLITE_TRANSIENT   ((sqlite3_destructor_type)-1)

static int my_bind_text(sqlite3_stmt *stmt, int n, char *val, int length) {
	return sqlite3_bind_text(stmt, n, val, length, SQLITE_STATIC);
}

static int my_bind_blob(sqlite3_stmt *stmt, int n, void *val, int length) {
	return sqlite3_bind_blob(stmt, n, val, length, SQLITE_STATIC);
}
*/
import "C"

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"unsafe"
)

type Stmt struct {
	Impl *C.sqlite3_stmt
}

func stringPtr(s string) unsafe.Pointer {
	return unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)
}

func bytesPtr(b []byte) unsafe.Pointer {
	return unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data)
}

func (c *Conn) Prepare(cmd string) (s Stmt, err os.Error) {
	if c == nil || c.db == nil {
		err = os.NewError("nil sqlite database")
		return
	}
	var stmt *C.sqlite3_stmt
	var tail *C.char
	rv := C.sqlite3_prepare_v2(c.db, (*C.char)(stringPtr(cmd)), C.int(len(cmd)), &stmt, &tail)
	if rv != 0 {
		err = Errno(rv)
		return
	}
	s = Stmt{stmt}
	return
}

func (s Stmt) Good() bool {
	return s.Impl != nil
}

func (s Stmt) Finalize() os.Error {
	rv := C.sqlite3_finalize(s.Impl)
	if rv != 0 {
		return Errno(rv)
	}
	return nil
}

func (s Stmt) Exec(args ...interface{}) os.Error {
	err := s.Query(args...)
	if err != nil {
		return err
	}
	rv := C.sqlite3_step(s.Impl)
	if Errno(rv) != Done {
		return Errno(rv)
	}
	return nil
}

func (s Stmt) Query(args ...interface{}) os.Error {
	n := int(C.sqlite3_bind_parameter_count(s.Impl))
	if n != len(args) {
		fmt.Println("Stmt.Query require arguments:", n)
		return os.EINVAL
	}

	rv := C.sqlite3_reset(s.Impl)
	if rv != 0 {
		err := Errno(rv)
		fmt.Println("Stmt.Query failed:", err)
		return err
	}

	for i, v := range args {
		switch v := v.(type) {
		case int:
			rv = C.sqlite3_bind_int(s.Impl, C.int(i+1), C.int(v))
		case string:
			rv = C.my_bind_text(s.Impl, C.int(i+1), (*C.char)(stringPtr(v)), C.int(len(v)))
		case int64:
			rv = C.sqlite3_bind_int64(s.Impl, C.int(i+1), C.sqlite3_int64(v))
		case []byte:
			rv = C.my_bind_blob(s.Impl, C.int(i+1), bytesPtr(v), C.int(len(v)))
		case int32:
			rv = C.sqlite3_bind_int(s.Impl, C.int(i+1), C.int(v))
		case bool:
			var iv C.int
			if v {
				iv = 1
			}
			rv = C.sqlite3_bind_int(s.Impl, C.int(i+1), iv)
		case float64:
			rv = C.sqlite3_bind_double(s.Impl, C.int(i+1), C.double(v))
		default:
			fmt.Println("Stmt.Query argument", i, ": type is not valid")
			return os.EINVAL
		}
		if rv != 0 {
			return Errno(rv)
		}
	}
	return nil
}

func (s Stmt) Reset() os.Error {
	C.sqlite3_reset(s.Impl)
	return nil
}

func (s Stmt) NextRow(args ...interface{}) os.Error {
	err := s.Next()
	if err != nil {
		return err
	}
	return s.Scan(args...)
}

func (s Stmt) Next() os.Error {
	rv := C.sqlite3_step(s.Impl)
	err := Errno(rv)
	if err == Row {
		return nil
	} else if err == Done {
		return Done
	}
	return err
}

func (s Stmt) Scan(args ...interface{}) os.Error {

	n := int(C.sqlite3_column_count(s.Impl))
	if n != len(args) {
		return os.EINVAL
	}

	for i, v := range args {
		n := C.sqlite3_column_bytes(s.Impl, C.int(i))
		p := C.sqlite3_column_blob(s.Impl, C.int(i))
		if p == nil && n > 0 {
			return os.NewError("got nil blob")
		}
		var data []byte
		if n > 0 {
			data = (*[1 << 30]byte)(unsafe.Pointer(p))[0:n]
		}
		switch v := v.(type) {
		case *[]byte:
			*v = data
		case *string:
			*v = string(data)
		case *bool:
			*v = string(data) == "1"
		case *int:
			x, err := strconv.Atoi(string(data))
			if err != nil {
				return os.NewError("arg " + strconv.Itoa(i) + " as int: " + err.String())
			}
			*v = x
		case *int64:
			x, err := strconv.Atoi64(string(data))
			if err != nil {
				return os.NewError("arg " + strconv.Itoa(i) + " as int64: " + err.String())
			}
			*v = x
		case *float64:
			x, err := strconv.Atof64(string(data))
			if err != nil {
				return os.NewError("arg " + strconv.Itoa(i) + " as float64: " + err.String())
			}
			*v = x
		default:
			return os.NewError("unsupported type in Scan: " + reflect.TypeOf(v).String())
		}
	}
	return nil
}
