package iotest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	fss "qbox.us/api/fs"
	"qbox.us/api/ios"
	"qbox.us/cc/sha1"
	"qbox.us/multipart"
	"qbox.us/rpc"
	"qbox.us/ts"
	"strings"
	"testing"
)

func DoUpload(io1 *ios.Service, fs *fss.Service, t *testing.T) {

	ret, code, err := io1.PutAuth()
	if err != nil {
		ts.Fatal(t, "PutAuth:", code, err)
	}
	fmt.Println("PutAuth ret:", ret)

	file := "ioTest.tmp"
	ioutil.WriteFile(file, []byte("file content"), 0777)
	defer os.Remove(file)

	r, ct, err := multipart.Open(map[string][]string{
		"action": {"/fs-put/!000-default"},
		"file":   {"@" + file},
	})
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}
	defer r.Close()

	fmt.Println("\nContent-Type:", ct)

	req, err := http.NewRequest("POST", ret.URL, r)
	if err != nil {
		ts.Fatal(t, "NewRequest failed:", err)
	}

	req.Header.Set("Content-Type", ct)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ts.Fatal(t, "http.Client.Do failed:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		ts.Fatal(t, "http.Client.Do status code:", resp.StatusCode)
	}

	io.Copy(os.Stdout, resp.Body)
}

func Do(io *ios.Service, fs *fss.Service, t *testing.T) {
	DoExt(io, fs, true, t)
}

func DoExt(io *ios.Service, fs *fss.Service, realIo bool, t *testing.T) {

	if realIo {
		DoUpload(io, fs, t)
	}

	c, code, err := io.Mkchan()
	if err != nil || code != 200 {
		ts.Fatal(t, "Mkchan:", err, code)
	}
	var base string
	{
		var ret fss.PutRet
		fname := "/foo.txt"
		callback := fs.Host + "/put/" + rpc.EncodeURI(fname)
		data := "Hello, world!"
		body := strings.NewReader(data)
		fid, code, err := c.CreateFile(&ret, callback, 13, body)
		if err != nil || code != 200 {
			ts.Fatal(t, "CreateFile:", err, code)
		}
		if ret.Id == "" || ret.Hash == "" {
			ts.Fatal(t, "CreateFile fail:", fname, ret.Id, ret.Hash)
		}
		fmt.Println("CreateFile - fid:", fid)
		fmt.Println("CreateFile - ret:", ret)
		base = ret.Hash

		sid := ios.CreateSid()
		getRet, code, err := fs.Get(fname, sid)
		if err != nil || code != 200 {
			ts.Fatal(t, "Get:", fname, err, code)
		}
		fmt.Println("Get ret:", getRet)
		if getRet.Fsize != int64(len(data)) {
			ts.Fatal(t, "Get fail:", fname, getRet)
		}

		{
			r, err := ios.Get(getRet.URL, sid)
			if err != nil {
				ts.Fatal(t, "Get:", fname, err, r.StatusCode)
			}
			defer r.Body.Close()
			if r.StatusCode != 200 {
				ts.Fatal(t, "Get:", fname, r.StatusCode)
			}

			data1, err := ioutil.ReadAll(r.Body)
			if err != nil {
				ts.Fatal(t, "Get: ReadAll -", err)
			}
			fmt.Println("Get ret data:", string(data1), len(data1))
			if data != string(data1) {
				ts.Fatal(t, "Get fail: stream data -", data, string(data1))
			}
		}

		{
			r, err := ios.GetRange(getRet.URL, sid, 2, 4)
			if err != nil {
				ts.Fatal(t, "Get:", fname, err, r.StatusCode)
			}
			defer r.Body.Close()
			if r.StatusCode != 206 {
				ts.Fatal(t, "Get:", fname, r.StatusCode)
			}

			data1, err := ioutil.ReadAll(r.Body)
			if err != nil {
				ts.Fatal(t, "Get: ReadAll -", err)
			}
			fmt.Println("Get ret data:", string(data1))
			if data[2:4] != string(data1) {
				ts.Fatal(t, "Get fail: stream data -", data[2:4], string(data1))
			}
		}
		{
			r, err := ios.GetRangeEx(getRet.URL, sid, "bytes=-3")
			if err != nil {
				ts.Fatal(t, "Get:", fname, err, r.StatusCode)
			}
			defer r.Body.Close()
			if r.StatusCode != 206 {
				ts.Fatal(t, "Get:", fname, r.StatusCode)
			}

			data1, err := ioutil.ReadAll(r.Body)
			if err != nil {
				ts.Fatal(t, "Get: ReadAll -", err)
			}
			fmt.Println("Get ret data:", string(data1))
			if data[len(data)-3:] != string(data1) {
				ts.Fatal(t, "Get fail: stream data -", data[len(data)-3:], string(data1))
			}
		}
	}
	{
		var ret fss.PutRet
		fname := "foo.txt"
		callback := fs.Host + "/put/" + rpc.EncodeURI(fname) + "/base/" + base
		body := strings.NewReader("Hello, world!!")
		fid, code, err := c.CreateFile(&ret, callback, 14, body)
		if err != nil || code != 200 {
			ts.Fatal(t, "CreateFile:", err, code)
		}
		if ret.Id == "" || ret.Hash == "" {
			ts.Fatal(t, "CreateFile fail:", fname)
		}
		fmt.Println("CreateFile - fid:", fid)
		fmt.Println("CreateFile - ret:", ret)
	}
	{
		fname := "/foo.txt"
		callback := fs.Host + "/put/" + rpc.EncodeURI(fname)
		body := strings.NewReader("Hello, world!!!")
		_, code, err := c.CreateFile(nil, callback, 15, body)
		if err == nil || err.Error() != "conflicted" || code != fss.Conflicted {
			ts.Fatal(t, "CreateFile:", err, code)
		}
	}
	{
		var ret fss.PutRet
		fname := "foo%#!.txt"
		callback := fs.Host + "/put/" + rpc.EncodeURI(fname)
		body := strings.NewReader("Hello, ")
		fid, code, err := c.CreateFile(&ret, callback, 13, body)
		if err != nil || code != 200 {
			ts.Fatal(t, "CreateFile:", err, code)
		}
		if ret.Id != "" || ret.Hash != "" {
			ts.Fatal(t, "CreateFile fail:", ret)
		}
		fmt.Println("CreateFile - fid:", fid)

		body2 := strings.NewReader("world!")
		code, err = c.WriteAt(&ret, fid, 7, body2)
		if err != nil || code != 200 {
			ts.Fatal(t, "WriteAt:", err, code)
		}
		if ret.Id == "" || ret.Hash == "" {
			ts.Fatal(t, "WriteAt fail:", fname)
		}
		fmt.Println("WriteAt - ret:", ret)
	}
	{
		b := bytes.NewBuffer(nil)
		b.Write(sha1.Hash([]byte("Hello, world!!")))
		b.Write(sha1.Hash([]byte("!!not exist!!")))
		result, code, err := io.Query(b.Bytes())
		if err != nil || code != 200 {
			ts.Fatal(t, "Query:", err, code)
		}
		if len(result) != 2 || result[0] != 1 || result[1] != 0 {
			ts.Fatal(t, "Query fail:", result)
		}
	}
}
