package iam

import (
	"net/http"
	"strings"

	"github.com/teapots/inject"
	"github.com/teapots/teapot"

	"qbox.us/biz/component/client"
	"qbox.us/biz/component/sessions"
	"qbox.us/biz/services.v2/account"
	"qbox.us/iam/entity"
	"qbox.us/oauth"

	"qbox.us/biz/services.v2/iam"
	iamOAuth "qbox.us/biz/services.v2/iam/oauth"
	"qbox.us/biz/utils.v2/log"
)

const (
	LoginToken   = "iam_login_token"
	LoginRefresh = "iam_login_refresh"
	LoginExpired = "iam_login_expired"
	LoginRootUID = "iam_login_root_uid"
	LoginAlias   = "iam_login_alias"
)

func newOAuthTransport(host string, tr http.RoundTripper) *oauth.Transport {
	scopes := []string{
		iamOAuth.ScopeUserProfile.String(),
		iamOAuth.ScopeUserKeypairs.String(),
	}
	return &oauth.Transport{
		Config: &oauth.Config{
			Scope:    strings.Join(scopes, ","),
			TokenURL: host + "/iam/oauth2/token",
		},
		Transport: tr,
	}
}

func Service(hosts []string, dep string) interface{} {
	return inject.Provide{
		inject.Dep{0: dep},
		func(adminOAuth *oauth.Transport) iam.Service {
			return iam.NewService(hosts, adminOAuth)
		},
	}
}

func EmptyUserOAuth(host string) interface{} {
	return func(tr *client.TransportWithReqLogger) *iamOAuth.Transport {
		return iamOAuth.NewTransport(newOAuthTransport(host, tr))
	}
}

func UserOAuth(host string) interface{} {
	return func(ctx teapot.Context, tr *client.TransportWithReqLogger) *iamOAuth.Transport {
		var (
			accessToken string
			sess        sessions.SessionStore
			userOAuth   = iamOAuth.NewTransport(newOAuthTransport(host, tr))
		)

		ctx.Find(&sess, "")

		if sess != nil {
			accessToken = sess.Get(LoginToken).String()
			refreshToken := sess.Get(LoginRefresh).String()
			tokenExpiry := sess.Get(LoginExpired).MustInt64()

			if refreshToken != "" && isTokenExpired(tokenExpiry) {
				token, code, err := userOAuth.ExchangeByRefreshToken(refreshToken)

				var log teapot.Logger
				ctx.Find(&log, "")

				if code != http.StatusOK || err != nil {
					log.Warn("user refresh token:", code, err)
				} else {
					accessToken = token.AccessToken

					sess.Set(LoginToken, token.AccessToken)
					sess.Set(LoginRefresh, token.RefreshToken)
					sess.Set(LoginExpired, token.TokenExpiry)
				}
			}
		}

		userOAuth.Token = &oauth.Token{AccessToken: accessToken}

		return userOAuth
	}
}

func ResService(host string) interface{} {
	return func(tr *iamOAuth.Transport, l teapot.ReqLogger) iamOAuth.ResService {
		return iamOAuth.NewResService(host, tr, log.NewRpcWrapper(l))
	}
}

func UserInfo() interface{} {
	return func(ctx teapot.Context, sess sessions.SessionStore, l teapot.ReqLogger) *entity.User {
		user := loadUserFromSession(sess)
		if user != nil {
			return user
		}

		var s iamOAuth.ResService
		err := ctx.Find(&s, "")
		if err != nil {
			l.Warn("get iam resService failed:", err)
			return nil
		}

		user, err = s.Profile()
		if err != nil {
			l.Warn("get current iam user info failed:", err)
			return nil
		}

		saveUserToSession(sess, user)
		return user
	}
}

func RootUserInfo() interface{} {
	return func(ctx teapot.Context, info *account.UserInfo, l teapot.ReqLogger) *entity.RootUser {
		if info == nil {
			return nil
		}

		var s iam.Service
		ctx.Find(&s, "")
		user, err := s.GetRootUser(log.NewRpcWrapper(l), info.Uid)
		if err != nil {
			l.Warn("get iam root user info failed:", err)
			return nil
		}
		return user
	}
}

func IsLoginedIAMUser(sess sessions.SessionStore) bool {
	if sess != nil {
		return sess.Get(LoginAlias).MustString() != ""
	}
	return false
}
