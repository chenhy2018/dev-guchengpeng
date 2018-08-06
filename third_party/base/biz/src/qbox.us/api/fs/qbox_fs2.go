package fs

import (
	"io"
	"net/http"
	"qbox.us/rpc"
	"strconv"
)

// ----------------------------------------------------------

type Service2 struct {
	Service
	IOHost string
}

// ----------------------------------------------------------

func New2(fsHost string, ioHost string, t http.RoundTripper) *Service2 {
	client := &http.Client{Transport: t}
	return &Service2{Service: Service{fsHost, rpc.Client{client}}, IOHost: ioHost}
}

// ----------------------------------------------------------

func (p *Service2) Mkfile(remoteFile string, ctxes io.Reader, meta *FileMeta) (ret PutRet, code int, err error) {
	url := p.IOHost + "/mkfile/1/fs-put/" + rpc.EncodeURI(remoteFile)
	code, err = p.Conn.CallWithBinary1(&ret, meta.MakeURL(url), ctxes)
	return
}

// ----------------------------------------------------------

type FileMeta struct {
	EditTime int64
	Alt      string
	Base     string
	Perm     uint32
}

func (meta *FileMeta) MakeURL(callback string) (url string) {

	if meta.EditTime != 0 {
		callback += "/editTime/" + strconv.FormatInt(meta.EditTime, 10)
	}
	if meta.Alt != "" {
		callback += "/alt/" + rpc.EncodeURI(meta.Alt)
	}
	if meta.Base != "" {
		callback += "/base/" + meta.Base
	}
	if meta.Perm != 0 {
		callback += "/perm/" + strconv.Itoa(int(meta.Perm))
	}
	return callback
}

// ----------------------------------------------------------
