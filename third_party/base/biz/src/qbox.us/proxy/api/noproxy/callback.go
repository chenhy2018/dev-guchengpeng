package noproxy

import (
	"net/http"
	"strings"
	"time"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api"
	"qbox.us/proxy/api/proto"
	"qbox.us/servend/account"
)

type callbackInstance struct {
	tr         http.RoundTripper
	retryTimes int
	acc        account.InterfaceEx
}

func NewCallbackInstance(blockIPs []string, timeout time.Duration, retryTimes int, acc account.InterfaceEx) proto.CallbackProxy {

	return &callbackInstance{
		tr:         blockIPTransport(blockIPs, timeout, proto.CallbackUserAgent),
		retryTimes: retryTimes,
		acc:        acc,
	}
}

func (self *callbackInstance) Callback(l rpc.Logger,
	URLs []string, host, bodyType string, body string,
	accessKey string, config proto.CallbackConfig) (resp *http.Response, err error) {

	xl := xlog.NewWith(l.ReqId())

	var mac *digest.Mac
	if accessKey != "" {
		secretKey, ok := self.acc.GetSecret(accessKey)
		if !ok {
			err = errors.Info(api.EBadToken, "invalid accessKey in callback digital sign")
			return
		}
		mac = &digest.Mac{accessKey, secretKey}
	}

	tryTimes := self.retryTimes + 1
	if tryTimes < len(URLs) {
		tryTimes = len(URLs)
	}
	for i := 0; i < tryTimes; i++ {
		URL := URLs[i%len(URLs)]
		code := 0
		resp, err = self.httpPost(mac, xl, URL, host, bodyType, body, config)
		if err == nil {
			if resp.StatusCode/100 != 5 {
				break
			}
			if i < tryTimes-1 {
				resp.Body.Close()
			}
			code = resp.StatusCode
		}
		xl.Warnf("httpPosts: httpGet url:%v host:%v err:%v code:%v", URL, host, err, code)
	}
	return
}

func (self *callbackInstance) httpPost(mac *digest.Mac, xl *xlog.Logger,
	URL, host, bodyType, body string, config proto.CallbackConfig) (resp *http.Response, err error) {

	xl.Info("httpPost: url and host", URL, host)
	tr := self.tr
	if mac != nil {
		tr = digest.NewTransport(mac, self.tr)
	}
	client := rpc.Client{&http.Client{Transport: tr, Timeout: config.Timeout}}
	req, err := http.NewRequest("POST", URL, strings.NewReader(body))
	if err != nil {
		err = httputil.NewError(api.NetworkError, "callback service unavailable")
		return
	}
	if host != "" {
		req.Host = host
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = int64(len(body))
	resp, err = client.Do(xl, req)
	if err != nil {
		err = errors.Info(err, "httpPost: PostWith").Detail(err)
	}
	return
}
