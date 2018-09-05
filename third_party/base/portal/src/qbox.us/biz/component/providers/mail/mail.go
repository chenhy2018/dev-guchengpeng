package mail

import (
	"qbox.us/biz/services.v2/mail"
)

func MailService(conf interface{}) interface{} {

	switch c := conf.(type) {
	case MailgunConf:
		return func() (mailService mail.MailService) {
			return mail.NewMailgunService(c.ApiKey, c.MailDomain, c.From, c.Name, c.Reply)
		}
	case SMTPConf:
		return func() (mailService mail.MailService) {
			return mail.NewSMTPService(c.Host, c.User, c.Password, c.From, c.To, c.Name, c.Reply, c.Port)
		}
	default:
		panic("mail conf is not support.")
	}
}

type MailgunConf struct {
	ApiKey     string
	MailDomain string
	From       string
	Name       string
	Reply      string
}

type SMTPConf struct {
	Host     string
	User     string
	Password string
	From     string
	To       string
	Name     string
	Reply    string
	Port     int
}
