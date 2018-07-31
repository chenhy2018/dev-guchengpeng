package ios

import (
	"errors"
	"io"
	"log"
	"net/http"
	"qbox.us/api"
	"syscall"
)

var EIoUnresumable = errors.New("io unresumable")

// ----------------------------------------------------------

type DoGetRet struct {
	URL         string
	Data        interface{}
	Code        int
	Unresumable bool
}

type Getter interface {
	DoGet(requestURL string) (ret DoGetRet, err error)
}

type ResumableReader struct {
	RequestURL string
	Sid        string
	g          Getter

	body   io.ReadCloser
	offset int64
	DoGetRet
}

func OpenResumableReader(requestURL string, sid string, g Getter) (r *ResumableReader) {

	return &ResumableReader{
		RequestURL: requestURL, Sid: sid, g: g,
	}
}

func (r *ResumableReader) Close() (err error) {

	if r.body != nil {
		log.Println("ResumableReader terminating ...")
		go r.body.Close()
		r.body = nil
		log.Println("ResumableReader terminated!")
	}
	return nil
}

func (r *ResumableReader) Request() (err error) {

	if r.body != nil {
		return
	}

	return r.doRequest(false)
}

func (r *ResumableReader) doRequest(getBody bool) (err error) {

retry:
	if r.URL == "" {
		if r.Unresumable {
			return EIoUnresumable
		}
		r.DoGetRet, err = r.g.DoGet(r.RequestURL)
		if err != nil {
			if r.Unresumable {
				log.Println("ResumableReader unresumable:", err, "-", r.RequestURL)
				return EIoUnresumable
			}
			log.Println("ResumableReader.Get failed:", err, "-", r.RequestURL)
			return
		}
	}

	if getBody {
		var resp *http.Response
		resp, err = GetRange(r.URL, r.Sid, r.offset, -1)
		if err != nil {
			log.Println("ResumableReader.GetRange failed:", err)
			return
		}
		if resp.StatusCode/100 != 2 {
			if resp.StatusCode == Expired {
				r.URL = ""
				log.Println("ResumableReader: url expired, retry")
				goto retry
			}
			err = api.NewError(resp.StatusCode)
			log.Println("ResumableReader.GetRange failed:", err)
			resp.Body.Close()
			return
		}
		r.body = resp.Body
	}
	return
}

func (r *ResumableReader) doRead(buf []byte) (n int, err error) {

	if r.body == nil {
		err = r.doRequest(true)
		if err != nil {
			return
		}
	}

	n, err = r.body.Read(buf)
	if n != 0 {
		r.offset += int64(n)
	}

	if err != nil {
		r.body.Close()
		r.body = nil
	}
	return
}

func (r *ResumableReader) Read(buf []byte) (n int, err error) {

	n, err = r.doRead(buf)
	if err == nil || err == io.EOF || err == EIoUnresumable {
		return
	}

	n, err = r.doRead(buf)
	return
}

func (r *ResumableReader) Seek(offset int64, whence int) (ret int64, err error) {

	switch whence {
	case 0:
		ret = offset // nothing to do
	case 1:
		ret = offset + r.offset
	default:
		log.Println("ResumableReader.Seek: invalid arguments")
		err = syscall.EINVAL
		return
	}
	if ret == r.offset {
		return
	}
	r.offset = ret

	if r.body != nil {
		r.body.Close()
		r.body = nil
	}
	return
}

// ----------------------------------------------------------
