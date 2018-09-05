package fopg

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"

	"github.com/qiniu/errors"
	"github.com/qiniu/xlog.v1"
	"qbox.us/dcutil"
)

const (
	headerLen  = 256
	gMinHeader = 6
)

type readCloser struct {
	io.Reader
	io.Closer
}

func hasPrefixArr(data string, arr []string) bool {
	for _, value := range arr {
		if strings.HasPrefix(data, value) {
			return true
		}
	}

	return false
}

func decodeMeta(xl *xlog.Logger, rr io.ReadCloser, l int) (r io.ReadCloser, start, length int64, metas dcutil.M, err error) {
	r = rr
	buf := make([]byte, headerLen)
	nr := 0
	if nr, err = io.ReadFull(r, buf); err != nil {
		if err != io.ErrUnexpectedEOF {
			return
		}
		buf = buf[:nr]
		err = nil
	}

	if buf[0] != 0 {
		if i := bytes.IndexByte(buf, 0); i > 0 {
			buf = buf[:i]
		}
		metas, err = dcutil.ParseQuery(string(buf))
		if err != nil {
			err = errors.Info(err, "CacheExt.Get").Detail(err)
			return
		}
		start = headerLen
		length = int64(l) - start
		return
	}

	// old version backward compatibility:
	hl := int(binary.LittleEndian.Uint32(buf[2:6]))

	rbuf := bytes.NewReader(buf[6:])
	r = readCloser{io.MultiReader(rbuf, r), r}

	buf = make([]byte, hl)
	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}

	metas = parseMetasOld(string(buf))
	start = int64(hl + gMinHeader)
	length = int64(l) - start
	return
}

func parseMetasOld(s string) dcutil.M {
	m := make(dcutil.M)
	words := strings.Split(s, "&")
	for _, word := range words {
		if strings.HasPrefix(word, "m=") {
			m["m"] = word[2:]
		}
	}
	return m
}
