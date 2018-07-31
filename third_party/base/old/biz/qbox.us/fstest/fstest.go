package fstest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	fss "qbox.us/api/fs"
	"qbox.us/objid"
	"qbox.us/rpc"
	"qbox.us/ts"
	"testing"
)

var (
	NullHashArr [21]byte
	NullHash    = base64.URLEncoding.EncodeToString(NullHashArr[:])
)

type Root interface {
	CheckDir(path string, t *testing.T)
	IsDeleted(path string) bool
}

func checkDir(path string, root Root, t *testing.T) {
	root.CheckDir(path, t)
}

func isDeleted(root Root, path string) bool {
	return root.IsDeleted(path)
}

func doTestMkdir(fs *fss.Service, t *testing.T, root Root) {
	{
		dir := "/!f+%o#o"
		id, code, err := fs.Mkdir(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Mkdir:", err, code)
		}
		fmt.Println("Mkdir:", id)
		checkDir(dir, root, t)
	}
	{
		dir := "/!f+%o#o"
		id, code, err := fs.Mkdir(dir)
		if err == nil || code != fss.EntryExists || err.Error() != "file exists" {
			ts.Fatal(t, "Mkdir:", dir, err, code)
		}
		fmt.Println("Mkdir:", id)
		checkDir(dir, root, t)
	}
	{
		dir := "f@o$o/b#*a+%r"
		id, code, err := fs.MkdirAll(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Mkdir:", err, code)
		}
		fmt.Println("Mkdir:", id)
		checkDir(dir, root, t)
	}
	{
		dir := "f@o$o/b#*a+%r"
		id, code, err := fs.MkdirAll(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Mkdir:", err, code)
		}
		fmt.Println("Mkdir:", id)
		checkDir(dir, root, t)
	}
	{
		dir := "/not/exists"
		_, code, err := fs.Mkdir(dir)
		if err == nil || code != fss.NoSuchEntry || err.Error() != "no such file or directory" {
			ts.Fatal(t, "Mkdir:", dir, err, code)
		}
	}
	{
		dir := "f@o$o/b#*a+%r"
		id, code, err := fs.MkdirAll(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Mkdir:", err, code)
		}
		fmt.Println("Mkdir:", id)
		checkDir(dir, root, t)
	}
	{
		dir := "f@o$o/b#*a+%r"
		id, code, err := fs.MkdirAll(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Mkdir:", err, code)
		}
		fmt.Println("Mkdir:", id)
		checkDir(dir, root, t)
	}
	{
		dir := "t.txt"
		_, code, err := fs.Mkdir(dir)
		if err == nil || code != fss.EntryExists || err.Error() != "file exists" {
			ts.Fatal(t, "Mkdir:", dir, err, code)
		}
	}
	{
		dir := "t.txt/aaa"
		_, code, err := fs.Mkdir(dir)
		if err == nil || code != fss.NotADirectory || err.Error() != "not a directory" {
			ts.Fatal(t, "Mkdir:", dir, err, code)
		}
	}
	{
		dir := "t.txt"
		_, code, err := fs.MkdirAll(dir)
		if err == nil || code != fss.NotADirectory || err.Error() != "not a directory" {
			ts.Fatal(t, "Mkdir:", dir, err, code)
		}
	}
	{
		dir := "t.txt/aaa"
		_, code, err := fs.MkdirAll(dir)
		if err == nil || code != fss.NotADirectory || err.Error() != "not a directory" {
			ts.Fatal(t, "Mkdir:", dir, err, code)
		}
	}
}

