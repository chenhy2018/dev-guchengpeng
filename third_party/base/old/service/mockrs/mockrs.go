package mockrs

import (
	"encoding/base64"
	"fmt"
	"http"
	"io/ioutil"
	"json"
	"os"
	"path"
	"qbox.us/api"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/sstore"
	"strconv"
	"strings"
	"time"
)

const DefaultPerm = 0755

type Config struct {
	Root    string
	IoHost  string
	Account account.Interface
}

type MockRS struct {
	Config
	handlers map[string]func(p *MockRS, query []string, user account.UserInfo) (code int, data interface{})
}

func New(cfg Config) *MockRS {
	handlers := map[string]func(p *MockRS, query []string, user account.UserInfo) (code int, data interface{}){
		"stat":   Stat,
		"delete": Delete,
		"move":   Move,
		"copy":   Copy,
	}
	p := &MockRS{cfg, handlers}
	return p
}

func (p *MockRS) init(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}

	user, err := account.GetAuth(p.Account, req)
	if err != nil {
		w1.ReplyWithCode(api.BadOAuthRequest)
		return
	}

	dir := path.Join(p.Root, user.Id)
	err = os.MkdirAll(dir, DefaultPerm)
	if err != nil {
		goto fail
	}

	w1.ReplyWithCode(api.OK)
	return

fail:
	w1.ReplyWithError(api.FunctionFail, err)
}

func (p *MockRS) batch(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}

	user, err := account.GetAuth(p.Account, req)
	if err != nil {
		w1.ReplyWithCode(api.BadOAuthRequest)
		return
	}

	err = req.ParseForm()
	if err != nil {
		w1.ReplyWithCode(api.InvalidArgs)
		return
	}

	ret := []map[string]interface{}{}
	rvcode := 200
	if ops, ok := req.Form["op"]; ok {
		fmt.Println(user.Id, ops)
		for _, op := range ops {
			if len(op) > 0 && op[0] == '/' {
				op = op[1:]
			}
			query := strings.Split(op, "/", -1)
			code, data := Process(p, query, user)
			if code != 200 {
				rvcode = api.PartialOK
			}
			item := map[string]interface{}{"code": code}
			if data != nil {
				item["data"] = data
			}
			ret = append(ret, item)
		}
	}

	w1.ReplyWith(rvcode, ret)
}

type Meta struct {
	Hash     string "hash"
	MimeType string "mimeType"
	Fhandle  []byte "fh"
	EditTime int64  "editTime"
	Fsize    int64  "fsize"
}

func getMeta(path string) (meta *Meta, err os.Error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	var meta1 Meta
	err = json.Unmarshal(data, &meta1)
	if err == nil {
		meta = &meta1
	}
	return
}

func putMeta(path string, meta *Meta) os.Error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, DefaultPerm)
}

//
// POST /put/<EncodedEntryURI>/fsize/<FileSize>/mimeType/<MimeType>
// Content-Type: application/octet-stream
// Body: <FileHandle>
//
func (p *MockRS) put(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}

	user, err := account.GetAuth(p.Account, req)
	if err != nil {
		w1.ReplyWithCode(api.BadOAuthRequest)
		return
	}

	query := strings.Split(req.RawURL[1:], "/", -1)

	code, path := getPath(p, query, user)
	if code != api.OK {
		w1.ReplyWithCode(code)
		return
	}

	meta := new(Meta)
	meta.Fhandle, err = ioutil.ReadAll(req.Body)
	if err != nil {
		w1.ReplyWithError(api.FunctionFail, err)
		return
	}

	var val int64
	meta.Fsize = -1
	for i := 2; i+1 < len(query); i += 2 {
		switch query[i] {
		case "mimeType":
			meta.MimeType, err = rpc.DecodeURI(query[i+1])
			if err != nil {
				w1.ReplyWithCode(api.InvalidArgs)
				return
			}
		case "fsize":
			val, err = strconv.Atoi64(query[i+1])
			if err != nil {
				w1.ReplyWithCode(api.InvalidArgs)
				return
			}
			meta.Fsize = val
		}
	}

	if meta.Fsize == -1 {
		w1.ReplyWithCode(api.InvalidArgs)
		return
	}

	meta.Hash = base64.URLEncoding.EncodeToString(meta.Fhandle)
	meta.EditTime = time.Nanoseconds() / 100 // 100ns

	err = putMeta(path, meta)
	if err != nil {
		w1.ReplyWithError(api.FunctionFail, err)
		return
	}
	w1.ReplyWith(api.OK, map[string]string{
		"hash": meta.Hash,
	})
	return
}

func (p *MockRS) get(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}
	user, _ := account.GetAuth(p.Account, req)

	query := strings.Split(req.RawURL[1:], "/", -1)
	code, data := Get(p, query, user, req)
	if code != 0 {
		if data != nil {
			w1.ReplyWith(code, data)
		} else {
			w1.ReplyWithCode(code)
		}
	} else {
		w1.ReplyError(400, "Bad method: "+query[0])
	}
}

func (p *MockRS) process(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}
	user, _ := account.GetAuth(p.Account, req)

	query := strings.Split(req.RawURL[1:], "/", -1)
	code, data := Process(p, query, user)
	if code != 0 {
		if data != nil {
			w1.ReplyWith(code, data)
		} else {
			w1.ReplyWithCode(code)
		}
	} else {
		w1.ReplyError(400, "Bad method: "+query[0])
	}
}

