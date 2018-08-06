package mail

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/qiniu/log.v1"
	"qbox.us/biz/component/client"
)

type MailService interface {
	Send(msg *MailMessage) error
	SendWithAttach(msg *MailMessage, attachFile []string) error
}

type MailMessage struct {
	Name  string
	From  string
	Reply string

	To  []string
	Cc  []string
	Bcc []string

	Tag         []string
	ExtraHeader map[string]string

	Subject string
	Content string

	Options map[string]string
}

func (msg *MailMessage) Validate() error {
	// 内部使用，不检查from和to的邮件地址有效性
	if msg.Name == "" {
		return errors.New("no name")
	}

	if msg.From == "" {
		return errors.New("no from")
	}

	if msg.To == nil || len(msg.To) == 0 {
		return errors.New("no to")
	}

	if msg.Subject == "" {
		return errors.New("no subject")
	}

	if msg.Content == "" {
		return errors.New("no content")
	}
	return nil
}

func formToField(writer *multipart.Writer, params url.Values) (err error) {
	for k, v := range params {
		if len(v) == 0 {
			continue
		}
		err = writer.WriteField(k, v[0])
		if err != nil {
			return err
		}
	}
	return nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// CreateFormFile is a convenience wrapper around CreatePart. It creates
// a new form-data header with the provided field name and file name.
func CreateFormFile(w *multipart.Writer, fieldname, filename, contentType string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}

func attach(writer *multipart.Writer, fieldname, file string) error {
	filename := path.Base(file)
	mimeType := mime.TypeByExtension(path.Ext(file))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	w, err := CreateFormFile(writer, fieldname, filename, mimeType)
	if err != nil {
		return err
	}
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

type MailConfig map[string]string

const _Retry = 3

//Transport
type Transport struct {
	tr http.RoundTripper
}

func NewTransport() http.RoundTripper {
	return client.DefaultTransport
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	resp, err = t.tr.RoundTrip(req)
	elaplsed := time.Since(start)
	var code int
	if resp != nil {
		code = resp.StatusCode
	}
	addr := req.Host + req.URL.Path
	str := fmt.Sprintln("Service:", addr, "Code:", code, "Err:", err, "Time:", elaplsed.Nanoseconds()/1e6, "ms")
	log.Info(str)
	return
}