func doTestDelete(fs *fss.Service, t *testing.T, root Root) {

	{
		dir := "/!f+%o#o"
		code, err := fs.Delete(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Delete:", dir, err, code)
		}
		if !isDeleted(root, dir) {
			ts.Fatal(t, "Delete fail:", dir)
		}
	}

	{
		dir := "/not-exist"
		code, err := fs.Delete(dir)
		if err == nil || code != fss.NoSuchEntry || err.Error() != "no such file or directory" {
			ts.Fatal(t, "Delete:", dir, err, code)
		}
		if isDeleted(root, dir) {
			ts.Fatal(t, "Delete fail:", dir)
		}
	}

	{
		dir := "f@o$o/b#*a+%r"
		code, err := fs.Delete(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Delete:", dir, err, code)
		}
		if !isDeleted(root, dir) {
			ts.Fatal(t, "Delete fail:", dir)
		}

		code, err = fs.Undelete(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Undelete:", dir, err, code)
		}
		if isDeleted(root, dir) {
			ts.Fatal(t, "Undelete fail:", dir)
		}
	}
	{
		dir := "/aaa/bbb/ccc"
		fs.MkdirAll(dir)
		fs.Delete(dir)
		fs.Delete("/aaa")
		fs.Undelete("/aaa")
		ret, code, err := fs.Stat(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "Stat:", dir, err, code)
		}
		if ret.Deleted != 1 {
			ts.Fatal(t, "Delete/Undelete:", ret, ret.Deleted)
		}
	}
}

func doTestMove(fs *fss.Service, t *testing.T, root Root) {
	{
		fs.Mkdir("abc")
		fs.MkdirAll("/foo/bar")

		dir1 := "abc"
		dir2 := "/foo/bar/efg"
		code, err := fs.Move(dir1, dir2)
		if err != nil || code != 200 {
			ts.Fatal(t, "Move:", dir1, dir2, err, code)
		}
	}
	{
		fs.Mkdir("efg")
		fs.MkdirAll("/foo/bar")

		dir1 := "efg"
		dir2 := "/foo/bar"
		code, err := fs.Move(dir1, dir2)
		if err == nil || (code != fss.EntryExists && code != fss.DirectoryNotEmpty) ||
			(err.Error() != "file exists" && err.Error() != "directory not empty") {
			ts.Fatal(t, "Move:", dir1, dir2, err, code)
		}
	}
}

// POST /put/<EncodedEntryURI>/base/<BaseHash>/editTime/<EditTime>/fsize/<FileSize>

func doTestPut(fs *fss.Service, t *testing.T, root Root) {

	sid := objid.Encode(1234, 567890)
	{
		file := "/t.txt"
		fh := []byte("<FileHandle>")
		url := fs.Host + "/put/" + rpc.EncodeURI(file) + "/fsize/67/hash/" + base64.URLEncoding.EncodeToString(fh)

		var ret fss.PutRet
		code, err := fs.Conn.CallWithParam(&ret, url, "application/octet-stream", fh)
		if err != nil || code != 200 {
			ts.Fatal(t, "Put:", file, err, code)
		}
		fmt.Println("Put ret:", ret)
		if ret.Id == "" || ret.Hash != base64.URLEncoding.EncodeToString(fh) {
			ts.Fatal(t, "Put fail:", file, ret)
		}

		getRet, code, err := fs.Get(file, sid)
		if err != nil || code != 200 {
			ts.Fatal(t, "Get:", file, err, code)
		}
		fmt.Println("Get ret:", getRet)
		if getRet.Fsize != 67 || getRet.Hash != base64.URLEncoding.EncodeToString(fh) {
			ts.Fatal(t, "Get fail:", file, getRet)
		}

		getRet, code, err = fs.GetIfNotModified(file, sid, getRet.Hash)
		if err != nil || code != 200 {
			ts.Fatal(t, "Get:", file, err, code)
		}
		fmt.Println("Get ret:", getRet)
		if getRet.Fsize != 67 || getRet.Hash != base64.URLEncoding.EncodeToString(fh) {
			ts.Fatal(t, "Get fail:", file, getRet)
		}

		_, code, err = fs.GetIfNotModified(file, sid, base64.URLEncoding.EncodeToString([]byte("1235")))
		if err == nil || code != fss.FileModified {
			ts.Fatal(t, "Get:", file, err, code)
		}

		url = fs.Host + "/put/" + rpc.EncodeURI("/u.txt") + "/fsize/67"
		code, err = fs.Conn.CallWithParam(&ret, url+"/hash/"+NullHash, "application/octet-stream", fh)
		if err != nil || code != 200 {
			ts.Fatal(t, "Put:", "/u.txt", err, code)
		}
		fmt.Println("Put ret:", ret)
		if ret.Id == "" || ret.Hash != NullHash {
			ts.Fatal(t, "Put fail:", "/u.txt", ret)
		}

		getRet, code, err = fs.Get("/u.txt", sid)
		if err != nil || code != 200 {
			t.Fatal("Get:", "/u.txt", err, code)
		}
		fmt.Println("Get ret:", getRet)
		if getRet.Fsize != 67 || getRet.Hash != NullHash {
			ts.Fatal(t, "Get fail:", "/u.txt", getRet)
		}
	}
}