func Process(p *MockRS, query []string, user account.UserInfo) (code int, data interface{}) {
	if handler, ok := p.handlers[query[0]]; ok {
		return handler(p, query, user)
	}
	return 400, map[string]string{"error": "Bad method: " + query[0]}
}

func getPath(p *MockRS, query []string, user account.UserInfo) (code int, path1 string) {

	if user.Id == "" {
		return api.BadOAuthRequest, ""
	}

	if len(query) < 2 {
		return api.InvalidArgs, ""
	}

	uri, err := rpc.DecodeURI(query[1])
	if err != nil {
		return api.InvalidArgs, ""
	}

	base := path.Join(p.Root, user.Id)
	if len(uri) <= 0 {
		return api.OK, base
	}

	absPath := base + "/" + base64.URLEncoding.EncodeToString([]byte(uri))
	return api.OK, absPath
}

func getPath2(p *MockRS, query []string, user account.UserInfo) (code int, path1 string, path2 string) {

	if user.Id == "" {
		return api.BadOAuthRequest, "", ""
	}

	if len(query) < 2 {
		return api.InvalidArgs, "", ""
	}

	uri, err := rpc.DecodeURI(query[1])
	if err != nil {
		return api.InvalidArgs, "", ""
	}

	uri2, err := rpc.DecodeURI(query[2])
	if err != nil {
		return api.InvalidArgs, "", ""
	}

	base := path.Join(p.Root, user.Id)

	absPath := base + "/" + base64.URLEncoding.EncodeToString([]byte(uri))
	absPath2 := base + "/" + base64.URLEncoding.EncodeToString([]byte(uri2))

	return api.OK, absPath, absPath2
}

func Delete(p *MockRS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path := getPath(p, query, user)
	if code != api.OK {
		return
	}

	err := os.Remove(path)
	if err != nil {
		return api.FunctionFail, rpc.Error(err)
	}

	return api.OK, nil
}

func Move(p *MockRS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path1, path2 := getPath2(p, query, user)
	if code != api.OK {
		return
	}

	err := os.Rename(path1, path2)
	if err != nil {
		return api.FunctionFail, rpc.Error(err)
	}

	return api.OK, nil
}

func Copy(p *MockRS, query []string, user account.UserInfo) (code int, data interface{}) {
	return
}

func getEntry(path string) map[string]interface{} {
	meta, _ := getMeta(path)
	if meta != nil {
		entry := map[string]interface{}{}
		entry["hash"] = meta.Hash
		entry["fsize"] = meta.Fsize
		entry["editTime"] = meta.EditTime
		entry["mimeType"] = meta.MimeType
		return entry
	}
	return nil
}

func Stat(p *MockRS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path := getPath(p, query, user)
	if code != api.OK {
		return
	}

	entry := getEntry(path)
	if entry == nil {
		return api.FunctionFail, &rpc.ErrorRet{"no such file or directory"}
	}

	return api.OK, entry
}

const KeyHint = 133

var Key = []byte("qbox.mockrs")

func Get(p *MockRS, query []string, user account.UserInfo, req *http.Request) (code int, data interface{}) {

	code, path1 := getPath(p, query, user)
	if code != api.OK {
		return
	}

	var sid string
	for i := 2; i+1 < len(query); i += 2 {
		switch query[i] {
		case "sid":
			sid = query[i+1]
		case "base":
			//base = query[i+1]
		}
	}

	meta, err := getMeta(path1)
	if err != nil {
		return api.FunctionFail, rpc.Error(err)
	}

	_, fname := path.Split(path1)

	fi := &sstore.FhandleInfo{
		Fhandle:  meta.Fhandle,
		MimeType: meta.MimeType,
		AttName:  fname,
		Fsize:    meta.Fsize,
		Uid:      user.Uid,
		Deadline: time.Nanoseconds() + (1e9 * 60 * 60 * 1), // 1小时
		KeyHint:  KeyHint,
	}

	err = sstore.DecodeOid(fi, sid)
	if err != nil {
		return api.InvalidArgs, api.EInvalidArgs
	}

	fmt.Println("Fhandle:", base64.URLEncoding.EncodeToString(meta.Fhandle), "From:", req.RemoteAddr)

	if len(fi.Fhandle) == 20 { // oldver patch
		fi.Fhandle = append([]byte{0}, fi.Fhandle...)
	}

	data1 := map[string]interface{}{
		"hash":     meta.Hash,
		"fsize":    meta.Fsize,
		"mimeType": meta.MimeType,
		"url":      p.IoHost + "/file/" + sstore.EncodeFhandle(fi, Key),
	}
	return api.OK, data1
}

func Mklink(p *MockRS, query []string, user account.UserInfo) (code int, data interface{}) {
	return
}

func RegisterHandlers(mux *http.ServeMux, cfg Config) os.Error {
	if cfg.Account == nil {
		return os.EINVAL
	}
	p := New(cfg)
	mux.HandleFunc("/put/", func(w http.ResponseWriter, req *http.Request) { p.put(w, req) })
	mux.HandleFunc("/get/", func(w http.ResponseWriter, req *http.Request) { p.get(w, req) })
	mux.HandleFunc("/batch", func(w http.ResponseWriter, req *http.Request) { p.batch(w, req) })
	mux.HandleFunc("/init", func(w http.ResponseWriter, req *http.Request) { p.init(w, req) })
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { p.process(w, req) })
	return nil
}

func Run(addr string, cfg Config) os.Error {
	mux := http.DefaultServeMux
	RegisterHandlers(mux, cfg)
	return http.ListenAndServe(addr, mux)
}
