package mq

import (
	"net/http"

	"qbox.us/errors"
	"qbox.us/net/httputil"
	"qbox.us/servend/account"

	"github.com/qiniu/http/flag.v1"
)

type makeArgs struct {
	MqName  string `flag:"_"`
	Expires uint32 `flag:"expires"`
}

//
// POST /admin-make/<UID-MQID>/expires/<ExpiresInSeconds>
//
func (r *Service) DoAdmin_make_(w http.ResponseWriter, req *http.Request) {

	err := account.CheckOperator(r.Account, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	var args makeArgs
	err = flag.Parse(&args, req.URL.Path[1:])
	if err != nil {
		httputil.Error(w, err)
		return
	}

	err = r.makeMQ(args.MqName, args.Expires)
	if err != nil {
		err = errors.Info(err, "make failed").Detail(err)
		httputil.Error(w, err)
	}

	httputil.ReplyWithCode(w, 200)
	return
}

func (r *Service) makeMQ(name string, expires uint32) (err error) {

	err = checkValidMQ(name)
	if err != nil {
		return
	}

	if expires <= 0 || expires > 48*3600 { // 不能超过 48 小时
		err = errors.Info(errors.EINVAL, "invalid expires, must <= 48 * 3600")
		return
	}

	r.mutex.RLock()
	_, ok := r.mqs[name]
	r.mutex.RUnlock()
	if ok {
		err = errors.Info(errors.EEXIST, "mq exists")
		return
	}

	mq, err := OpenInstance(r.DataPath+name, r.ChunkBits, expires)
	if err != nil {
		err = errors.Info(err, "open mq failed", name)
		return
	}

	r.mutex.Lock()
	_, ok = r.mqs[name]
	if !ok {
		r.mqs[name] = mq
	}
	r.mutex.Unlock()

	if ok {
		go mq.Close()
	}
	return
}
