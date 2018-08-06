package gaea

import (
	"net/http"

	"github.com/teapots/teapot"

	"qbox.us/biz/services.v2/gaea"
)

// Filter for consuming verification state on GAEA.
//
// This will always consume the state no matter what response is. If you only want state consumed after a successful
// operation, consider `VerificationConsumeIfOK` filter instead.
//
// Usage:
//
//		Router(Filter(gaea.VerificationCheck), Filter(gaea.VerificationConsumeAlways), Get(user.User{})),
//
// Note that `gaea.VerificationService` must be provided first.
//

func VerificationConsumeAlways(types int) interface{} {

	return func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {
		var verification gaea.VerificationService
		err := ctx.Find(&verification, "")
		cookies := req.Cookies()

		if err == nil {
			err := verification.ConsumeWithCookie(cookies, types)
			if err != nil {
				log.Warn(err)
			}
		} else {
			// Service dependency not found. This could be an implementation mistake.
			log.Warn(err)
		}

		return
	}
}
