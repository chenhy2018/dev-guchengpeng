package file

import (
	"io"
	"os"
)

const eor = '\n'

// ----------------------------------------------------------
// NOTE: 这个包已经迁移到 github.com/qiniu/osl/log

type Logger struct {
	io.Writer
}

func New(f *os.File) Logger {
	return Logger{f}
}

func (r Logger) Log(msg []byte) {
	msg = append(msg, eor)
	r.Write(msg)
}

// ----------------------------------------------------------

var Stdout = Logger{os.Stdout}
var Stderr = Logger{os.Stderr}

// ----------------------------------------------------------
