package gaea

import (
	"net/http"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/api"
	"qbox.us/biz/services.v2/gaea"
)

// Filter for checking verification state on GAEA.
//
// Usage:
//
//		Router(Filter(gaea.VerificationCheck), Get(user.User{})),
//
// For unverified state, will response with:
//
//		{ "code": 401, "message": "unauthorized" }
//
// Note that `gaea.VerificationService` must be provided first.
//
func VerificationCheck(types int) teapot.Handler {
	return func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {
		var (
			verification gaea.VerificationService
		)
		err := ctx.Find(&verification, "")
		cookies := req.Cookies()

		if err == nil {
			ok, err := verification.CheckWithCookie(cookies, types)

			if err == nil && ok {
				// Verification OK
				return
			} else {
				if err != nil {
					log.Warn(err)
				}
			}
		} else {
			// Service dependency not found. This could be an implementation mistake.
			log.Warn(err)
		}

		res := &api.JsonResult{
			Code: api.OpIsNotConfirmed,
		}
		res.Write(ctx, rw, req)
	}
}
