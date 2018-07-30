// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Use temporary database - http://www.sqlite.org/inmemorydb.html

package sqlite

import (
	"qbox.us/ts"
	"testing"
)

func TestOpen(t *testing.T) {
	db, err := Open("")
	if err != nil {
		t.Errorf("couldn't open database file: %s", err)
	}
	if db == nil {
		t.Error("opened database is nil")
	}
	db.Close()
}

func TestCreateTable(t *testing.T) {
	db, err := Open("")
	db.Exec("DROP TABLE test")
	err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY AUTOINCREMENT," +
		" float_num REAL, int_num INTEGER, a_string TEXT)")
	if err != nil {
		ts.Errorf(t, "error creating table: %s", err)
	}
}

func TestInsert(t *testing.T) {
	db, _ := Open("")
	db.Exec("DROP TABLE test")
	db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY AUTOINCREMENT," +
		" float_num REAL, int_num INTEGER, a_string TEXT)")
	for i := 0; i < 1000; i++ {
		ierr := db.Exec("INSERT INTO test (float_num, int_num, a_string)"+
			" VALUES (?, ?, ?)", float64(i)*float64(3.14), i, "hello")
		if ierr != nil {
			ts.Fatalf(t, "insert error: %s", ierr)
		}
	}

	cs, _ := db.Prepare("SELECT COUNT(*) FROM test")
	cs.Query()
	if cs.Next() != nil {
		t.Error("no result for count")
	}
	var i int
	err := cs.Scan(&i)
	if err != nil {
		t.Errorf("error scanning count: %s", err)
	}
	if i != 1000 {
		t.Errorf("count should be 1000, but it is %d", i)
	}
}

func TestInsertWithStatement(t *testing.T) {
	db, _ := Open("")
	defer db.Close()
	db.Exec("DROP TABLE test")
	db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY AUTOINCREMENT," +
		" float_num REAL, int_num INTEGER, a_string TEXT)")
	s, serr := db.Prepare("INSERT INTO test (float_num, int_num, a_string)" +
		" VALUES (?, ?, ?)")
	if serr != nil {
		ts.Errorf(t, "prepare error: %s", serr)
	}
	if !s.Good() {
		ts.Error(t, "statement is nil")
	}

	for i := 0; i < 1000; i++ {
		ierr := s.Exec(float64(i)*float64(3.14), i, "hello")
		if ierr != nil {
			ts.Fatal(t, "insert error:", ierr)
		}
	}
	s.Finalize()

	cs, _ := db.Prepare("SELECT COUNT(*) FROM test")
	cs.Query()
	if cs.Next() != nil {
		t.Error("no result for count")
	}
	var i int
	err := cs.Scan(&i)
	if err != nil {
		t.Errorf("error scanning count: %s", err)
	}
	if i != 1000 {
		t.Errorf("count should be 1000, but it is %d", i)
	}
}

func BenchmarkScan(b *testing.B) {
	b.StopTimer()
	db, _ := Open("")
	db.Exec("DROP TABLE test")
	db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY AUTOINCREMENT," +
		" float_num REAL, int_num INTEGER, a_string TEXT)")
	s, _ := db.Prepare("INSERT INTO test (float_num, int_num, a_string)" +
		" VALUES (?, ?, ?)")

	for i := 0; i < 1000; i++ {
		s.Exec(float64(i)*float64(3.14), i, "hello")
	}
	s.Finalize()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cs, _ := db.Prepare("SELECT float_num, int_num, a_string FROM test")
		cs.Query()

		var fnum float64
		var inum int64
		var sstr string

		for cs.Next() == nil {
			cs.Scan(&fnum, &inum, &sstr)
		}
	}
}

func BenchmarkInsert(b *testing.B) {
	db, _ := Open("")
	db.Exec("DROP TABLE test")
	db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY AUTOINCREMENT," +
		" float_num REAL, int_num INTEGER, a_string TEXT)")
	s, _ := db.Prepare("INSERT INTO test (float_num, int_num, a_string)" +
		" VALUES (?, ?, ?)")

	for i := 0; i < b.N; i++ {
		s.Exec(float64(i)*float64(3.14), i, "hello")
	}
	s.Finalize()
}
