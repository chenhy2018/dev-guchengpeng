package rtutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/qiniu/http/httputil.v1"
)

type RTOf570 struct {
	http.RoundTripper
}

func New570RT(rt http.RoundTripper) (nrt http.RoundTripper) {

	if rt == nil {
		rt = http.DefaultTransport
	}
	return RTOf570{rt}
}

var StatusCodeKeyWords map[int]string = map[int]string{
	499: "context canceled",
	570: "connection refused",
}

func (rt RTOf570) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	resp, err = rt.RoundTripper.RoundTrip(req)
	if err != nil {
		for code, words := range StatusCodeKeyWords {
			if strings.Contains(err.Error(), words) {
				b, _ := json.Marshal(map[string]string{"error": err.Error()})
				resp = &http.Response{
					StatusCode:    code,
					Header:        map[string][]string{"Content-Type": []string{"application/json"}},
					ContentLength: int64(len(b)),
					Body:          ioutil.NopCloser(bytes.NewReader(b)),
				}
				err = nil
				return
			}
		}
	}
	return
}

func (rt RTOf570) CancelRequest(req *http.Request) {

	if rc, ok := httputil.GetRequestCanceler(rt.RoundTripper); ok {
		rc.CancelRequest(req)
	}
}
