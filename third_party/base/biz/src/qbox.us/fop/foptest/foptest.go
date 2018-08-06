package foptest

import (
	"github.com/qiniu/ts"
	"io"
	"net/http/httptest"
	"os"
	"qbox.us/fop"
	"testing"
)

// --------------------------------------------------------------------

type Recoder *httptest.ResponseRecorder

func Call(op func(fop.Writer, fop.Reader), r fop.Reader) Recoder {

	w := httptest.NewRecorder()
	op(w, r)
	return w
}

func CallEx(op func(fop.Writer, fop.Reader), r fop.Reader, file string, t *testing.T) Recoder {

	w := Call(op, r)

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		ts.Fatal(t, "foptest.CallEx", "open file failed", file, err)
		return nil
	}
	defer f.Close()

	_, err = io.Copy(f, w.Body)
	if err != nil {
		ts.Fatal(t, "foptest.CallEx", "io.Copy failed", err)
		return nil
	}
	return w
}

// --------------------------------------------------------------------
