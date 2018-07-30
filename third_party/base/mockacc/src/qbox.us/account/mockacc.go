package account

import (
	"errors"

	"github.com/qiniu/log.v1"

	"labix.org/v2/mgo"
	"qbox.us/mockacc"

	alogh "qbox.us/audit/logh"
	accs "qbox.us/account-api"
)

var (
	ErrTokenExpired = errors.New("token expired")
)

const (
	ACCESS_TOKEN_EXPIRES = 1800
)

func init() {
	// important things
	log.Warn("----------------------------------------------")
	log.Warn("-- Hi, you current use mock account service --")
	log.Warn("-- Hi, you current use mock account service --")
	log.Warn("-- Hi, you current use mock account service --")
	log.Warn("----------------------------------------------")
}

// ---------------------------------------------------------------------------------------

type Account struct {
	Account mockacc.Account
}

func (a Account) ParseAccessToken(token string) (accs.UserInfo, error) {
	return a.Account.ParseAccessToken(token)
}

func (a Account) MakeAccessToken(user accs.UserInfo) string {
	return a.Account.MakeAccessToken(user)
}

type MailService interface {
	SendInvitationMail(from, to, url string) (code int, err error)
	SendRegistrationMail(to, url string) (code int, err error)
	SendActivationMail(to, referurl, inviteurl string) (code int, err error)
	SendInvitationBounsMail(to, referurl, inviteurl, num string) (code int, err error)
	SendForgetPasswordMail(to, url string) (code int, err error)
	SendFeedbackMail(from, body string) (code int, err error)
}

type AuditLogConf struct {
	Logger    alogh.Logger
	Hosts     []string
	BodyLimit int32
}

type Config struct {
	DB     *mgo.Database
	FsHost string
	WbHost string
	MaHost string
	AuditLogConf
	Mailer MailService
}

func Run(addr string, cfg *Config, useTls bool) error {

	sa := mockacc.GetSa()
	return mockacc.Run(addr, sa)
}

func DecodeToken(token string) (result map[string]interface{}, err error) {
	err = errors.New("not supported in mockacc")
	return
}

// ---------------------------------------------------------------------------------------
