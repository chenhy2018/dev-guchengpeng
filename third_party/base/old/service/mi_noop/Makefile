# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=mi_noop
GOFILES=\
	mi_noop.go\

include $(GOROOT)/src/Make.cmd

.PHONY: gofmt
gofmt:
	gofmt -w *.go