func doTestList(fs *fss.Service, t *testing.T, root Root) {
	{
		dir := "/"
		entries, code, err := fs.List(dir)
		if err != nil || code != 200 {
			ts.Fatal(t, "List:", dir, err, code)
		}
		result, _ := json.Marshal(entries)
		fmt.Println("List:", string(result))
	}
	{
		dir := "/foo"
		fs.Mkdir(dir)
		entry, code, err := fs.Stat(dir)
		if err != nil || code != 200 {
			t.Fatal("List:", dir, err, code)
		}
		result, _ := json.Marshal(entry)
		fmt.Println("Stat:", string(result))
		if entry.URI != "foo" || entry.Type != fss.Dir {
			ts.Fatal(t, "Stat fail:", entry)
		}
	}
	{
		dir := ""
		entry, code, err := fs.Stat(dir)
		if err != nil || code != 200 {
			t.Fatal("List:", dir, err, code)
		}
		result, _ := json.Marshal(entry)
		fmt.Println("Stat:", string(result))
		if entry.URI != dir || entry.Type != fss.Dir {
			ts.Fatal(t, "Stat fail:", entry)
		}
	}
	{
		dir := "/not/exists"
		entry, code, err := fs.Stat(dir)
		result, _ := json.Marshal(entry)
		fmt.Println("Stat:", string(result))
		if err == nil || err.Error() != "no such file or directory" || code != fss.NoSuchEntry {
			ts.Fatal(t, "Stat:", dir, err, code)
		}
	}
	{
		fs.Mkdir("stat")
		dir := "/stat/not-exists"
		entry, code, err := fs.Stat(dir)
		result, _ := json.Marshal(entry)
		fmt.Println("Stat:", string(result))
		if err == nil || err.Error() != "no such file or directory" || code != fss.NoSuchEntry {
			ts.Fatal(t, "Stat:", dir, err, code)
		}
	}
	{
		dir := "uploading"
		fs.Mkdir(dir)

		fh := []byte("<FileHandle>")
		var ret fss.PutRet
		url := fs.Host + "/put/" + rpc.EncodeURI(dir+"/1.txt") + "/fsize/67"
		code, err := fs.Conn.CallWithParam(&ret, url+"/hash/"+NullHash, "application/octet-stream", fh)
		if err != nil || code != 200 {
			ts.Fatal(t, "Put:", dir+"/1.txt", err, code)
		}
		url = fs.Host + "/put/" + rpc.EncodeURI(dir+"/2.txt") + "/fsize/67"
		code, err = fs.Conn.CallWithParam(&ret, url+"/hash/"+base64.URLEncoding.EncodeToString(fh), "application/octet-stream", fh)
		if err != nil || code != 200 {
			ts.Fatal(t, "Put:", dir+"/2.txt", err, code)
		}

		entries, code, err := fs.ListWith(dir, fss.ShowDefault)
		fmt.Println(entries, code, err)
		if err != nil || code != 200 {
			ts.Fatal(t, "List:", dir, err, code)
		}
		if len(entries) != 1 {
			ts.Fatal(t, "list uploading fatal")
		}
		entries, code, err = fs.ListWith(dir, fss.ShowIncludeUploading)
		fmt.Println(entries, code, err)
		if err != nil || code != 200 {
			ts.Fatal(t, "List:", dir, err, code)
		}
		if len(entries) != 2 {
			ts.Fatal(t, "list uploading fatal")
		}

		entry, code, err := fs.Stat(dir + "/1.txt")
		fmt.Println("stat uploading:", entry, code, err)
		if err != nil || code != 200 {
			ts.Fatal(t, "Stat:", dir, err, code)
		}
		if entry.Type != fss.UploadingFile {
			ts.Fatal(t, "stat fatal")
		}
		entry, code, err = fs.Stat(dir + "/2.txt")
		fmt.Println("stat uploading:", entry, code, err)
		if err != nil || code != 200 {
			ts.Fatal(t, "Stat:", dir, err, code)
		}
		if entry.Type != fss.File {
			ts.Fatal(t, "stat fatal")
		}
	}
}

