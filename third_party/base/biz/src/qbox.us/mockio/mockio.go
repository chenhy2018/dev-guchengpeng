package mockio

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/api"
	"qbox.us/auditlog"
	"qbox.us/cc/time"
	"qbox.us/fileop"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/servend/oauth"
	"qbox.us/servestk"
	"qbox.us/sstore"
)

const DefaultPerm = 0755
const ChunkBits = 22
const ChunkSize = 1 << ChunkBits // 4M

type Config struct {
	Root    string
	Account account.InterfaceEx
	Client  map[string]string
	auditlog.Config
}

type MockIO struct {
	Config
	*sync.Mutex
	chanId, fid      int64
	chanDir, dataDir string
	MyHost           string
}

var Operations = fileop.NewOperations()

func errorOf(err error) interface{} {
	return rpc.ErrorRet{err.Error()}
}

func (r *MockIO) Fopproxy(ctx context.Context, w http.ResponseWriter, req fileop.Reader) {}

func (r *MockIO) GetAccount() account.Interface {
	return r.Account
}

func (r *MockIO) GetIoHost() string {
	return "" // r.MyHost
}

func (r *MockIO) GetHost(svrName string) (svrHost string, ok bool) {
	// svrHost, ok = r.Clients[svrName]
	return
}

func (r *MockIO) GetCache() fileop.Cache {
	return nil
}

func (r *MockIO) FileOp(tpCtx context.Context, w fileop.Writer, r1 fileop.Reader) {
	Operations.Do(tpCtx, w, r1)
}

func New(cfg Config) (p *MockIO, err error) {
	chanDir := cfg.Root + "/chan/"
	dataDir := cfg.Root + "/data/"
	err = os.MkdirAll(chanDir, DefaultPerm)
	if err != nil {
		return
	}
	err = os.MkdirAll(dataDir, DefaultPerm)
	if err != nil {
		return
	}
	myHost := "http://localhost:7779"
	p = &MockIO{cfg, new(sync.Mutex), 0, 0, chanDir, dataDir, myHost}
	return
}

func Mkchan(p *MockIO, user account.UserInfo) (code int, data interface{}) {
	p.Lock()
	defer p.Unlock()
	p.chanId++
	return api.OK, map[string]interface{}{
		"id":        strconv.FormatInt(p.chanId, 10),
		"chunkSize": ChunkSize,
	}
}

func CreateFile(p *MockIO, chanId, callback string, fsize int64, first io.Reader, user account.UserInfo) (code int, data interface{}) {
	p.Lock()
	defer p.Unlock()
	p.fid++
	fsize1 := strconv.FormatInt(fsize, 10)
	fid := chanId + ":" + strconv.FormatInt(p.fid, 10) + ":" + fsize1 + ":" + callback

	fname := p.chanDir + "/" + chanId
	code = api.FunctionFail
	f, err := os.Create(fname)
	if err != nil {
		return code, errorOf(err)
	}
	defer f.Close()

	n, err := io.Copy(f, first)
	if err != nil {
		return code, errorOf(err)
	}

	var data1 interface{}
	if n >= fsize {
		if n > fsize {
			return api.InvalidArgs, nil
		}
		data1, code, err = p.commit(f, fsize, callback, user)

		if err != nil {
			return code, errorOf(err)
		}
	}

	result := map[string]interface{}{"fid": fid}
	if data1 != nil {
		result["data"] = data1
	}
	return api.OK, result
}

func Upload(w rpc.ResponseWriter, p *MockIO, query []string, user account.UserInfo, body io.Reader, fsize int64, replyIfOk bool, doCache bool, ins bool) (notReplyed bool) {

	method := "/put/"
	callback1 := p.Client["rs"] + method + query[1]

	p.Lock()
	p.chanId++
	p.Unlock()
	chan1 := strconv.FormatInt(p.chanId, 10)
	code, data := CreateFile(p, chan1, callback1, fsize, body, user)
	if code != 200 {
		w.ReplyWith(code, data)
		return
	}

	result := data.(map[string]interface{})
	if replyIfOk {
		w.ReplyWith(200, result["data"])
		return
	}
	return true
}

