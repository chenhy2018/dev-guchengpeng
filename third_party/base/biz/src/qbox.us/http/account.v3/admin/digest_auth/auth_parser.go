package digest_auth

import (
	"errors"

	sacc "qbox.us/servend/account"
	digest_auth "qbox.us/servend/digest_auth/admin"
)

// ---------------------------------------------------------------------------

type Config digest_auth.Config

func New(cfg *Config) sacc.InterfaceEx {

	acc1 := deprecatedAccount{}
	return digest_auth.New((*digest_auth.Config)(cfg), acc1)
}

func NewParser(cfg *Config) sacc.OldParserEx {

	return sacc.OldParserEx{New(cfg)}
}

// ---------------------------------------------------------------------------

var ErrDeprecatedToken = errors.New("deprecated token")

type deprecatedAccount struct {
}

func (acc deprecatedAccount) ParseAccessToken(token string) (user sacc.UserInfo, err error) {

	err = ErrDeprecatedToken
	return
}

func (acc deprecatedAccount) MakeAccessToken(user sacc.UserInfo) string {

	panic("MakeAccessToken is deprecated")
}
