package access

import (
	"net/http"
	"time"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/api"
	"qbox.us/biz/component/sessions"
)

const (
	SessionOpConfirmKey = "op_confirmed"
)

func OpConfirm(ctx teapot.Context, rw http.ResponseWriter,
	req *http.Request, session sessions.SessionStore,
	log teapot.ReqLogger) {

	value := session.Get(SessionOpConfirmKey)

	expireAt := value.MustInt64()
	if time.Now().Unix() < expireAt {
		session.Set(SessionOpConfirmKey, 0)
		return
	}

	log.Info("get op confirm expire at failed: ")

	res := &api.JsonResult{
		Code: api.OpIsNotConfirmed,
	}

	res.Write(ctx, rw, req)
}
