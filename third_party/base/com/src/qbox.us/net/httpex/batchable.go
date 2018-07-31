package httpex

import (
	"bytes"
	"fmt"
	"net/http"
	"qbox.us/net/httputil"
	"reflect"
)

// ---------------------------------------------------------------------------

type batchResponse struct {
	header       http.Header
	buf          *bytes.Buffer
	Code         int
	fwriteHeader bool
	fnext        bool
}

var startTag = []byte{'['}

func newBatchResponse() *batchResponse {

	return &batchResponse{
		header: make(http.Header),
		buf:    bytes.NewBuffer(startTag),
		Code:   200,
	}
}

func (p *batchResponse) Next() {

	if p.fnext {
		p.buf.WriteByte(',')
	} else {
		p.fnext = true
	}
	p.buf.WriteByte('}')
	p.fwriteHeader = false
}

func (p *batchResponse) Done() []byte {

	p.buf.WriteByte(']')
	return p.buf.Bytes()
}

func (p *batchResponse) Header() http.Header {

	return p.header
}

func (p *batchResponse) Write(buf []byte) (int, error) {

	if !p.fwriteHeader {
		p.WriteHeader(200)
	}
	return p.buf.Write(buf)
}

func (p *batchResponse) WriteHeader(code int) {

	p.fwriteHeader = true
	fmt.Fprintf(p.buf, "{code:%d,data:", code)
	if code/100 != 2 {
		p.Code = 298
	}
}

// ---------------------------------------------------------------------------

func (mux *ServeMux) Call(rcvr reflect.Value, w http.ResponseWriter, req *http.Request, env interface{}) {

	if !mux.Dispatch(rcvr, w, req, env) {
		httputil.ReplyError(w, "Bad method", 400)
	}
}

func (mux *ServeMux) CallIntf(rcvr interface{}, w http.ResponseWriter, req *http.Request, env interface{}) {

	rcvr1 := reflect.ValueOf(rcvr)
	mux.Call(rcvr1, w, req, env)
}

func (mux *ServeMux) Batch(rcvr reflect.Value, w http.ResponseWriter, req *http.Request, env interface{}) {

	err := req.ParseForm()
	if err != nil {
		httputil.Error(w, err)
		return
	}

	bw := newBatchResponse()
	if ops, ok := req.Form["op"]; ok {
		for _, op := range ops {
			req.URL.Path = op
			mux.Call(rcvr, bw, req, env)
			bw.Next()
		}
	}
	data := bw.Done()

	httputil.ReplyWith(w, bw.Code, "application/json", data)
}

// ---------------------------------------------------------------------------
