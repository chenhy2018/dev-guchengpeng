package iam

import (
	"net/http"

	"github.com/teapots/inject"
	"github.com/teapots/teapot"
	qiniumac "qiniu.com/auth/qiniumac.v1"

	"qbox.us/biz/component/sessions"
	"qbox.us/biz/services.v2/account"
	iamOAuth "qbox.us/biz/services.v2/iam/oauth"
	utils "qbox.us/biz/utils.v2"
	"qbox.us/iam/entity"
	"qbox.us/oauth"
)

// WrapUserInfo 会根据当前用户是否为 iam 用户决定返回的 UserInfo 代表着什么：
// 当用户是非 IAM 用户时，返回的 UserInfo 代表着当前用户;
// 当用户是 IAM 用户时，返回的 UserInfo 代表着其根用户.
//
// TODO: 这样设计的原因是希望尽可能少的对现有代码造成影响，未来应该在代码重构时对这部分也一同进行改造。
func WrapUserInfo(dep string) interface{} {
	return inject.Provide{
		inject.Dep{0: dep},
		func(info *account.UserInfo, iamUser *entity.User, ctx teapot.Context, l teapot.ReqLogger) *account.UserInfo {
			if iamUser != nil {
				info = &account.UserInfo{
					Uid: iamUser.RootUID,
				}

				var accService account.AdminAccountService
				ctx.Find(&accService, "")
				if accService != nil {
					uinfo, err := accService.FindInfoByUid(iamUser.RootUID)
					if err != nil {
						l.Warn(err)
					} else {
						info = &uinfo.UserInfo
					}
				}
			}

			return info
		},
	}
}

// WrapOAuth 会根据当前用户是否为`IAM 用户`决定返回何种实现的 `*oauth.Transport`:
// 当用户是非`IAM 用户`时, 使用默认的`auth`(用户身份是否合法的问题交给默认的`auth`来处理);
// 当用户是`IAM 用户`时,创建一个新的`auth`,这个`auth`已经不是真正意义上的`OAuthTransport`了,只是借助其类型来完成`qiniumac`的签名.
//
// TODO: 这样设计的原因是希望尽可能少的对现有代码造成影响，未来应该在代码重构时对这部分也一同进行改造。
func WrapOAuth(dep string) interface{} {
	return inject.Provide{
		inject.Dep{0: dep},
		func(auth *oauth.Transport, iamUser *entity.User, l teapot.ReqLogger, ctx teapot.Context, req *http.Request) *oauth.Transport {
			if iamUser == nil {
				return auth
			}

			return NewIAMOAuthWapper(iamUser, l, ctx, req)
		},
	}
}

func NewIAMOAuthWapper(iamUser *entity.User, l teapot.ReqLogger, ctx teapot.Context, req *http.Request) *oauth.Transport {
	var (
		keypair    *entity.Keypair
		resService iamOAuth.ResService
		transport  = &oauth.Transport{
			Config:    &oauth.Config{},
			Token:     &oauth.Token{},
			Transport: http.DefaultTransport,
		}
	)

	var sess sessions.SessionStore
	ctx.Find(&sess, "")
	if sess != nil {
		keypair = loadKeypairFromSession(sess)
	}

	if keypair == nil {
		if err := ctx.Find(&resService, ""); err != nil {
			l.Errorf("<NewIAMOAuthWapper> find iam resService failed: %s", err)
			return transport
		}

		keypairs, err := resService.Keypairs()
		if err != nil {
			l.Errorf("<IAMOemOAuthWrapper> ListUserKeypairs(rootUID: %d, alias: %s) failed: %s", iamUser.RootUID, iamUser.Alias, err)
			return transport
		}

		if len(keypairs) == 0 {
			l.Warnf("<IAMOemOAuthWrapper> iam user(rootUID: %d, alias: %s) have no ak/sk", iamUser.RootUID, iamUser.Alias)
			return transport
		}

		keypair = &keypairs[0]

		if sess != nil {
			saveKeypairToSession(sess, keypair)
		}
	}

	transport.Transport = NewRealInfoTransport(
		qiniumac.NewTransport(
			&qiniumac.Mac{
				AccessKey: keypair.AccessKey,
				SecretKey: []byte(keypair.SecretKey),
			},
			http.DefaultTransport,
		),
		&RealInfo{
			IP: utils.RealIp(req),
			UA: req.UserAgent(),
		},
	)
	return transport
}
