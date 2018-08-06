package jsonrpc

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"qbox.us/api"
	"qbox.us/errors"
	"strings"
)

// ---------------------------------------------------------------------------

type errorRet struct {
	Error string `json:"error"`
}

func callRet(ret interface{}, resp *http.Response) (code int, err error) {

	defer resp.Body.Close()

	code = resp.StatusCode
	if code/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				err = errors.Info(api.EUnexpectedResponse, "jsonrpc.callRet").Detail(err)
				code = api.UnexpectedResponse
			}
		}
	} else {
		if resp.ContentLength != 0 {
			if ct, ok := resp.Header["Content-Type"]; ok && ct[0] == "application/json" {
				var ret1 errorRet
				json.NewDecoder(resp.Body).Decode(&ret1)
				if ret1.Error != "" {
					err = errors.Info(errors.New(ret1.Error), "jsonrpc.callRet")
					return
				}
			}
		}
		err = errors.Info(api.NewError(code), "jsonrpc.callRet")
	}
	return
}

func PostEx(url_ string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {

	req, err := http.NewRequest("POST", url_, body)
	if err != nil {
		err = errors.Info(api.ENetworkError, "PostEx").Detail(err)
		return
	}

	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = bodyLength
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Info(api.ENetworkError, "PostEx").Detail(err)
	}
	return
}

func CallWithBinary(ret interface{}, url_ string, body io.Reader, bodyLength int64) (code int, err error) {

	resp, err := PostEx(url_, "application/octet-stream", body, bodyLength)
	if err != nil {
		code = api.HttpCode(err)
		err = errors.Info(err, "jsonrpc.CallWithBinary", url_).Detail(err)
		return
	}
	code, err = callRet(ret, resp)
	if err != nil {
		err = errors.Info(err, "jsonrpc.CallWithBinary", url_).Detail(err)
	}
	return
}

func CallWithForm(ret interface{}, url_ string, params url.Values) (code int, err error) {

	msg := params.Encode()
	resp, err := PostEx(url_, "application/x-www-form-urlencoded", strings.NewReader(msg), int64(len(msg)))
	if err != nil {
		code = api.HttpCode(err)
		err = errors.Info(err, "jsonrpc.CallWithForm", url_, params).Detail(err)
		return
	}
	code, err = callRet(ret, resp)
	if err != nil {
		err = errors.Info(err, "jsonrpc.CallWithForm", url_, params).Detail(err)
	}
	return
}

// ---------------------------------------------------------------------------
