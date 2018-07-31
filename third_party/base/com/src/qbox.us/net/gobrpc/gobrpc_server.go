// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gobrpc

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/http"
	"qbox.us/api"
	"qbox.us/net/rpc"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------

type serverCodec struct{}

func (c serverCodec) ReadRequestHeader(req *http.Request) (method string, err error) {

	rawUrl := req.URL.Path
	pos := strings.LastIndex(rawUrl, "/")

	return rawUrl[pos+1:], nil
}

func (c serverCodec) ReadRequestBody(r io.Reader, args interface{}) (err error) {

	return gob.NewDecoder(r).Decode(args)
}

func (c serverCodec) WriteResponse(w http.ResponseWriter, body interface{}, err error) {

	var code int

	b := bytes.NewBuffer(nil)
	if err != nil {
		code = api.HttpCode(err)
		gob.NewEncoder(b).Encode(errorRet{err.Error()})
	} else {
		code = 200
		gob.NewEncoder(b).Encode(body)
	}

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(b.Len()))
	h.Set("Content-Type", "application/octet-stream")
	w.WriteHeader(code)
	w.Write(b.Bytes())
}

// ---------------------------------------------------------------------------

func NewServer() *rpc.Server {
	return rpc.NewServer(ServerCodec)
}

// ---------------------------------------------------------------------------

var ServerCodec serverCodec
var DefaultServer = rpc.NewServer(ServerCodec)

// ---------------------------------------------------------------------------
