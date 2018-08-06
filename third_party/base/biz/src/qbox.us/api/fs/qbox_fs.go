package fs

import (
	"io"
	"log"
	"net/http"
	"os"
	"qbox.us/api"
	"qbox.us/api/ios"
	"qbox.us/cc/osl"
	"qbox.us/rpc"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ----------------------------------------------------------

type Service struct {
	Host string
	Conn rpc.Client
}

// ----------------------------------------------------------

func New(fsHost string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{fsHost, rpc.Client{client}}
}

// ----------------------------------------------------------

func NameWithTime(alt string, editTime int64) string {
	ets := time.Unix(editTime/1e7, 0).Format("2006-01-02_15_04_05")
	pos := strings.LastIndex(alt, ".")
	if pos < 0 {
		pos = len(alt)
	}
	return alt[:pos] + "(" + ets + ")" + alt[pos:]
}

// ----------------------------------------------------------

type UploadParams struct {
	EditTime int64
	Alt      string
	Base     string
	ios.UploadParams
	Perm uint32
}

func OpenToUpload(localFile string) (f *os.File, params *UploadParams, err error) {

	f, err = os.Open(localFile)
	if err != nil {
		return
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return
	}
	if !!fi.IsDir() {
		f.Close()
		err = syscall.ENFILE
		return
	}

	params = new(UploadParams)
	params.Perm = osl.Permission(fi)
	params.EditTime = osl.Mtime(fi) / 100
	params.Fsize = fi.Size()

	return
}

func (params *UploadParams) MakeURL(callback string) {

	if params.EditTime != 0 {
		callback += "/editTime/" + strconv.FormatInt(params.EditTime, 10)
	}
	if params.Alt != "" {
		callback += "/alt/" + rpc.EncodeURI(params.Alt)
	}
	if params.Base != "" {
		callback += "/base/" + params.Base
	}
	if params.Perm != 0 {
		callback += "/perm/" + strconv.Itoa(int(params.Perm))
	}
	params.Callback = callback
}

func (qbox *Service) UploadWith(
	c ios.Channel, remoteFile string, f io.Reader, params *UploadParams) (ret PutRet, code int, err error) {

	params.MakeURL(qbox.Host + "/put/" + rpc.EncodeURI(remoteFile))
	code, err = c.UploadEx(&ret, params.Callback, f, params.Fsize, params.Mode)
	return
}

func (qbox *Service) ResumableUploadWith(
	fid string, c ios.Channel, remoteFile string,
	f io.ReadSeeker, params *UploadParams) (ret PutRet, code int, err error) {

	params.MakeURL(qbox.Host + "/put/" + rpc.EncodeURI(remoteFile))
	f2 := ios.NewEtagReader(f, 1<<22)
	code, err = c.ResumableUploadEx(fid, &ret, f2, &params.UploadParams)
	if err == nil {
		if !ios.Validate(f2, params.Fsize, ret.Hash) {
			// @@todo: revert to old version
			code, err = api.NetworkError, EBadData
		}
	}
	return
}

// ----------------------------------------------------------

func (qbox *Service) Upload(
	c ios.Channel, remoteFile, localFile string, base ...string) (ret PutRet, code int, err error) {

	return qbox.UploadEx(c, ios.NormalMode, remoteFile, localFile, base...)
}

func (qbox *Service) UploadEx(
	c ios.Channel, mode int, remoteFile, localFile string, base ...string) (ret PutRet, code int, err error) {

	f, params, err := OpenToUpload(localFile)
	if err != nil {
		return
	}
	defer f.Close()

	if len(base) > 0 {
		params.Base = base[0]
	}
	params.MakeURL(qbox.Host + "/put/" + rpc.EncodeURI(remoteFile))
	code, err = c.UploadEx(&ret, params.Callback, f, params.Fsize, mode)
	return
}

func (qbox *Service) ResumableUpload(
	fid string, c ios.Channel, mode int,
	remoteFile, localFile string, base ...string) (ret PutRet, code int, err error) {

	f, params, err := OpenToUpload(localFile)
	if err != nil {
		return
	}
	defer f.Close()

	if len(base) > 0 {
		params.Base = base[0]
	}
	params.MakeURL(qbox.Host + "/put/" + rpc.EncodeURI(remoteFile))
	params.Mode = mode
	params.InitFastUpload(ios.Goroutine{}, 64, 1<<22)
	code, err = c.MetricResumableUploadEx(fid, &ret, f, &params.UploadParams)
	if ret.Id == "" {
		code, err = BadData, EBadData
		log.Println("MetricResumableUploadEx: no id")
	}
	return
}

// ----------------------------------------------------------
