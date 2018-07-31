package qconfgetb

import (
	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/httputil.v1"
	"net/http"
)

// ------------------------------------------------------------------------

/*
POST /getb?id=<Id>

200 OK
Content-Type: application/bson
<BsonDoc>
*/
func Do(w http.ResponseWriter, req *http.Request, getter func(id string) (doc interface{}, err error)) {

	//TODO: auth

	id := req.FormValue("id")
	doc, err := getter(id)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	bsonrpc.Reply(w, 200, doc)
}

// ------------------------------------------------------------------------