func doTestBatch(fs *fss.Service, t *testing.T, root Root) {

	bat := new(fss.Batcher)
	bat.Mkdir("/!f+%o#o_2")
	bat.MkdirAll("f@o$o_2/b#*a+%r")
	if bat.Len() != 2 {
		ts.Fatal(t, "Batch failed:", bat.Len())
	}
	ret, code, err := bat.Do(fs)
	if err != nil || code != 200 || len(ret) != 2 {
		ts.Fatal(t, "Batch:", err, code)
	}
	for i, item := range ret {
		if item.Code != 200 {
			ts.Fatal(t, i, "Batch fail:", item.Code)
		}
		if data, ok := item.Data.(*fss.MakeRet); ok {
			fmt.Println(i, "Batch:", data)
		} else {
			ts.Fatal(t, i, "Batch fail:", item)
		}
	}

	bat.Reset()
	if bat.Len() != 0 {
		ts.Fatal(t, "Batch failed:", bat.Len())
	}

	bat.Delete("/!f+%o#o_2")
	bat.Delete("f@o$o_2")
	bat.Undelete("/!f+%o#o_2")
	ret, code, err = bat.Do(fs)
	if err != nil || code != 200 || len(ret) != 3 {
		ts.Fatal(t, "Batch:", err, code)
	}
	for i, item := range ret {
		if item.Code != 200 || item.Data != nil {
			ts.Fatal(t, i, "Batch fail:", item)
		}
	}
}

func doTestPath(fs *fss.Service, t *testing.T, root Root) {

	sid := objid.Encode(1234, 567890)

	cid, code, err := fs.MkdirAll("a/b/c")
	if code != 200 || err != nil {
		t.Fatal("MkdirAll:", code, err)
	}

	_, code, err = fs.MkdirAll(cid + ":d/e/f")
	if code != 200 || err != nil {
		t.Fatal("MkdirAll:", code, err)
	}

	fid, code, err := fs.MkdirAll(cid + ":/d2/e/f")
	if code != 200 || err != nil {
		t.Fatal("MkdirAll:", code, err)
	}

	{
		file := fid + ":t.txt"
		url := fs.Host + "/put/" + rpc.EncodeURI(file) + "/fsize/67"
		fh := []byte("<FileHandle>")

		var ret fss.PutRet
		code, err := fs.Conn.CallWithParam(&ret, url+"/hash/"+base64.URLEncoding.EncodeToString(fh), "application/octet-stream", fh)
		if err != nil || code != 200 {
			ts.Fatal(t, "Put:", file, err, code)
		}
		fmt.Println("Put ret:", ret)
		if ret.Id == "" || ret.Hash != base64.URLEncoding.EncodeToString(fh) {
			ts.Fatal(t, "Put fail:", file, ret)
		}

		getRet, code, err := fs.Get(ret.Id+":", sid)
		if err != nil || code != 200 {
			ts.Fatal(t, "Get:", file, err, code)
		}
		fmt.Println("Get ret:", getRet)
		if getRet.Fsize != 67 || getRet.Hash != base64.URLEncoding.EncodeToString(fh) {
			ts.Fatal(t, "Get fail:", file, getRet)
		}
	}
}

func Do(fs *fss.Service, t *testing.T, root1 Root) {
	{
		code, err := fs.Init()
		if err != nil || code != 200 {
			ts.Fatal(t, "Init:", err, code)
		}
		fmt.Println("Init OK!")
	}
	doTestPut(fs, t, root1)
	doTestMkdir(fs, t, root1)
	doTestBatch(fs, t, root1)
	doTestDelete(fs, t, root1)
	doTestList(fs, t, root1)
	doTestMove(fs, t, root1)
	doTestPath(fs, t, root1)
}
