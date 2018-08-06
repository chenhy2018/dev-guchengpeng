package fs

import (
	"qbox.us/api"
	"qbox.us/api/ios"
	"qbox.us/rpc"
	"strconv"
)

// ----------------------------------------------------------

const (
	OutOfSpace        = 607 // FS: 空间满（User over quota）
	FileModified      = 608 // FS: 文件被修改（see fs.GetIfNotModified）
	Conflicted        = 609 // FS: 冲突
	NotAFile          = 610 // FS: 指定的 Entry 不是一个文件
	NotADirectory     = 611 // FS: 指定的 Entry 不是一个目录
	NoSuchEntry       = 612 // FS: 指定的 Entry 不存在或已经 Deleted
	NotADeletedEntry  = 613 // FS: 指定的 Entry 不是一个已经删除的条目
	EntryExists       = 614 // FS: 要创建的 Entry 已经存在
	CircularAction    = 615 // FS: 操作发生循环，无法完成
	NoSuchDirectory   = 616 // FS: Move 操作的 Parent Directory 不存在
	Locked            = 617 // FS: 要操作的 Entry 被锁，操作暂时无法进行
	DirectoryNotEmpty = 618 // FS: rmdir - directory not empty
	BadData           = 619 // FS: 数据已被破坏
	ConditionNotMeet  = 620 // FS: 条件不满足
)

var (
	EOutOfSpace        = api.RegisterError(OutOfSpace, "out of space")
	EFileModified      = api.RegisterError(FileModified, "file modified")
	EConflicted        = api.RegisterError(Conflicted, "conflicted")
	ENotAFile          = api.RegisterError(NotAFile, "not a file")
	ENotADirectory     = api.RegisterError(NotADirectory, "not a directory")
	ENoSuchEntry       = api.RegisterError(NoSuchEntry, "no such file or directory")
	ENotADeletedEntry  = api.RegisterError(NotADeletedEntry, "not a deleted entry")
	EEntryExists       = api.RegisterError(EntryExists, "file exists")
	ECircularAction    = api.RegisterError(CircularAction, "circular action")
	ENoSuchDirectory   = api.RegisterError(NoSuchDirectory, "no such directory")
	ELocked            = api.RegisterError(Locked, "locked")
	EDirectoryNotEmpty = api.RegisterError(DirectoryNotEmpty, "directory not empty")
	EBadData           = api.RegisterError(BadData, "bad data")
	EConditionNotMeet  = api.RegisterError(ConditionNotMeet, "contition not meet")
)

// ----------------------------------------------------------

const (
	File          = 0x0001
	Dir           = 0x0002
	UploadingFile = 0x0003
)

const (
	ShowDefault          = 0x0000
	ShowNormal           = 0x0001
	ShowDirOnly          = 0x0002
	ShowIncludeUploading = 0x0003
)

type Info struct {
	Used  int64 `json:"used"`
	Quota int64 `json:"quota"`
}

type ExtInfo struct {
	Used         int64                 `json:"used" bson:"used"`
	Quota        int64                 `json:"quota" bson:"quota"`
	Prize        map[string]PrizeEntry `json:"prize" bson:"prize"`
	BaseQuotaInM int                   `json:"base" bson:"base"`
}

type PrizeEntry struct {
	Deadline int64  `json:"exp" bson:"exp"`
	Reason   string `json:"msg" bson:"msg"`
	SpaceInM int    `json:"val" bson:"val"`
}

type MakeRet struct {
	Id string `json:"id"`
}

type PutRet struct {
	Id   string `json:"id"`
	Hash string `json:"hash"`
	Alt  string `json:"alt"`
}

type GetRet struct {
	Id  string `json:"id"`
	URL string `json:"url"`

	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	EditTime int64  `json:"editTime"`
	MimeType string `json:"mimeType"`
	Perm     uint32 `json:"perm"`
}

