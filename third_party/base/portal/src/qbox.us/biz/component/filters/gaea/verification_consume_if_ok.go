package gaea

import (
	"net/http"
	"reflect"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/api"
	"qbox.us/biz/services.v2/gaea"
)

// Filter for consuming verification state on GAEA after a successful operation.
//
// This will consume the state only if controller returns a response with code equal to `api.OK`, which can be useful
// if you want user can retry without having to be verified agian.
//
// Usage:
//
//		Router(Filter(gaea.VerificationCheck), Filter(gaea.VerificationConsumeIfOK), Get(user.User{})),
//
// Note that `gaea.VerificationService` must be provided first.
//

func VerificationConsumeIfOK(types int) interface{} {

	return func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {
		ctx.Next()

		// Get controller response
		var res teapot.ActionOut
		err := ctx.Find(&res, "")
		if res == nil {
			log.Warn(err)
			return
		}

		out := res.Out()
		if len(out) == 0 {
			return
		}

		var body reflect.Value
		if out[len(out)-1].Kind() == reflect.Int {
			if len(out) == 1 {
				return
			}

			body = out[len(out)-2]
		} else {
			body = out[len(out)-1]
		}

		if !body.CanInterface() {
			return
		}

		itf := body.Interface()

		// Check response type and code
		if result, ok := itf.(*api.JsonResult); ok {
			if result.Code == nil || result.Code == api.OK {
				// Should consume verification
				var verification gaea.VerificationService
				err = ctx.Find(&verification, "")

				if err != nil {
					// Service dependency not found. This could be an implementation mistake.
					log.Warn(err)
					return
				}

				cookies := req.Cookies()
				err = verification.ConsumeWithCookie(cookies, types)
				if err != nil {
					log.Warn(err)
				}
			}
		}

		return
	}
}
