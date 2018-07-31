package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

var (
	DefaultSettings = Settings{}
)

func Send(msg *Message) error {
	return DefaultSettings.Send(msg)
}

type Message struct {
	Subject string
	Content string
	To      string
	Cc      string
	Bcc     string
}

type Settings struct {
	Address            string `json:"Address"`
	Port               int    `json:"Port"`
	Domain             string `json:"Domain"`
	UserName           string `json:"UserName"`
	UserNameText       string `json:"UserNameText"`
	PassWord           string `json:"PassWord"`
	ReplyUser          string `json:"ReplyUser"`
	Authentication     string `json:"Authentication"`
	EnableStarttlsAuto bool   `json:"EnableStarttlsAuto"`
}

func (m *Settings) Send(msg *Message) error {
	a := smtp.PlainAuth("", m.UserName, m.PassWord, m.Domain)
	var crlf = "\r\n"
	var buf bytes.Buffer

	allTo := strings.Split(msg.To, ",")
	fmt.Fprintf(&buf, "Reply-To: %s%s", m.ReplyUser, crlf)
	fmt.Fprintf(&buf, "Date: %s%s", time.Now().Format(time.RFC822), crlf)
	fmt.Fprintf(&buf, "From: =?UTF-8?B?%s?= <%s>%s", base64.StdEncoding.EncodeToString([]byte(m.UserNameText)), m.UserName, crlf)
	fmt.Fprintf(&buf, "To: %s%s", msg.To, crlf)
	if len(msg.Cc) > 0 {
		fmt.Fprintf(&buf, "Cc: %s%s", msg.Cc, crlf)
		allTo = append(allTo, strings.Split(msg.Cc, ",")...)
	}

	if len(msg.Bcc) > 0 {
		//fmt.Fprintf(&buf, "Bcc: %s%s", bcc, crlf)
		allTo = append(allTo, strings.Split(msg.Bcc, ",")...)
	}

	//Content-Type: multipart/alternative; boundary=047d7b33d6d615d96d04cc67487b
	fmt.Fprintf(&buf, "Subject: =?UTF-8?B?%s?=%s", base64.StdEncoding.EncodeToString([]byte(msg.Subject)), crlf)
	fmt.Fprint(&buf, "Content-Type: multipart/alternative; boundary=047d7b33d6d615d96d04cc67487b", crlf, crlf)
	fmt.Fprint(&buf, "--047d7b33d6d615d96d04cc67487b", crlf)
	fmt.Fprint(&buf, "Content-Type: text/html; charset=UTF-8", crlf)
	fmt.Fprint(&buf, "Content-Transfer-Encoding: base64", crlf, crlf)
	fmt.Fprint(&buf, base64.StdEncoding.EncodeToString([]byte(msg.Content)), crlf)
	fmt.Fprint(&buf, "--047d7b33d6d615d96d04cc67487b--", crlf)

	return m.send(m.Address+":"+strconv.Itoa(m.Port), a, m.UserName, allTo, buf.Bytes())
}

func (m *Settings) send(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}

	if ok, _ := c.Extension("AUTH"); a != nil && ok {
		resp := []byte("\x00" + m.UserName + "\x00" + m.PassWord)
		str := base64.StdEncoding.EncodeToString(resp)
		_, _, err := m.cmd(c, 0, "AUTH %s %s", "PLAIN", str)
		if err != nil {
			return err
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func (m *Settings) cmd(c *smtp.Client, expectCode int, format string, args ...interface{}) (int, string, error) {
	id, err := c.Text.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	code, msg, err := c.Text.ReadResponse(expectCode)
	return code, msg, err
}
