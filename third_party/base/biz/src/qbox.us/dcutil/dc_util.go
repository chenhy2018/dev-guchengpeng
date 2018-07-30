package dcutil

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api/dc"
	"qbox.us/errors"
)

// --------------------------------------------------------------------

type M map[string]string

type Interface interface {
	Set(xl *xlog.Logger, key []byte, r io.Reader, length int64, metas M) (err error)
	SetWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64, metas M) (host string, err error)
	Get(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, metas M, err error)
	RangeGet(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, metas M, err error)
	RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, r io.ReadCloser, length int64, metas M, err error)
}

func ParseQuery(query string) (metas M, err error) {

	metas = make(M)
	for query != "" {
		key := query
		if i := strings.Index(key, "&"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		value, err1 := url.QueryUnescape(value)
		if err1 != nil {
			err = err1
			continue
		}
		metas[key] = value
	}
	return
}

func (metas M) Encode() string {

	parts := make([]string, 0, len(metas))
	for k, v := range metas {
		parts = append(parts, k+"="+url.QueryEscape(v))
	}
	return strings.Join(parts, "&")
}

// --------------------------------------------------------------------

type Cache interface {
	Get(xl *xlog.Logger, key []byte) (rc io.ReadCloser, n int, err error)
	RangeGet(xl *xlog.Logger, key []byte, from, to int64) (rc io.ReadCloser, n int, err error)
	RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, rc io.ReadCloser, n int, err error)
	Put(xl *xlog.Logger, key []byte, r io.Reader, n int) error
	PutWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, n int) (string, error)
}

type CacheExt struct {
	dc Cache
}

func NewExt(dc Cache) CacheExt {
	return CacheExt{dc}
}

// --------------------------------------------------------------------

const (
	headerLen = 256
)

const (
	gVersion   = 0
	gMinHeader = 6
)

var ErrMetasTooLarge = errors.New("Metas too large")

func (s CacheExt) Set(xl *xlog.Logger, key []byte, r io.Reader, length int64, metas M) (err error) {

	if length < 0 {
		err = errors.Info(errors.EINVAL, "CacheExt.Set: invalid length")
		return
	}

	m := metas.Encode()
	if len(m) > headerLen {
		err = errors.Info(ErrMetasTooLarge, "CacheExt.Set", "metas", m).Warn()
		return
	}

	b := make([]byte, headerLen)
	copy(b, m)

	buf := bytes.NewReader(b)

	if rt, ok := r.(io.ReaderAt); ok {
		err = s.dc.Put(xl, key, MultiReaderAt(b, rt), int(length)+headerLen)
	} else {
		err = s.dc.Put(xl, key, io.MultiReader(buf, r), int(length)+headerLen)
	}
	return
}

func (s CacheExt) SetWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64, metas M) (host string, err error) {

	if length < 0 {
		err = errors.Info(errors.EINVAL, "CacheExt.SetWithHostRet: invalid length")
		return
	}

	m := metas.Encode()
	if len(m) > headerLen {
		err = errors.Info(ErrMetasTooLarge, "CacheExt.SetWithHostRet", "metas", m).Warn()
		return
	}

	b := make([]byte, headerLen)
	copy(b, m)

	buf := bytes.NewReader(b)

	if rt, ok := r.(io.ReaderAt); ok {
		host, err = s.dc.PutWithHostRet(xl, key, MultiReaderAt(b, rt), int(length)+headerLen)
	} else {
		host, err = s.dc.PutWithHostRet(xl, key, io.MultiReader(buf, r), int(length)+headerLen)
	}
	return
}

type readCloser struct {
	io.Reader
	io.Closer
}

func (s CacheExt) get(xl *xlog.Logger, key []byte) (r io.ReadCloser, start, length int64, metas M, err error) {

	r, l, err := s.dc.Get(xl, key)
	if err != nil {
		return
	}
	return decodeMeta(xl, r, l)
}

func decodeMeta(xl *xlog.Logger, rr io.ReadCloser, l int) (r io.ReadCloser, start, length int64, metas M, err error) {
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
		metas, err = ParseQuery(string(buf))
		if err != nil {
			err = errors.Info(err, "CacheExt.Get").Detail(err)
			return
		}
		start = headerLen
		length = int64(l) - start
		return
	}

	//
	// old version backward compatibility:
	//

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

func (s CacheExt) Get(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, metas M, err error) {

	r, _, length, metas, err = s.get(xl, key)
	return
}

func (s CacheExt) RangeGet(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, metas M, err error) {

	gr, start, _, metas, err := s.get(xl, key)
	if err != nil {
		return
	}
	gr.Close()

	r, l, err := s.dc.RangeGet(xl, key, from+start, to+start)
	length = int64(l)
	return
}

func (s CacheExt) RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, r io.ReadCloser, length int64, metas M, err error) {

	gr, start, _, metas, err := s.get(xl, key)
	if err != nil {
		return
	}
	gr.Close()

	host, r, l, err := s.dc.RangeGetAndHost(xl, key, from+start, to+start)
	length = int64(l)
	return
}

func parseMetasOld(s string) M {

	m := make(M)
	words := strings.Split(s, "&")
	for _, word := range words {
		if strings.HasPrefix(word, "m=") {
			m["m"] = word[2:]
		}
	}
	return m
}

// --------------------------------------------------------------------
// Get With Specific Host

type hostGetter interface {
	GetWithHost(xl *xlog.Logger, host string, key []byte) (r io.ReadCloser, length int64, err error)
	RangeGetWithHost(xl *xlog.Logger, host string, key []byte, from int64, to int64) (r io.ReadCloser, length int64, err error)
}

var defaultTransport = rpc.NewTransportTimeoutWithConnsPool(5*time.Second, 3*time.Second, 5)

var defaultConn hostGetter = dc.NewConn("", defaultTransport)

func InitDefaultConnWithTimeout(dialMs, respMs int) {

	defaultConn = dc.NewConnWithTimeout("", &dc.TimeoutOptions{
		DialMs: dialMs,
		RespMs: respMs,
	})
}

func getWithHost(xl *xlog.Logger, host string, key []byte) (r io.ReadCloser, start, length int64, metas M, err error) {

	r, l, err := defaultConn.GetWithHost(xl, host, key)
	if err != nil {
		return
	}
	return decodeMeta(xl, r, int(l))
}

func GetWithHost(xl *xlog.Logger, host string, key []byte) (r io.ReadCloser, length int64, metas M, err error) {

	r, _, length, metas, err = getWithHost(xl, host, key)
	return
}

func RangeGetWithHost(xl *xlog.Logger, host string, key []byte, from, to int64) (r io.ReadCloser, length int64, metas M, err error) {

	gr, start, _, metas, err := getWithHost(xl, host, key)
	if err != nil {
		return
	}
	gr.Close()

	r, l, err := defaultConn.RangeGetWithHost(xl, host, key, from+start, to+start)
	length = int64(l)
	return
}

// --------------------------------------------------------------------