func WriteAt(p *MockIO, chanId, fid string, offset int64, next io.Reader, user account.UserInfo) (code int, data interface{}) {

	parts := strings.SplitN(fid, ":", 4)
	if len(parts) != 4 || parts[0] != chanId {
		return api.InvalidArgs, nil
	}

	fsize, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || offset > fsize {
		return api.InvalidArgs, nil
	}

	fname := p.chanDir + chanId

	code = api.FunctionFail
	f, err := os.OpenFile(fname, os.O_RDWR, 0)
	if err != nil {
		return code, errorOf(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return code, errorOf(err)
	}
	if fi.Size() < offset {
		return api.InvalidArgs, nil
	}

	_, err = f.Seek(offset, os.SEEK_SET)
	if err != nil {
		return code, errorOf(err)
	}

	n, err := io.Copy(f, next)
	if err != nil {
		return code, errorOf(err)
	}
	n += offset

	if n >= fsize {
		if n > fsize {
			return api.InvalidArgs, nil
		}
		var data1 interface{}
		data1, code, err = p.commit(f, fsize, parts[3], user)
		if err != nil {
			return code, errorOf(err)
		}
		return api.OK, map[string]interface{}{
			"data": data1,
		}
	}

	return api.OK, nil
}

func (p *MockIO) mkchan(w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}
	user, err := account.GetAuth(p.Account, req)
	if err != nil {
		w.ReplyWithCode(api.BadToken)
		return
	}

	code, data := Mkchan(p, user)
	w.ReplyWith(code, data)
}

func (p *MockIO) commit(f *os.File, fsize int64, callback string, user account.UserInfo) (data interface{}, code int, err error) {
	code = api.FunctionFail
	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return
	}

	hash := sha1.New()

	if fsize <= ChunkSize {
		_, err = io.Copy(hash, f)
		if err != nil {
			return
		}
	} else {
		hash1 := sha1.New()
		for off := int64(0); off < fsize; off += ChunkSize {
			_, err = io.CopyN(hash1, f, ChunkSize)
			if err != nil && err != io.EOF {
				return
			}
			hash.Write(hash1.Sum(nil))
			hash1.Reset()
		}
	}

	callback1 := callback

	if err != nil && err != io.EOF {
		return
	}

	fh := append([]byte{ChunkBits}, hash.Sum(nil)...)
	hash1 := base64.URLEncoding.EncodeToString(fh)
	fto := p.dataDir + hash1
	fsrc := f.Name()
	f.Close()
	err = os.Rename(fsrc, fto)
	if err != nil {
		_, err = os.Lstat(fto)
		if err != nil {
			return
		}
	}

	token := p.Account.MakeAccessToken(user)
	conn := rpc.Client{oauth.NewClient(token, nil)}
	callback1 += "/fsize/" + strconv.FormatInt(fsize, 10) + "/hash/" + hash1
	code, err = conn.CallWithParam(&data, callback1, "application/octet-stream", fh)
	return
}

