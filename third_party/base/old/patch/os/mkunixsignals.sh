#!/usr/bin/env bash
# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

echo '// ./mkunix.sh' "$1"
echo '// MACHINE GENERATED BY THE COMMAND ABOVE; DO NOT EDIT'
echo

cat <<EOH
package os

import (
  "syscall"
)

var _ = syscall.Syscall  // in case there are zero signals

const (
EOH

sed -n 's/^[ 	]*\(SIG[A-Z0-9][A-Z0-9]*\)[ 	].*/  \1 = UnixSignal(syscall.\1)/p' "$1"

echo ")"
