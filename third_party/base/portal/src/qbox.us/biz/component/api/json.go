package api

import (
	"encoding/json"
	"net/http"

	"github.com/teapots/teapot"
)

var JsonContentType = "application/json; charset=UTF-8"

type JsonResult struct {
	Code    Code                   `json:"code"`
	Data    interface{}            `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

type _jsonResult struct {
	Code    Code                   `json:"code"`
	Data    interface{}            `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

func NewJsonResult() *JsonResult {
	return &JsonResult{
		Extra: map[string]interface{}{},
	}
}

func (r *JsonResult) MarshalJSON() ([]byte, error) {
	if r.Code == nil || r.Code.Code() == 0 {
		r.Code = OK
	}

	if r.Message == "" && r.Code != OK {
		r.Message = r.Code.Humanize()
	}

	// 这里直接使用 JsonResult 会循环自己，导致栈溢出
	v := &_jsonResult{
		Code:    r.Code,
		Message: r.Message,
		Data:    r.Data,
		Extra:   r.Extra,
	}

	return json.Marshal(v)
}

func (r *JsonResult) Write(ctx teapot.Context, rw http.ResponseWriter, req *http.Request) {
	if r.Code == nil {
		r.Code = OK
	}

	var res = map[string]interface{}{
		"code": r.Code.Code(),
	}

	if r.Data != nil {
		res["data"] = r.Data
	}

	if r.Message == "" && r.Code != OK {
		res["message"] = r.Code.Humanize()
	}

	for k, v := range r.Extra {
		res[k] = v
	}

	config := new(teapot.Config)
	// use struct, so u can just skip error
	ctx.Find(&config, "")

	var body []byte
	var err error
	if config.RunMode.IsDev() {
		body, err = json.MarshalIndent(res, "", "  ")
	} else {
		body, err = json.Marshal(res)
	}

	rw.Header().Set("Content-Type", JsonContentType)

	if err == nil {
		_, err = rw.Write(body)
	} else {
		errRes := &map[string]interface{}{
			"code":    ResultError.Code(),
			"message": ResultError.Humanize(),
		}
		if config.RunMode.IsDev() {
			body, _ = json.MarshalIndent(errRes, "", "  ")
		} else {
			body, _ = json.Marshal(errRes)
		}
		rw.Write(body)
	}

	if err != nil {
		var logger teapot.Logger
		ctx.Find(&logger, "")
		logger.Error("JsonResult.Write", err)
	}
}