//POST /chan/<ChannelId>/create/<EncodedCallback>/fsize/<FileSize>
//POST /chan/<ChannelId>/put/<Fid>/<Offset>
func (p *MockIO) chanop(w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}
	user, err := account.GetAuth(p.Account, req)
	if err != nil {
		w.ReplyWithCode(api.BadToken)
		return
	}

	query := strings.Split(req.URL.Path[1:], "/")
	if len(query) < 3 {
		w.ReplyWithCode(api.InvalidArgs)
		return
	}

	var val int64
	var code int
	var data interface{}

	switch query[2] {
	case "put":
		if len(query) < 5 {
			w.ReplyWithCode(api.InvalidArgs)
			return
		}
		val, err = strconv.ParseInt(query[4], 10, 64)
		if err != nil {
			w.ReplyWithCode(api.InvalidArgs)
			return
		}
		code, data = WriteAt(p, query[1], query[3], val, req.Body, user)
	case "create":
		if len(query) < 6 || query[4] != "fsize" {
			w.ReplyWithCode(api.InvalidArgs)
			return
		}
		val, err = strconv.ParseInt(query[5], 10, 64)
		if err != nil {
			w.ReplyWithCode(api.InvalidArgs)
			return
		}
		code, data = CreateFile(p, query[1], query[3], val, req.Body, user)
	default:
		w.ReplyWithCode(api.InvalidArgs)
		return
	}

	w.ReplyWith(code, data)
	return
}

// GET /file/<EncodedFileHandle>/<Operation>/<OperationParams>
func (p *MockIO) file(tpCtx context.Context, w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}

	query := strings.Split(req.URL.Path[1:], "/")
	if len(query) < 2 {
		w.ReplyWithCode(api.InvalidArgs)
		return
	}

	var method func(context.Context, fileop.Writer, fileop.Reader)
	if len(query) > 2 {
		if method = Operations[query[2]]; method == nil {
			w.ReplyError(400, "Bad method")
			return
		}
	}

	fhi := sstore.DecodeFhandle(query[1], "", KeyFinder)
	if fhi == nil {
		w.ReplyWithCode(api.BadOAuthRequest)
		return
	}

	if fhi.Fhandle[0] == 0 { // oldver patch
		fhi.Fhandle = fhi.Fhandle[1:]
	}

	file := p.dataDir + base64.URLEncoding.EncodeToString(fhi.Fhandle)
	f, err := os.Open(file)
	if err != nil {
		w.ReplyWithError(api.FunctionFail, err)
		return
	}
	defer f.Close()

	if method != nil {
		r := fileop.Reader{
			Source:      FileSource{f},
			FhandleInfo: fhi,
			Request:     req,
			Env:         p,
			Query:       query[2:],
		}
		method(tpCtx, fileop.Writer{w}, r)
		return
	}

	r := fileop.Reader{
		Source:      FileSource{f},
		FhandleInfo: fhi,
		Request:     req,
	}
	hash, err := r.QueryHash()
	if err != nil {
		w.ReplyWithError(api.FunctionFail, err)
		return
	}

	meta := &rpc.Metas{
		ETag:            base64.URLEncoding.EncodeToString(hash),
		MimeType:        fhi.MimeType,
		DispositionType: "attachment",
		FileName:        fhi.AttName,
		CacheControl:    "public, max-age=31536000",
	}
	w.ReplyRange2(f, fhi.Fsize, meta, req)
}

type UploadHandle struct {
	Uid      uint32
	Utype    uint32
	Devid    uint32
	Appid    uint32
	Expires  uint32
	Callback string
}

type AppCallback struct {
	URL      string
	BodyType string
}