type Entry struct {
	Id      string `json:"id"`
	LinkId  string `json:"linkId"`
	URI     string `json:"uri"`
	Type    int32  `json:"type"`
	Deleted int32  `json:"deleted"`

	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	EditTime int64  `json:"editTime"`
	MimeType string `json:"mimeType"`
	FPub     int    `json:"fpub"`
	Perm     uint32 `json:"perm"`
}

const (
	API_VER = "1"
)

// ----------------------------------------------------------

func (fs *Service) Init() (code int, err error) {
	return fs.Conn.CallWithForm(nil, fs.Host+"/init", map[string][]string{"apiVer": {API_VER}})
}

func (fs *Service) Info() (info Info, code int, err error) {
	code, err = fs.Conn.Call(&info, fs.Host+"/info")
	return
}

func (fs *Service) ExtInfo() (info ExtInfo, code int, err error) {
	code, err = fs.Conn.Call(&info, fs.Host+"/ext_info")
	return
}

// ----------------------------------------------------------

func (fs *Service) Get(entryURI string, sid string) (data GetRet, code int, err error) {
	code, err = fs.Conn.Call(&data, fs.Host+"/get/"+rpc.EncodeURI(entryURI)+"/sid/"+sid)
	return
}

func (fs *Service) GetIfNotModified(entryURI string, sid string, base string) (data GetRet, code int, err error) {
	code, err = fs.Conn.Call(&data, fs.Host+"/get/"+rpc.EncodeURI(entryURI)+"/sid/"+sid+"/base/"+base)
	return
}

func (fs *Service) GetResumable(entryURI string, sid string, base string) (r *ios.ResumableReader) {
	requestUrl := fs.Host + "/get/" + rpc.EncodeURI(entryURI) + "/sid/" + sid + "/base/" + base
	return ios.OpenResumableReader(requestUrl, sid, fs)
}

func (fs *Service) DoGet(requestURL string) (ret ios.DoGetRet, err error) {

	var getRet GetRet
	ret.Code, err = fs.Conn.Call(&getRet, requestURL)
	ret.Data = &getRet
	ret.URL = getRet.URL
	if err != nil {
		if ret.Code == FileModified || ret.Code == NotAFile || ret.Code == NoSuchEntry {
			ret.Unresumable = true
		}
	}
	return
}

// ----------------------------------------------------------

type publishRet struct {
	Id string `json:"id" bson:"id"`
}

func (fs *Service) Publish(entryURI string) (publishId string, code int, err error) {
	var ret publishRet
	code, err = fs.Conn.Call(&ret, fs.Host+"/publish/"+rpc.EncodeURI(entryURI))
	if err == nil {
		publishId = ret.Id
	}
	return
}

func (fs *Service) Unpublish(entryURI string) (code int, err error) {
	code, err = fs.Conn.Call(nil, fs.Host+"/unpublish/"+rpc.EncodeURI(entryURI))
	return
}

// ----------------------------------------------------------

func (fs *Service) Stat(entryURI string) (entry Entry, code int, err error) {
	code, err = fs.Conn.Call(&entry, fs.Host+"/stat/"+rpc.EncodeURI(entryURI))
	return
}

func (fs *Service) List(entryURI string) (entries []Entry, code int, err error) {
	code, err = fs.Conn.Call(&entries, fs.Host+"/list/"+rpc.EncodeURI(entryURI))
	return
}

func (fs *Service) ListWith(entryURI string, showType int) (entries []Entry, code int, err error) {
	code, err = fs.Conn.Call(&entries, fs.Host+"/list/"+rpc.EncodeURI(entryURI)+"/showType/"+strconv.Itoa(showType))
	return
}

// ----------------------------------------------------------

func (fs *Service) mksth(entryURI string, method string) (id string, code int, err error) {
	var ret MakeRet
	code, err = fs.Conn.Call(&ret, fs.Host+method+rpc.EncodeURI(entryURI))
	if code == 200 && ret.Id == "" {
		return "", api.UnexpectedResponse, api.EUnexpectedResponse
	}
	id = ret.Id
	return
}

