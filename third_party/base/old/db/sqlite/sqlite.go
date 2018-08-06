// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sqlite provides access to the SQLite library, version 3.
package sqlite

/*
#include <sqlite3.h>
#include <stdlib.h>

static int my_exec(sqlite3* db, char* sql) {
	return sqlite3_exec(db, (const char*)sql, NULL, NULL, NULL);
}
*/
import "C"

import (
	"fmt"
	"os"
	"time"
	"unsafe"
)

type Errno int

func (e Errno) String() string {
	s := errText[e]
	if s == "" {
		return fmt.Sprintf("errno %d", int(e))
	}
	return s
}

var (
	ErrError      os.Error = Errno(1)   //    /* SQL error or missing database */
	ErrInternal   os.Error = Errno(2)   //    /* Internal logic error in SQLite */
	ErrPerm       os.Error = Errno(3)   //    /* Access permission denied */
	ErrAbort      os.Error = Errno(4)   //    /* Callback routine requested an abort */
	ErrBusy       os.Error = Errno(5)   //    /* The database file is locked */
	ErrLocked     os.Error = Errno(6)   //    /* A table in the database is locked */
	ErrNoMem      os.Error = Errno(7)   //    /* A malloc() failed */
	ErrReadOnly   os.Error = Errno(8)   //    /* Attempt to write a readonly database */
	ErrInterrupt  os.Error = Errno(9)   //    /* Operation terminated by sqlite3_interrupt()*/
	ErrIOErr      os.Error = Errno(10)  //    /* Some kind of disk I/O error occurred */
	ErrCorrupt    os.Error = Errno(11)  //    /* The database disk image is malformed */
	ErrFull       os.Error = Errno(13)  //    /* Insertion failed because database is full */
	ErrCantOpen   os.Error = Errno(14)  //    /* Unable to open the database file */
	ErrEmpty      os.Error = Errno(16)  //    /* Database is empty */
	ErrSchema     os.Error = Errno(17)  //    /* The database schema changed */
	ErrTooBig     os.Error = Errno(18)  //    /* String or BLOB exceeds size limit */
	ErrConstraint os.Error = Errno(19)  //    /* Abort due to constraint violation */
	ErrMismatch   os.Error = Errno(20)  //    /* Data type mismatch */
	ErrMisuse     os.Error = Errno(21)  //    /* Library used incorrectly */
	ErrNolfs      os.Error = Errno(22)  //    /* Uses OS features not supported on host */
	ErrAuth       os.Error = Errno(23)  //    /* Authorization denied */
	ErrFormat     os.Error = Errno(24)  //    /* Auxiliary database format error */
	ErrRange      os.Error = Errno(25)  //    /* 2nd parameter to sqlite3_bind out of range */
	ErrNotDB      os.Error = Errno(26)  //    /* File opened that is not a database file */
	Row                    = Errno(100) //   /* sqlite3_step() has another row ready */
	Done                   = Errno(101) //   /* sqlite3_step() has finished executing */
)

var errText = map[Errno]string{
	1:   "SQL error or missing database",
	2:   "Internal logic error in SQLite",
	3:   "Access permission denied",
	4:   "Callback routine requested an abort",
	5:   "The database file is locked",
	6:   "A table in the database is locked",
	7:   "A malloc() failed",
	8:   "Attempt to write a readonly database",
	9:   "Operation terminated by sqlite3_interrupt()*/",
	10:  "Some kind of disk I/O error occurred",
	11:  "The database disk image is malformed",
	12:  "NOT USED. Table or record not found",
	13:  "Insertion failed because database is full",
	14:  "Unable to open the database file",
	15:  "NOT USED. Database lock protocol error",
	16:  "Database is empty",
	17:  "The database schema changed",
	18:  "String or BLOB exceeds size limit",
	19:  "Abort due to constraint violation",
	20:  "Data type mismatch",
	21:  "Library used incorrectly",
	22:  "Uses OS features not supported on host",
	23:  "Authorization denied",
	24:  "Auxiliary database format error",
	25:  "2nd parameter to sqlite3_bind out of range",
	26:  "File opened that is not a database file",
	100: "sqlite3_step() has another row ready",
	101: "sqlite3_step() has finished executing",
}

func (c *Conn) error(rv C.int) os.Error {
	if c == nil || c.db == nil {
		return os.NewError("nil sqlite database")
	}
	if rv == 0 {
		return nil
	}
	if rv == 21 { // misuse
		return Errno(rv)
	}
	return os.NewError(Errno(rv).String() + ": " + C.GoString(C.sqlite3_errmsg(c.db)))
}

type Conn struct {
	db *C.sqlite3
}

func Version() string {
	p := C.sqlite3_libversion()
	return C.GoString(p)
}

func Open(filename string) (*Conn, os.Error) {
	if C.sqlite3_threadsafe() == 0 {
		return nil, os.NewError("sqlite library was not compiled for thread-safe operation")
	}

	var db *C.sqlite3
	name := C.CString(filename)
	defer C.free(unsafe.Pointer(name))
	rv := C.sqlite3_open_v2(name, &db,
		C.SQLITE_OPEN_FULLMUTEX|
			C.SQLITE_OPEN_READWRITE|
			C.SQLITE_OPEN_CREATE,
		nil)
	if rv != 0 {
		return nil, Errno(rv)
	}
	if db == nil {
		return nil, os.NewError("sqlite succeeded without returning a database")
	}
	return &Conn{db}, nil
}

type Backup struct {
	sb       *C.sqlite3_backup
	dst, src *Conn
}

