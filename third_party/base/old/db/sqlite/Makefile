# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=qbox.us/sqlite

CGOFILES=\
	sqlite.go\
	stmt.go\

ifeq ($(GOOS),darwin)
CGO_LDFLAGS=/usr/lib/libsqlite3.0.dylib
else
CGO_CFLAGS=-I/usr/local/include
CGO_LDFLAGS=-L/usr/local/lib/oeis -L/usr/local/lib -lsqlite3
endif

include $(GOROOT)/src/Make.pkg
