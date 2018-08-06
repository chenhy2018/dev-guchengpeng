package mock

/*
 == fopd: 底层 fileop 服务 ==

 * 无用户、无授权概念
 * 不需要考虑 cache 支持
 * Input: Cmd, FileHandle (转 Source)、其他 FileOp 参数
 * Output: 标准的输出流 (need Content-Length)
 * 提供方式：web service

 请求包：/op/<Cmd>?fh=<FileHandle>&其他 FileOp 参数
*/

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"qbox.us/net/httputil"
	"strconv"
	"strings"
	"time"
)

type Fopd struct{}

func (s Fopd) op(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		httputil.ReplyError(w, "ioutil.ReadAll failed", 599)
		return
	}
	// mock execute long time op for lb1
	if strings.HasPrefix(req.URL.Query().Get("cmd"), "lb1") {
		rand.Seed(time.Now().UnixNano())
		msec := rand.Intn(3000)
		time.Sleep(time.Duration(msec) * time.Millisecond)
	}

	msgHeader := "mockfopd."
	w.Header().Set("Content-Length", strconv.Itoa(len(bs)+len(msgHeader)))
	w.Header().Set("Content-Type", req.Header.Get("Content-Type"))
	w.Header().Set("X-Qiniu-Fop-Stats", "HD:1500")
	w.WriteHeader(200)
	w.Write([]byte("mockfopd."))
	io.Copy(w, bytes.NewReader(bs))
}

func (s Fopd) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/op", func(w http.ResponseWriter, req *http.Request) {
		s.op(w, req)
	})
	return mux
}
