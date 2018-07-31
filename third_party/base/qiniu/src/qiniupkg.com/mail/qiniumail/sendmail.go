package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"qiniupkg.com/dyn/text.v1"
	"qiniupkg.com/dyn/unsafe.v1"
	"qiniupkg.com/mail/mailgun.v1"
	"qiniupkg.com/x/config.v7"
	"qiniupkg.com/x/errors.v7"
)

// -----------------------------------------------------------------------------

type MailConfig struct {
	ApiKey       string `json:"apiKey"`
	PublicApiKey string `json:"publicApiKey"`

	From   string `json:"from"`
	Expiry int    `json:"expiry"`

	Tags           []string `json:"o:tags"`
	Tracking       bool     `json:"o:tracking"`
	TrackingClicks bool     `json:"o:tracking-clicks"`
	TrackingOpen   bool     `json:"o:tracking-opens"`
}

// -----------------------------------------------------------------------------

func ForEachLines(file string, doSth func(line string) error) (err error) {

	var f *os.File
	if file == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(file)
		if err != nil {
			return
		}
		defer f.Close()
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		err = doSth(scanner.Text())
		if err != nil {
			return
		}
	}
	err = scanner.Err()
	return
}

// -----------------------------------------------------------------------------

func Exitln(code int, v ...interface{}) {

	fmt.Fprintln(os.Stderr, v...)
	os.Exit(code)
}

func help() {
	Exitln(1, `Usage:
  qiniumail [-f <mail.conf>] -to <MailTo> -subject <MailSubject> <mail.body> [<maildata.mjson>]
  cat <maildata.mjson> | qiniumail [-f <mail.conf>] -to <MailTo> -subject <MailSubject> <mail.body> -
`)
}

// -----------------------------------------------------------------------------

var (
	mailTo      = flag.String("to", "", "mail to")
	mailSubject = flag.String("subject", "", "mail subject")
)

func readFile(file string) string {

	b, err := ioutil.ReadFile(file)
	if err != nil {
		Exitln(9, "read file failed:", err)
	}
	return string(b)
}

func subst(prompt, v string, data interface{}) string {

	v2, err := text.Subst(v, data, text.Fmttype_Text, true)
	if err != nil {
		Exitln(8, prompt, v, "-", errors.Detail(err))
	}
	return v2
}

func substs(prompt string, items []string, data interface{}) []string {

	for i, v := range items {
		v2, err := text.Subst(v, data, text.Fmttype_Text, true)
		if err != nil {
			Exitln(8, prompt, v, "-", errors.Detail(err))
		}
		items[i] = v2
	}
	return items
}

func setMailBody(mail *mailgun.Mail, file, body string) {

	switch filepath.Ext(file) {
	case ".htm", ".html":
		mail.HtmlBody = body
	default:
		mail.TextBody = body
	}
}

func main() {

	config.Init("f", "qiniumail", "mail.conf")

	var conf MailConfig
	err := config.Load(&conf)
	if err != nil {
		os.Exit(2)
	}

	narg := flag.NArg()
	if narg < 1 || *mailTo == "" || *mailSubject == "" {
		help()
	}
	if conf.ApiKey == "" || conf.PublicApiKey == "" || conf.From == "" {
		Exitln(3, "invalid mail conf: field `apiKey`, `publicApiKey` or `from` not specified")
	}

	pos := strings.Index(conf.From, "@")
	if pos <= 0 {
		Exitln(4, "invalid mail conf: field `from` is invalid email address")
	}
	domain := conf.From[pos+1:]

	to := mailgun.Emails(*mailTo)
	if len(to) == 0 {
		Exitln(5, "invalid switch `-to <MailTo>`")
	}

	mg := mailgun.New(domain, conf.ApiKey, conf.PublicApiKey)
	file := flag.Arg(0)
	body := readFile(file)
	if narg == 1 {
		mail := &mailgun.Mail{
			From:           conf.From,
			To:             to,
			Subject:        *mailSubject,
			Tags:           conf.Tags,
			Tracking:       conf.Tracking,
			TrackingClicks: conf.TrackingClicks,
			TrackingOpen:   conf.TrackingOpen,
			Expiry:         conf.Expiry,
		}
		setMailBody(mail, file, body)
		_, _, err = mg.Send(mail)
		if err != nil {
			Exitln(6, "send mail failed:", err)
		}
	} else {
		err = ForEachLines(flag.Arg(1), func(line string) error {
			var data interface{}
			err := json.Unmarshal(unsafe.ToBytes(line), &data)
			if err != nil {
				return err
			}
			mail := &mailgun.Mail{
				From:           conf.From,
				To:             substs("invalid send-to address:", to, data),
				Subject:        subst("invalid subject:", *mailSubject, data),
				Tags:           conf.Tags,
				Tracking:       conf.Tracking,
				TrackingClicks: conf.TrackingClicks,
				TrackingOpen:   conf.TrackingOpen,
				Expiry:         conf.Expiry,
			}
			setMailBody(mail, file, subst("invalid mail body:", body, data))
			_, _, err = mg.Send(mail)
			if err != nil {
				Exitln(6, "send mail failed:", err)
			}
			return nil
		})
		if err != nil {
			Exitln(7, errors.Detail(err))
		}
	}
}

// -----------------------------------------------------------------------------

