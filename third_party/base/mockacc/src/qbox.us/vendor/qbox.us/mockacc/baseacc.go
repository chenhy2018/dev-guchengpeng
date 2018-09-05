package mockacc

import (
	"fmt"
	"qbox.us/servend/account"
	"qbox.us/servend/oauth"
)

// ---------------------------------------------------------------------------------------

type Base struct{}

func (a Base) ParseAccessToken(token string) (user account.UserInfo, err error) {
	_, err = fmt.Sscanf(token, "%d%d%d%d%d", &user.Uid, &user.Utype, &user.Sudoer, &user.UtypeSu, &user.Appid)
	return
}

func (a Base) MakeAccessToken(user account.UserInfo) string {
	return fmt.Sprintf("%d %d %d %d %d", user.Uid, user.Utype, user.Sudoer, user.UtypeSu, user.Appid)
}

func MakeTransport(uid, utype uint32) *oauth.Transport {
	token := Base{}.MakeAccessToken(account.UserInfo{
		Uid:   uid,
		Utype: utype,
	})
	return oauth.NewTransport(token, nil)
}

// ---------------------------------------------------------------------------------------