func (p *MockIO) upload(w1 http.ResponseWriter, req *http.Request) {
	w := rpc.ResponseWriter{w1}
	query := strings.SplitN(req.URL.Path[1:], "/", 2)
	if len(query) != 2 {
		log.Warn("Invalid query:", req.URL.Path)
		w.ReplyWithCode(400)
		return
	}

	b, err := base64.URLEncoding.DecodeString(query[1])
	if err != nil {
		log.Warn("Base64Decode:", err)
		w.ReplyWithCode(400)
		return
	}

	var handle UploadHandle
	err = json.Unmarshal(b, &handle)
	if err != nil {
		log.Warn("auth.Decode:", err)
		w.ReplyWithCode(400)
		return
	}

	user := account.UserInfo{
		Uid:     handle.Uid,
		Utype:   handle.Utype,
		Devid:   handle.Devid,
		Appid:   handle.Appid,
		Expires: handle.Expires,
	}

	var callback *AppCallback

	if handle.Callback != "" {
		callback = &AppCallback{URL: handle.Callback, BodyType: "application/x-www-form-urlencoded"}
	}
	if err := req.ParseMultipartForm(16 * 1024); err != nil || req.MultipartForm == nil {
		log.Warn("Upload - ParseMultipartForm failed:", err)
		w.ReplyWithCode(400)
		return
	}

	multiForm := req.MultipartForm
	defer multiForm.RemoveAll()

	file := multiForm.File["file"]
	action := multiForm.Value["action"]
	params := multiForm.Value["params"]
	if file == nil || action == nil || action[0] == "" {
		log.Info("MultiForm: no required field")
		w.ReplyWithCode(400)
		return
	}

	if callback != nil && (params == nil || params[0] == "") {
		log.Warn("Upload - AppCallbacks: params required")
		w.ReplyWithCode(400)
		return
	}
	query = strings.SplitN(action[0][1:], "/", 2)

	if strings.HasSuffix(query[0], "-put") {
		f, err := file[0].Open()
		if err != nil {
			w.ReplyWithError(api.FunctionFail, err)
			return
		}

		defer f.Close()
		fsize, err := f.Seek(0, 2)
		if err != nil {
			w.ReplyWithError(api.FunctionFail, err)
			return
		}

		f.Seek(0, 0)
		if Upload(w, p, query, user, f, fsize, callback == nil, true, true) {
			var ret interface{}
			msg := params[0]
			body := strings.NewReader(msg)
			log.Info("AppCallback:", callback.URL, callback.BodyType, msg)
			code, err := rpc.DefaultClient.CallWithBinaryEx(&ret, callback.URL, callback.BodyType, body, len(msg))
			if err != nil {
				w.ReplyWithError(code, err)
			} else {
				w.ReplyWith(code, ret)
			}
		}

	} else {
		log.Info("Invalid action:", query)
		w.ReplyWithCode(400)
		return
	}

}

func (p *MockIO) putAuth(w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}

	expiresIn := 3600
	callBack := ""

	query := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	log.Debug("IO.putAuth:", query)
	if len(query) > 1 {
		var err error
		expiresIn, err = strconv.Atoi(query[1])
		if err != nil {
			log.Warn("putAuth, atoi,", err)
			w.ReplyWithCode(api.InvalidArgs)
			return
		}
		if expiresIn > 86400 {
			expiresIn = 86400
		} // 1 day

		for i := 1; i < len(query); i++ {
			switch query[i] {
			case "callback":
				cb, err := base64.URLEncoding.DecodeString(query[i+1])
				if err != nil {
					log.Warn("putAuth, decode callback,", err)
					w.ReplyWithCode(api.InvalidArgs)
					return
				}
				callBack = string(cb)
				if !strings.HasPrefix(callBack, "http://") && !strings.HasPrefix(callBack, "https://") {
					log.Warn("putAuth, invalid callback,", callBack)
					w.ReplyWithCode(api.InvalidArgs)
					return
				}
			}
		}
	}

	user, err := account.GetAuthExt(p.Account, req)
	if err != nil {
		w.ReplyWithCode(api.BadToken)
		return
	}

	user.Expires = uint32(time.Seconds()) + uint32(expiresIn+3)
	handle := &UploadHandle{
		Uid:      user.Uid,
		Utype:    user.Utype,
		Devid:    user.Devid,
		Appid:    user.Appid,
		Expires:  user.Expires,
		Callback: callBack,
	}

	b, err := json.Marshal(handle)
	if err != nil {
		w.ReplyWithError(api.FunctionFail, err)
		return
	}

	w.ReplyWith(200, map[string]interface{}{
		"url":       p.MyHost + "/upload/" + base64.URLEncoding.EncodeToString(b),
		"expiresIn": expiresIn,
	})
}

