package mq

import (
	"net/http"
	"net/url"
	"strconv"

	"qbox.us/errors"
	"qbox.us/net/httputil"
	"qbox.us/servend/account"

	"github.com/qiniu/http/flag.v1"
	"github.com/qiniu/xlog.v1"
)

type nameArg struct {
	MqName string `flag:"_"`
}

//
// POST /admin-stat/<UID-MQID>
//
func (r *Service) DoAdmin_stat_(w http.ResponseWriter, req *http.Request) {

	xl := xlog.New(w, req)

	err := account.CheckOperator(r.Account, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	var args nameArg
	err = flag.Parse(&args, req.URL.Path[1:])
	if err != nil {
		httputil.Error(w, err)
		return
	}

	mq, err := r.adminGetMQ(args.MqName)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	ret := make(map[string]int)
	err = mq.FilterMsgs(func(msg []byte) bool {
		statMsg(xl, msg, ret)
		return false
	})
	if err != nil {
		err = errors.Info(err, "filterMsgs from mq failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	httputil.Reply(w, 200, ret)
	return
}

func statMsg(xl *xlog.Logger, msg []byte, ret map[string]int) {

	uid, err := getMsgUid(xl, msg)
	if err != nil {
		return
	}
	uidStr := strconv.Itoa(int(uid))

	ret[uidStr]++
}

func getMsgUid(xl *xlog.Logger, msg []byte) (uidRet uint32, err error) {

	params, err := url.ParseQuery(string(msg))
	if err != nil {
		xl.Warn("parseMsg: decode msg failed", msg, err)
		return
	}
	xl.Debugf("msg: %#v\n", params)

	uid, err := strconv.ParseUint(params.Get("owner"), 10, 32)
	if err != nil {
		xl.Warn("parseMsg: invalid uid", err)
		return
	}
	return uint32(uid), nil
}
