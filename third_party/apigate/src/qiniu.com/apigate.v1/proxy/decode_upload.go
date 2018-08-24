package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"strings"

	. "code.google.com/p/go.net/context"
	"github.com/qiniu/apigate.v1"
	"github.com/qiniu/rtutil"
	"github.com/qiniu/xlog.v1"

	"qiniu.com/apigate.v1/auth/common"
	"qiniu.com/auth/account.v1"
	"qiniu.com/auth/authstub.v1"
	authp "qiniu.com/auth/proto.v1"
)

var (
	peekSize = 1024 * 64
)

func respErr(xl *xlog.Logger, code int, msg string) (resp *http.Response) {
	xl.Warnf("respErr: code: %v, msg: %s\n", code, msg)
	b, _ := json.Marshal(map[string]string{"error": msg})
	resp = &http.Response{
		StatusCode: code,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		ContentLength: int64(len(b)),
		Body:          ioutil.NopCloser(bytes.NewReader(b)),
	}
	return
}

type UpTransport struct {
	transport http.RoundTripper
	acc       authp.Interface
}

func NewUpTransport(transport http.RoundTripper, acc authp.Interface) *UpTransport {

	if transport == nil {
		transport = http.DefaultTransport
	}
	transport = rtutil.New570RT(transport)
	transport = rtutil.NewUnexpectedEOFRT(transport)

	upProxy := &UpTransport{
		transport: transport,
		acc:       acc,
	}

	return upProxy
}

func filterBody(r io.Reader, wOut io.WriteCloser, c chan []byte, boundary string) io.Reader {

	nlDashBoundaryDash := []byte("\r\n--" + boundary + "--")
	pr, pw := io.Pipe()
	r = io.TeeReader(r, pw)
	bufr := bufio.NewReaderSize(r, peekSize)
	go func() {
		defer wOut.Close()
		defer pw.Close()
		// copy to wOut until encounter final boundary
		for {
			peek, err := bufr.Peek(peekSize)
			if err == io.EOF && len(peek) == 0 {
				// not found final boundary
				return
			}
			if err != nil && err != io.EOF {
				pw.CloseWithError(err)
				return
			}

			idx := bytes.Index(peek, nlDashBoundaryDash)
			if idx == -1 {
				// \r may be the member of final boundary
				rIdx := bytes.LastIndex(peek, []byte{'\r'})
				if rIdx == -1 || rIdx < peekSize-len(nlDashBoundaryDash) {
					rIdx = peekSize - 1
				}
				io.CopyN(wOut, bufr, int64(rIdx))
			} else {
				io.CopyN(wOut, bufr, int64(idx))
				break
			}
		}

		// remain should write to req.Body for multipart's buffer
		// expected to be `\r\n--boundary--`
		remain, err := ioutil.ReadAll(bufr)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()

		// hold final boundary until insert a part
		b := <-c
		if b != nil {
			// insert part
			wOut.Write(b)
		}
		wOut.Write(remain)
		return
	}()
	return pr
}

var testHook = func() {}

func (p *UpTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	if req.Method == "OPTIONS" || req.Method == "HEAD" {
		return p.transport.RoundTrip(req)
	}

	xl := xlog.NewWithReq(req)

	v := req.Header.Get("Content-Type")
	if v == "" {
		return respErr(xl, 400, "request Content-Type isn't multipart/form-data but EMPTY"), nil
	}
	d, params, err := mime.ParseMediaType(v)
	if err != nil || d != "multipart/form-data" {
		return respErr(xl, 400, "request Content-Type isn't multipart/form-data but "+d), nil
	}
	boundary, ok := params["boundary"]
	if !ok {
		return respErr(xl, 400, fmt.Sprintf("no multipart boundary param in Content-Type(%s)", v)), nil
	}

	ctx := xlog.NewContext(Background(), xl)

	pr, pw := io.Pipe()
	body := req.Body
	go func() {

		var sent bool
		c := make(chan []byte, 1)
		defer func() {
			if !sent {
				c <- nil
			}
			testHook()
		}()

		r := filterBody(body, pw, c, boundary)
		defer io.Copy(ioutil.Discard, r)
		mr := multipart.NewReader(r, boundary)

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				xl.Error("mr.NextPart failed:", err)
				return
			}
			if part.FormName() == "token" && !sent {
				// parse token
				// set stubtoken
				b, err := ioutil.ReadAll(part)
				if err != nil {
					xl.Error("ioutil.ReadAll failed:", err)
					return
				}
				user, data, err := common.ParseDataAuth(p.acc, ctx, string(b))
				if err != nil {
					xl.Error("common.ParseDataAuth failed:", err)
					part.Close()
					// set response code as 401, so do nothing
					continue
				}

				stubtoken := authstub.FormatStubData(&user, data)
				by := encodePart(boundary, "stubtoken", stubtoken)
				c <- by
				sent = true

			}
			part.Close()
		}
	}()

	req.ContentLength = -1
	req.Body = &readCloser{PipeReader: pr, Closer: body}

	resp, err = p.transport.RoundTrip(req)
	return
}

type readCloser struct {
	*io.PipeReader
	io.Closer
}

func (rc *readCloser) Close() error {
	rc.PipeReader.Close()
	return rc.Closer.Close()
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func encodePart(boundary, fieldname, value string) []byte {

	b := new(bytes.Buffer)
	// not the first part
	fmt.Fprintf(b, "\r\n--%s\r\n", boundary)
	fmt.Fprintf(b, "%s: %s\r\n",
		"Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(fieldname)))
	fmt.Fprintf(b, "\r\n")

	b.Write([]byte(value))

	return b.Bytes()
}

func nilDirector(req *http.Request) {}

func InitUpProxy(tp http.RoundTripper, acc account.Account) {

	t := NewUpTransport(tp, acc)
	proxy := &httputil.ReverseProxy{
		Director:  nilDirector,
		Transport: t,
	}

	apigate.RegisterProxy("qbox/multipart-token", proxy)
}