// ioHost + "/rs-put/" + rpc.EncodeURI(entryURI) + "/mimeType/" + rpc.EncodeURI(mimeType)
func (p *MockIO) rsPut(w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}
	expiresIn := 3600
	callBack := ""
	query := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	log.Debug("IO.rs-put", query)

	user, err := account.GetAuthExt(p.Account, req)
	if err != nil {
		w.ReplyWithCode(api.BadToken)
		return
	}

	user.Expires = uint32(time.Seconds()) + uint32(expiresIn+3)
	handle := &UploadHandle{
		Uid:      user.Uid,
		Utype:    user.Utype,
		Devid:    user.Devid,
		Appid:    user.Appid,
		Expires:  user.Expires,
		Callback: callBack,
	}
	// =========================================

	var callback *AppCallback

	if handle.Callback != "" {
		callback = &AppCallback{URL: handle.Callback, BodyType: "application/x-www-form-urlencoded"}
	}

	if strings.HasSuffix(query[0], "-put") {
		if Upload(w, p, query, user, req.Body, req.ContentLength, callback == nil, true, true) {
			var ret interface{}
			msg := ""
			body := strings.NewReader(msg)
			log.Info("AppCallback:", callback.URL, callback.BodyType, msg)
			code, err := rpc.DefaultClient.CallWithBinaryEx(&ret, callback.URL, callback.BodyType, body, len(msg))
			if err != nil {
				w.ReplyWithError(code, err)
			} else {
				w.ReplyWith(code, ret)
			}
		}
		req.Body.Close()

	} else {
		log.Info("Invalid action:", query)
		w.ReplyWithCode(400)
		return
	}

}

type FileSource struct {
	*os.File
}

func (r FileSource) QueryFhandle() (fh []byte, err error) {
	return nil, syscall.EINVAL
}

func (r FileSource) UploadedSize() (uploaded int64, err error) {
	return 0, syscall.EINVAL
}

func (r FileSource) RangeRead(w io.Writer, from, to int64) (err error) {
	_, err = r.Seek(from, os.SEEK_SET)
	if err != nil {
		return
	}
	_, err = io.CopyN(w, r.File, to-from)
	return
}

func RegisterHandlers(mux1 *http.ServeMux, cfg Config) error {
	if cfg.Account == nil {
		return syscall.EINVAL
	}
	p, err := New(cfg)
	if err != nil {
		return err
	}
	lh := auditlog.NewExt("MockIo", &cfg.Config, cfg.Account)
	mux := servestk.New(mux1, lh.Handler())
	mux.HandleFunc("/mkchan", func(w http.ResponseWriter, req *http.Request) { p.mkchan(w, req) })
	mux.HandleFunc("/chan/", func(w http.ResponseWriter, req *http.Request) { p.chanop(w, req) })
	mux.HandleFunc("/file/", handleCtx(p.file))
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, req *http.Request) { p.upload(w, req) })
	mux.HandleFunc("/put-auth/", func(w http.ResponseWriter, req *http.Request) { p.putAuth(w, req) })
	mux.HandleFunc("/rs-put/", func(w http.ResponseWriter, req *http.Request) { p.rsPut(w, req) })
	return nil
}

func Run(addr string, cfg Config) error {
	mux := http.NewServeMux()
	RegisterHandlers(mux, cfg)
	return http.ListenAndServe(addr, mux)
}

func handleCtx(op func(context.Context, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		xl := xlog.New(w, req)
		ctx, cancel := context.WithCancel(context.Background())
		ctx = xlog.NewContext(ctx, xl)

		reqDone := make(chan bool, 1)
		go func() {
			op(ctx, w, req)
			reqDone <- true
		}()
		select {
		case <-httputil.GetCloseNotifierSafe(w).CloseNotify():
			cancel()
			xl.Info("connection was closed by peer:", ctx.Err())
			<-reqDone
			return
		case <-reqDone:
		}
	}

}
