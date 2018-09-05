// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc

import (
	"encoding/json"
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

	return json.NewDecoder(r).Decode(args)
}

func (c serverCodec) WriteResponse(w http.ResponseWriter, body interface{}, err error) {

	var msg []byte
	var code int

	if err != nil {
		code = api.HttpCode(err)
		msg, _ = json.Marshal(errorRet{err.Error()})
	} else {
		code = 200
		msg, _ = json.Marshal(body)
	}

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
}

// ---------------------------------------------------------------------------

func NewServer() *rpc.Server {
	return rpc.NewServer(ServerCodec)
}

// ---------------------------------------------------------------------------

var ServerCodec serverCodec
var DefaultServer = rpc.NewServer(ServerCodec)

// ---------------------------------------------------------------------------
