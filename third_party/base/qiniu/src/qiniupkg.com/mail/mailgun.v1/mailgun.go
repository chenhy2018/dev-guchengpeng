package mailgun

import (
	"strings"
	"time"

	mailgun "github.com/mailgun/mailgun-go"
	qlog "qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

// eg.
//	 emails := Emails("x@qiniu.com;foo@a.com,bar@b.com") // emails = ["x@qiniu.com", "foo@a.com", "bar@b.com"]
//
func Emails(s string) (emails []string) {

	for s != "" {
		pos := strings.IndexAny(s, ",;")
		if pos < 0 {
			return append(emails, s)
		}
		if pos > 0 {
			emails = append(emails, s[:pos])
		}
		s = s[pos+1:]
	}
	return
}

// -----------------------------------------------------------------------------

const (
	defaultExpiry = 3600 * 24
)

type Client struct {
	Impl mailgun.Mailgun
}

func New(domain string, apiKey, publicApiKey string) Client {

	mg := mailgun.NewMailgun(domain, apiKey, publicApiKey)
	return Client{mg}
}

type Mail struct {
	From string   `json:"from"`
	To   []string `json:"to,omitempty"`
	CC   []string `json:"cc,omitempty"`
	BCC  []string `json:"bcc,omitempty"`

	Subject  string `json:"subject,omitempty"`
	TextBody string `json:"text,omitempty"`
	HtmlBody string `json:"html,omitempty"`

	Tags           []string `json:"o:tags,omitempty"`
	Tracking       bool     `json:"o:tracking,omitempty"`
	TrackingClicks bool     `json:"o:tracking-clicks,omitempty"`
	TrackingOpen   bool     `json:"o:tracking-opens,omitempty"`

	Expiry int `json:"expiry,omitempty"`
}

func (p Client) Send(mail *Mail) (mes string, id string, err error) {

	m := p.Impl.NewMessage(mail.From, mail.Subject, mail.TextBody, mail.To...)

	if mail.HtmlBody != "" {
		m.SetHtml(mail.HtmlBody)
	}

	if mail.Tracking {
		m.SetTracking(mail.Tracking)
	}

	if mail.TrackingClicks {
		m.SetTrackingClicks(mail.TrackingClicks)
	}

	if mail.TrackingOpen {
		m.SetTrackingOpens(mail.TrackingOpen)
	}

	for _, cc := range mail.CC {
		m.AddCC(cc)
	}

	for _, bcc := range mail.BCC {
		m.AddBCC(bcc)
	}

	for _, tag := range mail.Tags {
		m.AddTag(tag)
	}

	expiry := mail.Expiry
	if expiry == 0 {
		expiry = defaultExpiry
	}
	now := time.Now()
	m.SetDeliveryTime(now.Add(time.Duration(expiry) * time.Second))

	mes, id, err = p.Impl.Send(m)
	if err != nil {
		qlog.Warn("mailgun.SendMail failed:", mes, "err:", err)
	}
	return
}

// -----------------------------------------------------------------------------