func (fs *Service) Mkdir(entryURI string) (id string, code int, err error) {
	return fs.mksth(entryURI, "/mkdir/")
}

func (fs *Service) MkdirAll(entryURI string) (id string, code int, err error) {
	return fs.mksth(entryURI, "/mkdir_p/")
}

func (fs *Service) Mklink(entryURI string) (id string, code int, err error) {
	return fs.mksth(entryURI, "/mklink/")
}

// ----------------------------------------------------------

func (fs *Service) Delete(entryURI string) (code int, err error) {
	return fs.Conn.Call(nil, fs.Host+"/delete/"+rpc.EncodeURI(entryURI))
}

func (fs *Service) Undelete(entryURI string) (code int, err error) {
	return fs.Conn.Call(nil, fs.Host+"/undelete/"+rpc.EncodeURI(entryURI))
}

func (fs *Service) Purge(entryURI string) (code int, err error) {
	return fs.Conn.Call(nil, fs.Host+"/purge/"+rpc.EncodeURI(entryURI))
}

// ----------------------------------------------------------

func (fs *Service) Move(entryURISrc, entryURIDest string) (code int, err error) {
	return fs.Conn.Call(nil, fs.Host+"/move/"+rpc.EncodeURI(entryURISrc)+"/"+rpc.EncodeURI(entryURIDest))
}

func (fs *Service) Copy(entryURISrc, entryURIDest string) (code int, err error) {
	return fs.Conn.Call(nil, fs.Host+"/copy/"+rpc.EncodeURI(entryURISrc)+"/"+rpc.EncodeURI(entryURIDest))
}

// ----------------------------------------------------------

type BatchRet struct {
	Data interface{} `json:"data"`
	Code int         `json:"code"`
}

type Batcher struct {
	op  []string
	ret []BatchRet
}

func (b *Batcher) mksth(entryURI string, method string) {
	var ret MakeRet
	b.op = append(b.op, method+rpc.EncodeURI(entryURI))
	b.ret = append(b.ret, BatchRet{Data: &ret})
}

func (b *Batcher) Mklink(entryURI string) {
	b.mksth(entryURI, "/mklink/")
}

func (b *Batcher) Mkdir(entryURI string) {
	b.mksth(entryURI, "/mkdir/")
}

func (b *Batcher) MkdirAll(entryURI string) {
	b.mksth(entryURI, "/mkdir_p/")
}

func (b *Batcher) Operate(method string, entryURI string) {
	b.op = append(b.op, method+rpc.EncodeURI(entryURI))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) Operate2(method string, entryURISrc, entryURIDest string) {
	b.op = append(b.op, method+rpc.EncodeURI(entryURISrc)+"/"+rpc.EncodeURI(entryURIDest))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) Delete(entryURI string) {
	b.Operate("/delete/", entryURI)
}

func (b *Batcher) Undelete(entryURI string) {
	b.Operate("/undelete/", entryURI)
}

func (b *Batcher) Purge(entryURI string) {
	b.Operate("/purge/", entryURI)
}

func (b *Batcher) Move(entryURISrc, entryURIDest string) {
	b.Operate2("/move/", entryURISrc, entryURIDest)
}

func (b *Batcher) Copy(entryURISrc, entryURIDest string) {
	b.Operate2("/copy/", entryURISrc, entryURIDest)
}

func (b *Batcher) Reset() {
	b.op = nil
	b.ret = nil
}

func (b *Batcher) Len() int {
	return len(b.op)
}

func (b *Batcher) Do(fs *Service) (ret []BatchRet, code int, err error) {
	code, err = fs.Conn.CallWithForm(&b.ret, fs.Host+"/batch", map[string][]string{"op": b.op})
	ret = b.ret
	return
}

// ----------------------------------------------------------
