package fopd

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func TestUrlReader(t *testing.T) {

	ast := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("helloworld"))
	}))
	defer ts.Close()

	ur := NewUrlReader(xlog.NewDummy(), ts.URL)
	defer ur.Close()
	b, err := ioutil.ReadAll(ur)
	ast.Nil(err)
	ast.Equal(string(b), "helloworld")
}
