package gobrpc

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/http"
	"qbox.us/api"
	"qbox.us/errors"
	"github.com/qiniu/log.v1"
)

// ---------------------------------------------------------------------------

type errorRet struct {
	Error string
}

func callRet(ret interface{}, resp *http.Response) (err error) {

	defer resp.Body.Close()

	code := resp.StatusCode
	if code/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = gob.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				err = errors.Info(api.EUnexpectedResponse, "gobrpc.callRet").Detail(err)
			}
		}
	} else {
		if resp.ContentLength != 0 {
			var ret1 errorRet
			gob.NewDecoder(resp.Body).Decode(&ret1)
			if ret1.Error != "" {
				err = errors.Info(errors.New(ret1.Error), "gobrpc.callRet")
				return
			}
		}
		err = errors.Info(api.NewError(code), "gobrpc.callRet")
	}
	return
}

func PostEx(url string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		err = errors.Info(api.ENetworkError, "gobrpc.PostEx").Detail(err)
		return
	}

	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = bodyLength
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Info(api.ENetworkError, "gobrpc.PostEx").Detail(err)
	}
	return
}

// ---------------------------------------------------------------------------

type Client struct {
	BaseUrl string
}

func NewClient(baseUrl string) Client {
	return Client{baseUrl}
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client Client) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {

	log.Debug("gobrpc.Call:", serviceMethod, args)

	b := bytes.NewBuffer(nil)
	err = gob.NewEncoder(b).Encode(args)
	if err != nil {
		err = errors.Info(errors.EINVAL, "gobrpc.Call", serviceMethod, args).Detail(err).Warn()
		return
	}

	resp, err := PostEx(client.BaseUrl+serviceMethod, "application/octet-stream", b, int64(b.Len()))
	if err != nil {
		err = errors.Info(err, "gobrpc.Call", serviceMethod, args).Detail(err).Warn()
		return
	}

	err = callRet(reply, resp)
	if err != nil {
		err = errors.Info(err, "gobrpc.Call", serviceMethod, args).Detail(err).Warn()
		return
	}

	log.Debug("gobrpc.Call:", serviceMethod, args, reply)
	return
}

// ---------------------------------------------------------------------------
