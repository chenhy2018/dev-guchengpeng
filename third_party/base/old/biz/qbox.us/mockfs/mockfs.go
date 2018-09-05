package mockfs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"qbox.us/api"
	"qbox.us/api/fs"
	"qbox.us/cc/osl"
	"qbox.us/cc/time"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/sstore"
	"strconv"
	"strings"
	"syscall"
)

const DefaultPerm = 0755
const DeletedPerm = 0700

type Config struct {
	Root    string
	IoHost  string
	Account account.InterfaceEx
}

type MockFS struct {
	Config
	handlers map[string]func(p *MockFS, query []string, user account.UserInfo) (code int, data interface{})
}

func New(cfg Config) *MockFS {
	handlers := map[string]func(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}){
		"stat":     Stat,
		"list":     List,
		"mkdir":    Mkdir,
		"mkdir_p":  MkdirAll,
		"delete":   Delete,
		"undelete": Undelete,
		"purge":    Purge,
		"move":     Move,
		"copy":     Copy,
	}
	p := &MockFS{cfg, handlers}
	return p
}

func error_(err error) (code int, data interface{}) {
	if e, ok := err.(*os.PathError); ok {
		err = e.Err
	} else if e, ok := err.(*os.LinkError); ok {
		err = e.Err
	}
	msg := err.Error()
	if msg == "directory not empty" {
		return fs.EntryExists, rpc.ErrorRet{"file exists"}
	}
	return api.HttpCode(err), rpc.ErrorRet{msg}
}

func replyError(w rpc.ResponseWriter, err error) {
	code, data := error_(err)
	w.ReplyWith(code, data)
}

func (p *MockFS) init(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}

	user, err := account.GetAuthExt(p.Account, req)
	if err != nil {
		w1.ReplyWithCode(api.BadToken)
		return
	}

	var apiVer1 int

	req.ParseForm()
	if apiVer, ok := req.Form["apiVer"]; ok {
		apiVer1, _ = strconv.Atoi(apiVer[0])
	}
	if apiVer1 < 1 {
		w1.ReplyWithCode(api.VersionTooOld)
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
	replyError(w1, err)
}

func (p *MockFS) batch(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}

	user, err := account.GetAuthExt(p.Account, req)
	if err != nil {
		w1.ReplyWithCode(api.BadToken)
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
			query := strings.Split(op, "/")
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
	Hash     string `json:"hash"`
	Fhandle  []byte `json:"fh"`
	EditTime int64  `json:"editTime"`
	Fsize    int64  `json:"fsize"`
	Perm     uint32 `json:"perm"`
}

func getMeta(path string) (meta *Meta, err error) {
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

func putMeta(path string, meta *Meta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, DefaultPerm)
}

//
// POST /put/<EncodedEntryURI>/base/<BaseHash>/editTime/<EditTime>/perm/<Permission>/fsize/<FileSize>/hash/<Hash>
// Content-Type: application/octet-stream
// Body: <FileHandle>
//
func (p *MockFS) put(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}

	user, err := account.GetAuthExt(p.Account, req)
	if err != nil {
		w1.ReplyWithCode(api.BadToken)
		return
	}

	query := strings.Split(req.URL.Path[1:], "/")

	code, path, _ := getPath(p, query, user)
	if code != api.OK {
		w1.ReplyWithCode(code)
		return
	}

	meta := new(Meta)
	meta.Fhandle, err = ioutil.ReadAll(req.Body)
	if err != nil {
		replyError(w1, err)
		return
	}

	var base, hash string
	var val int64
	var ival int
	meta.Fsize = -1
	for i := 2; i+1 < len(query); i += 2 {
		switch query[i] {
		case "base":
			base = query[i+1]
		case "hash":
			hash = query[i+1]
		case "editTime":
			val, err = strconv.ParseInt(query[i+1], 10, 64)
			if err != nil {
				w1.ReplyWithCode(api.InvalidArgs)
				return
			}
			meta.EditTime = val
		case "fsize":
			val, err = strconv.ParseInt(query[i+1], 10, 64)
			if err != nil {
				w1.ReplyWithCode(api.InvalidArgs)
				return
			}
			meta.Fsize = val
		case "perm":
			ival, err = strconv.Atoi(query[i+1])
			if err != nil {
				w1.ReplyWithCode(api.InvalidArgs)
				return
			}
			meta.Perm = uint32(ival)
		}
	}

	if meta.Fsize == -1 || hash == "" {
		replyError(w1, api.EInvalidArgs)
		return
	}

	if isDeleted(path) {
		os.RemoveAll(path)
	}
	meta0, _ := getMeta(path)
	if meta0 == nil {
		if base != "" {
			w1.ReplyError(fs.Conflicted, "conflicted")
			return
		}
	} else {
		if meta0.Hash != base {
			w1.ReplyError(fs.Conflicted, "conflicted")
			return
		}
	}

	meta.Hash = hash
	if meta.EditTime == 0 {
		meta.EditTime = time.Nanoseconds() / 100 // 100ns
	}
	err = putMeta(path, meta)
	if err != nil {
		replyError(w1, err)
		return
	}
	w1.ReplyWith(api.OK, map[string]string{
		"id":   path,
		"hash": meta.Hash,
	})
	return
}

