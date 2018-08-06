package audit

import (
	"io"
	"net/http"
	"qbox.us/cc"
	"github.com/qiniu/log.v1"
	"strconv"
)

// ------------------------------------------------------------------------------------------

func PostEx(
	r *http.Client, url string, bodyType string, body io.Reader, bodyLength int) (resp *http.Response, err error) {

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = int64(bodyLength)
	return r.Do(req)
}

// ------------------------------------------------------------------------------------------

type Logger struct {
	hosts []string
}

func New(hosts []string) *Logger {
	return &Logger{hosts}
}

func (r *Logger) Log(msg []byte) {

	cmd := "/put?len=" + strconv.Itoa(len(msg))
	for _, host := range r.hosts {
		url := host + cmd
		body := cc.NewBytesReader(msg)
		resp, err := PostEx(http.DefaultClient, url, "application/octet-stream", body, len(msg))
		if err == nil {
			resp.Body.Close()
			return
		}
	}
	log.Warn("audit.Logger.Log failed: service down")
}

// ------------------------------------------------------------------------------------------