func NewBackup(dst *Conn, dstTable string, src *Conn, srcTable string) (*Backup, os.Error) {
	dname := C.CString(dstTable)
	sname := C.CString(srcTable)
	defer C.free(unsafe.Pointer(dname))
	defer C.free(unsafe.Pointer(sname))

	sb := C.sqlite3_backup_init(dst.db, dname, src.db, sname)
	if sb == nil {
		return nil, dst.error(C.sqlite3_errcode(dst.db))
	}
	return &Backup{sb, dst, src}, nil
}

func (b *Backup) Step(npage int) os.Error {
	rv := C.sqlite3_backup_step(b.sb, C.int(npage))
	if rv == 0 || Errno(rv) == ErrBusy || Errno(rv) == ErrLocked {
		return nil
	}
	return Errno(rv)
}

type BackupStatus struct {
	Remaining int
	PageCount int
}

func (b *Backup) Status() BackupStatus {
	return BackupStatus{int(C.sqlite3_backup_remaining(b.sb)), int(C.sqlite3_backup_pagecount(b.sb))}
}

func (b *Backup) Run(npage int, sleepNs int64, c chan<- BackupStatus) os.Error {
	var err os.Error
	for {
		err = b.Step(npage)
		if err != nil {
			break
		}
		if c != nil {
			c <- b.Status()
		}
		time.Sleep(sleepNs)
	}
	return b.dst.error(C.sqlite3_errcode(b.dst.db))
}

func (b *Backup) Close() os.Error {
	if b.sb == nil {
		return os.EINVAL
	}
	C.sqlite3_backup_finish(b.sb)
	b.sb = nil
	return nil
}

func (c *Conn) BusyTimeout(ms int) os.Error {
	rv := C.sqlite3_busy_timeout(c.db, C.int(ms))
	if rv == 0 {
		return nil
	}
	return Errno(rv)
}

func (c *Conn) Exec(cmd string, args ...interface{}) os.Error {
	s, err := c.Prepare(cmd)
	if err != nil {
		return err
	}
	defer s.Finalize()
	return s.Exec(args...)
}

/*
** CAPI3REF: Count The Number Of Rows Modified
**
** ^This function returns the number of database rows that were changed
** or inserted or deleted by the most recently completed SQL statement
** on the [database connection] specified by the first parameter.
** ^(Only changes that are directly specified by the [INSERT], [UPDATE],
** or [DELETE] statement are counted.  Auxiliary changes caused by
** triggers or [foreign key actions] are not counted.)^ Use the
** [sqlite3_total_changes()] function to find the total number of changes
** including changes caused by triggers and foreign key actions.
**
** ^Changes to a view that are simulated by an [INSTEAD OF trigger]
** are not counted.  Only real table changes are counted.
**
** ^(A "row change" is a change to a single row of a single table
** caused by an INSERT, DELETE, or UPDATE statement.  Rows that
** are changed as side effects of [REPLACE] constraint resolution,
** rollback, ABORT processing, [DROP TABLE], or by any other
** mechanisms do not count as direct row changes.)^
**
** A "trigger context" is a scope of execution that begins and
** ends with the script of a [CREATE TRIGGER | trigger].
** Most SQL statements are
** evaluated outside of any trigger.  This is the "top level"
** trigger context.  If a trigger fires from the top level, a
** new trigger context is entered for the duration of that one
** trigger.  Subtriggers create subcontexts for their duration.
**
** ^Calling [sqlite3_exec()] or [sqlite3_step()] recursively does
** not create a new trigger context.
**
** ^This function returns the number of direct row changes in the
** most recent INSERT, UPDATE, or DELETE statement within the same
** trigger context.
**
** ^Thus, when called from the top level, this function returns the
** number of changes in the most recent INSERT, UPDATE, or DELETE
** that also occurred at the top level.  ^(Within the body of a trigger,
** the sqlite3_changes() interface can be called to find the number of
** changes in the most recently completed INSERT, UPDATE, or DELETE
** statement within the body of the same trigger.
** However, the number returned does not include changes
** caused by subtriggers since those have their own context.)^
**
** See also the [sqlite3_total_changes()] interface, the
** [count_changes pragma], and the [changes() SQL function].
**
** If a separate thread makes changes on the same database connection
** while [sqlite3_changes()] is running then the value returned
** is unpredictable and not meaningful.
 */
func (c *Conn) Changes() int {
	return int(C.sqlite3_changes(c.db))
}

func (c *Conn) BeginTransaction() os.Error {
	rv := C.my_exec(c.db, (*C.char)(stringPtr("BEGIN TRANSACTION\000")))
	if rv == 0 {
		return nil
	}
	return Errno(rv)
}

func (c *Conn) Commit() os.Error {
	rv := C.my_exec(c.db, (*C.char)(stringPtr("COMMIT\000")))
	if rv == 0 {
		return nil
	}
	return Errno(rv)
}

func (c *Conn) Rollback() os.Error {
	rv := C.my_exec(c.db, (*C.char)(stringPtr("ROLLBACK\000")))
	if rv == 0 {
		return nil
	}
	return Errno(rv)
}

//
// example:
//	  var rowId int
//	  err := db.LastInsertRowId(&rowId)
//
func (c *Conn) LastInsertRowId(id interface{}) (err os.Error) {

	cs, err := c.Prepare("SELECT LAST_INSERT_ROWID()")
	if err != nil {
		return
	}
	defer cs.Finalize()

	err = cs.Query()
	if err != nil {
		return
	}

	err = cs.NextRow(id)
	return
}

func (c *Conn) DropTable(tbl string) (err os.Error) {
	err = c.Exec(`DROP TABLE IF EXISTS ` + tbl)
	return
}

func (c *Conn) Close() os.Error {
	if c == nil || c.db == nil {
		return nil
	}
	rv := C.sqlite3_close(c.db)
	if rv != 0 {
		return c.error(rv)
	}
	c.db = nil
	return nil
}
