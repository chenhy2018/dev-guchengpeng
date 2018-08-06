package rewrite

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"qbox.us/api/qconf/appg"
	"qbox.us/api/qconf/bucketinfo.v2"
	"qbox.us/api/qconf/domaing"
)

var ErrRouterWithHost = errors.New("RouterWithHost Inconsist With Handler")

type Handler struct {
	Handle         func(string) string
	RouterWithHost bool //处理函数是否支持RouterWithHost
}

//添加新的原生处理函数后在此注册，配置文件中"native_handler"可配置项列表
var NativeHandler = map[string]Handler{
	"RemoveLeftSlash": {RemoveLeftSlash, false},
	"KODO4270":        {KODO4270, true},
}

/*仅保留一个URL开头斜杠*/
func RemoveLeftSlash(src string) (dest string) {
	return "/" + strings.TrimLeft(src, "/")
}

/*iQiyi KODO-4270*/
func KODO4270(src string) (dest string) {
	oSrc := src
	if QconfCli == nil {
		return oSrc
	}
	if !strings.HasPrefix(src, "http://") {
		src = "http://" + src
	}
	u, err := url.Parse(src)
	if err != nil {
		return oSrc
	}

	uPath := u.Path
	cond := (!strings.HasPrefix(uPath, "/videos/ots") &&
		!strings.HasPrefix(uPath, "/qpdxv/ots") &&
		!strings.HasPrefix(uPath, "/videos/vts/") &&
		!strings.HasPrefix(uPath, "/qpdxv/vts/")) &&
		(strings.HasSuffix(uPath, ".ts") ||
			strings.HasSuffix(uPath, ".m2ts"))
	if !cond {
		return oSrc
	}
	query := u.Query()
	startArg := query.Get("start")
	endArg := query.Get("end")
	if startArg == "" || endArg == "" {
		return oSrc
	}
	uPath = fmt.Sprintf("%s_%s_%s", uPath, startArg, endArg)
	ret, err := domaing.Client{QconfCli}.Get(nil, u.Host)
	if err != nil {
		return oSrc
	}
	bi, err := bucketinfo.Client{QconfCli}.GetBucketInfo(nil, ret.Uid, ret.Tbl)
	if err != nil {
		return oSrc
	}
	params := url.Values{}
	if bi.Source != "" {
		srcs := strings.Split(bi.Source, ";")
		for _, s := range srcs {
			if s != "" {
				m := strings.TrimRight(s, "/") + u.RequestURI()
				params.Add("qiniu_mirror", base64.URLEncoding.EncodeToString([]byte(m)))
			}
		}
	} else {
		return oSrc
	}
	if bi.Host != "" {
		params.Add("qiniu_mirror_host", base64.URLEncoding.EncodeToString([]byte(bi.Host)))
	}
	newUrl := "http://" + u.Host + uPath + "?" + params.Encode()
	ak, sk, err := appg.Client{QconfCli}.GetAkSk(nil, ret.Uid)
	if err != nil {
		return oSrc
	}
	return makePrivateUrl(newUrl, 0, ak, sk)
}