func (p *MockFS) get(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}
	user, _ := account.GetAuthExt(p.Account, req)

	query := strings.Split(req.URL.Path[1:], "/")
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

func (p *MockFS) process(w http.ResponseWriter, req *http.Request) {

	w1 := rpc.ResponseWriter{w}
	user, _ := account.GetAuthExt(p.Account, req)

	query := strings.Split(req.URL.Path[1:], "/")
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

func Process(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {
	if handler, ok := p.handlers[query[0]]; ok {
		return handler(p, query, user)
	}
	return 400, map[string]string{"error": "Bad method: " + query[0]}
}

func join(base, uri string) string {
	if len(uri) == 0 {
		return base
	}
	if uri[0] == ':' {
		// not supported
	}
	pos := strings.Index(uri, ":")
	if pos < 0 {
		return path.Join(base, uri)
	}
	id := uri[:pos]
	fname := uri[pos+1:]
	return path.Join("/", id, fname)
}

func getPath(p *MockFS, query []string, user account.UserInfo) (code int, path1, uri1 string) {

	if user.Id == "" {
		return api.BadToken, "", ""
	}

	if len(query) < 2 {
		return api.InvalidArgs, "", ""
	}

	uri, err := rpc.DecodeURI(query[1])
	if err != nil {
		return api.InvalidArgs, "", ""
	}

	base := path.Join(p.Root, user.Id)
	if len(uri) <= 0 {
		return api.OK, base, ""
	}

	absPath := join(base, uri)
	if uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}

	return api.OK, absPath, uri
}

func getPath2(p *MockFS, query []string, user account.UserInfo) (code int, path1 string, path2 string) {

	if user.Id == "" {
		return api.BadToken, "", ""
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
	return api.OK, join(base, uri), join(base, uri2)
}

func isDeletedEntry(fi os.FileInfo) bool {
	return (fi.Mode() & 0777) == DeletedPerm
}

func isDeleted(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return (fi.Mode() & 0777) == DeletedPerm
}

func mkdir(p *MockFS, query []string, user account.UserInfo, all bool) (code int, data interface{}) {

	code, dir, _ := getPath(p, query, user)
	if code != api.OK {
		return
	}

	if isDeleted(dir) {
		os.RemoveAll(dir)
	}

	var err error
	if all {
		err = os.MkdirAll(dir, DefaultPerm)
	} else {
		err = os.Mkdir(dir, DefaultPerm)
	}
	if err != nil {
		return error_(err)
	}

	return api.OK, map[string]string{"id": dir}
}

func Mkdir(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {
	return mkdir(p, query, user, false)
}

func MkdirAll(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {
	return mkdir(p, query, user, true)
}

func Delete(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path, _ := getPath(p, query, user)
	if code != api.OK {
		return
	}

	err := os.Chmod(path, DeletedPerm)
	if err != nil {
		return error_(err)
	}

	return api.OK, nil
}

func Undelete(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path, _ := getPath(p, query, user)
	if code != api.OK {
		return
	}

	err := os.Chmod(path, DefaultPerm)
	if err != nil {
		return error_(err)
	}

	return api.OK, nil
}

func Purge(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {
	return
}

func Move(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path1, path2 := getPath2(p, query, user)
	if code != api.OK {
		return
	}

	if isDeleted(path2) {
		os.RemoveAll(path2)
	}
	err := os.Rename(path1, path2)
	if err != nil {
		return error_(err)
	}

	return api.OK, nil
}

func Copy(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {
	return
}

var (
	NullHashArr [21]byte
	NullHash    = base64.URLEncoding.EncodeToString(NullHashArr[:])
)

func getEntry(fi os.FileInfo, path, uri string, showType int) map[string]interface{} {

	var ftype int

	if fi.IsDir() {
		ftype = fs.Dir
	} else {
		ftype = fs.File
		if showType == fs.ShowDirOnly {
			return nil
		}
	}

	entry := map[string]interface{}{
		"id":   path,
		"uri":  uri,
		"type": ftype,
	}

	if false { // linkId
		entry["linkId"] = ""
	}

	if (fi.Mode() & 0777) == DeletedPerm {
		if showType != -1 && showType != fs.ShowDefault {
			return nil
		}
		entry["deleted"] = 1
	}

	if !fi.IsDir() {
		meta, _ := getMeta(path)
		if meta != nil {
			if showType != -1 && showType != fs.ShowIncludeUploading {
				if meta.Hash == NullHash {
					return nil
				}
			} else if meta.Hash == NullHash {
				entry["type"] = fs.UploadingFile
			}
			entry["hash"] = meta.Hash
			entry["fsize"] = meta.Fsize
			entry["editTime"] = meta.EditTime
			if meta.Perm != 0 {
				entry["perm"] = meta.Perm
			}
			//entry["mimeType"] = ""
		}
	} else {
		entry["editTime"] = osl.Mtime(fi) / 100
	}
	return entry
}

func Stat(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path, uri := getPath(p, query, user)
	if code != api.OK {
		return
	}

	fi, err := os.Lstat(path)
	if err != nil {
		return error_(err)
	}

	entry := getEntry(fi, path, uri, -1)
	return api.OK, entry
}

func List(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {

	code, path, uri := getPath(p, query, user)
	if code != api.OK {
		return
	}

	var showType int = 0
	for i := 2; i+1 < len(query); i += 2 {
		switch query[i] {
		case "showType":
			var err error
			showType, err = strconv.Atoi(query[i+1])
			if err != nil {
				return api.InvalidArgs, rpc.ErrorRet{err.Error()}
			}
		}
	}

	fi, err := os.Lstat(path)
	if err != nil {
		return fs.NoSuchEntry, nil
	}
	if !fi.IsDir() {
		return fs.NotADirectory, nil
	} else if isDeletedEntry(fi) {
		return fs.NoSuchEntry, nil
	}

	var d *os.File
	d, err = os.Open(path)
	if err != nil {
		return error_(err)
	}
	defer d.Close()
	entries := []map[string]interface{}{}
	for {
		fis, _ := d.Readdir(1)
		if len(fis) == 0 {
			break
		}
		path1 := path + "/" + fis[0].Name()
		uri1 := uri + "/" + fis[0].Name()
		entry1 := getEntry(fis[0], path1, uri1, showType)
		if entry1 != nil {
			entries = append(entries, entry1)
		}
	}
	return api.OK, entries
}

const KeyHint = 113

var Key = []byte("qbox.mockfs")

//
// POST /get/<EncodedEntryURI>/base/<BaseHash>/attName/<AttName>
//
func Get(p *MockFS, query []string, user account.UserInfo, req *http.Request) (code int, data interface{}) {

	code, path1, _ := getPath(p, query, user)
	if code != api.OK {
		return
	}

	var base string
	for i := 2; i+1 < len(query); i += 2 {
		switch query[i] {
		//		case "sid":
		//			sid = query[i+1]
		case "base":
			base = query[i+1]
			//		case "attName":
			//			attName = query[i+1]
		}
	}

	if isDeleted(path1) {
		return error_(fs.ENoSuchEntry)
	}

	meta, err := getMeta(path1)
	if err != nil {
		return error_(err)
	}

	if base != "" {
		if meta.Hash != base { // file modified
			return error_(fs.EFileModified)
		}
	}

	_, fname := path.Split(path1)

	fi := &sstore.FhandleInfo{
		Fhandle:  meta.Fhandle,
		MimeType: "application/octet-stream",
		AttName:  fname,
		Fsize:    meta.Fsize,
		Uid:      user.Uid,
		Deadline: time.Nanoseconds() + (1e9 * 60 * 60 * 1), // 1小时
		KeyHint:  KeyHint,
	}

	/*	err = sstore.DecodeOid(fi, sid)
		if err != nil {
			return error_(api.EInvalidArgs)
		}
	*/
	fmt.Println("Fhandle:", base64.URLEncoding.EncodeToString(meta.Fhandle), "From:", req.RemoteAddr)

	if len(fi.Fhandle) == 20 { // oldver patch
		fi.Fhandle = append([]byte{0}, fi.Fhandle...)
	}

	data1 := map[string]interface{}{
		"id":       path1,
		"hash":     meta.Hash,
		"fsize":    meta.Fsize,
		"editTime": meta.EditTime,
		"url":      p.IoHost + "/file/" + sstore.EncodeFhandle(fi, Key),
	}
	if meta.Perm != 0 {
		data1["perm"] = meta.Perm
	}
	return api.OK, data1
}

func Mklink(p *MockFS, query []string, user account.UserInfo) (code int, data interface{}) {
	return
}

func RegisterHandlers(mux *http.ServeMux, cfg Config) error {
	if cfg.Account == nil {
		return syscall.EINVAL
	}
	p := New(cfg)
	mux.HandleFunc("/put/", func(w http.ResponseWriter, req *http.Request) { p.put(w, req) })
	mux.HandleFunc("/get/", func(w http.ResponseWriter, req *http.Request) { p.get(w, req) })
	mux.HandleFunc("/batch", func(w http.ResponseWriter, req *http.Request) { p.batch(w, req) })
	mux.HandleFunc("/init", func(w http.ResponseWriter, req *http.Request) { p.init(w, req) })
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { p.process(w, req) })
	return nil
}

func Run(addr string, cfg Config) error {
	mux := http.DefaultServeMux
	RegisterHandlers(mux, cfg)
	return http.ListenAndServe(addr, mux)
}
