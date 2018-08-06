package mq

import (
	"net/http"

	"qbox.us/errors"
	"qbox.us/net/httputil"
	"qbox.us/servend/account"

	"github.com/qiniu/http/flag.v1"
)

type updateArgs struct {
	MqName  string `flag:"_"`
	Expires uint32 `flag:"expires"`
}

//
// POST /admin-update/<UID-MQID>/expires/<ExpiresInSeconds>
//
func (r *Service) DoAdmin_update_(w http.ResponseWriter, req *http.Request) {

	err := account.CheckOperator(r.Account, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	var args updateArgs
	err = flag.Parse(&args, req.URL.Path[1:])
	if err != nil {
		httputil.Error(w, err)
		return
	}

	err = r.updateMQ(args.MqName, args.Expires)
	if err != nil {
		err = errors.Info(err, "make failed").Detail(err)
		httputil.Error(w, err)
	}

	httputil.ReplyWithCode(w, 200)
	return
}

func (r *Service) updateMQ(name string, expires uint32) (err error) {

	err = checkValidMQ(name)
	if err != nil {
		return
	}

	if expires <= 0 || expires > 48*3600 { // 不能超过 48 小时
		err = errors.Info(errors.EINVAL, "invalid expires, must <= 48 * 3600")
		return
	}

	r.mutex.RLock()
	mq, ok := r.mqs[name]
	r.mutex.RUnlock()
	if !ok {
		err = errors.Info(errors.ENOENT, "no such mq", name)
		return
	}

	err = mq.UpdateIndexExpires(expires)
	if err != nil {
		return
	}
	return
}
