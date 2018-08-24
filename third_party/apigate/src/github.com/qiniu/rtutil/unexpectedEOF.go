package rtutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type UEof struct {
	http.RoundTripper
}

func NewUnexpectedEOFRT(rt http.RoundTripper) (nrt http.RoundTripper) {
	return UEof{rt}
}

func (rt UEof) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	resp, err = rt.RoundTripper.RoundTrip(req)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected EOF") {
			b, _ := json.Marshal(map[string]string{"error": err.Error()})
			resp = &http.Response{
				StatusCode:    400,
				Header:        map[string][]string{"Content-Type": []string{"application/json"}},
				ContentLength: int64(len(b)),
				Body:          ioutil.NopCloser(bytes.NewReader(b)),
			}
			err = nil
		}
	}
	return
}
