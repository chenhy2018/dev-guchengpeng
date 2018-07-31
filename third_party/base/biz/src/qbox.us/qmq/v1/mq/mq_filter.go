package mq

import (
	"net/http"

	"qbox.us/errors"
	"qbox.us/net/httputil"
	"qbox.us/servend/account"

	"github.com/qiniu/http/flag.v1"
	"github.com/qiniu/xlog.v1"
)

type filterArgs struct {
	MqName   string `flag:"_"`
	MqNameTo string `flag:"to"`
	ByUid    uint32 `flag:"by"`
}

//
// POST /admin-filter/<UID-MQID>/by/<ByUid>/to/<UID-MQID2>
//
func (r *Service) DoAdmin_filter_(w http.ResponseWriter, req *http.Request) {

	xl := xlog.New(w, req)

	err := account.CheckOperator(r.Account, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	var args filterArgs
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

	mqTo, err := r.adminGetMQ(args.MqNameTo)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	err = mq.FilterMsgs(func(msg []byte) bool {
		return filterMsg(xl, msg, args.ByUid, mqTo)
	})
	if err != nil {
		err = errors.Info(err, "filterMsgs from mq failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	httputil.ReplyWithCode(w, 200)
	return
}

func filterMsg(xl *xlog.Logger, msg []byte, byUid uint32, mqTo *Instance) (filtered bool) {

	uid, err := getMsgUid(xl, msg)
	if err != nil {
		return
	}

	if uint32(uid) != byUid {
		return
	}

	_, err = mqTo.Put(msg)
	if err != nil {
		xl.Warn("filterMsg: mqTo.Put failed -", err)
		return
	}

	return true
}
