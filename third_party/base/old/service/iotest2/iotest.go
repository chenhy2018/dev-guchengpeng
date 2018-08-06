package iotest2

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash"
	"http"
	"io"
	"os"
	fss "qbox.us/api/fs"
	"qbox.us/api/ios"
	"qbox.us/api/ios2"
	"strconv"
	"strings"
	"testing"
)

var (
	SID string = "AAAAAAAAAAAAAAAA"
)

func genRandomBufferWithSize(size int32) []byte {

	key := make([]byte, size)
	rand.Read(key)
	return key
}

func UploadWithBlockSize(bsize int32, io2 *ios2.Service, fileHash hash.Hash) (ctx []byte, code int, err os.Error) {

	var windowSize int32 = 4 * 1024

	var bufSize = bsize
	if bufSize < 0 {
		bufSize = 4194304
	}

	var sentSize int32 = 0
	var toRead int32 = 0
	rawBuffer := genRandomBufferWithSize(bufSize)

	if bufSize < windowSize {
		toRead = bufSize
	} else {
		toRead = windowSize
	}

	bbuffer := bytes.NewBuffer(rawBuffer[:toRead])

	ret, code, err := io2.Mkblk(bsize, bbuffer)
	if err != nil {
		return
	}
	if code != 200 {
		err = os.NewError("io2.Mkblk Invalid Return code: " + strconv.Itoa(code))
		return
	}

	sentSize += toRead
	for sentSize < bufSize {
		if bufSize-sentSize < windowSize {
			toRead = bufSize - sentSize
		} else {
			toRead = windowSize
		}

		bbuffer = bytes.NewBuffer(rawBuffer[sentSize : sentSize+toRead])
		ret, code, err = io2.Bput(ret.Ctx, sentSize, bbuffer)
		if err != nil {
			return
		}
		if code != 200 {
			err = os.NewError("io2.Bput Invalid Return code: " + strconv.Itoa(code))
			return
		}

		sentSize += toRead
	}

	sha1 := sha1.New()
	sha1.Write(rawBuffer)
	key, _ := base64.URLEncoding.DecodeString(ret.Checksum)
	if !bytes.Equal(sha1.Sum(), key) {
		return nil, 400, os.NewError("Verify Failed")
	}

	ctx, _ = base64.URLEncoding.DecodeString(ret.Ctx)
	if bsize == 4194304 {
		if len(ctx) != 20 {
			return nil, 400, os.NewError("invalid block ctx length")
		}
	} else if bsize > 0 {
		if len(ctx) != 24 {
			return nil, 400, os.NewError("invalid block ctx length")
		}
		s := int32(binary.LittleEndian.Uint32(ctx[20:]))
		if bsize != s {
			return nil, 400, os.NewError("invalid block ctx, size not match")
		}
	}

	if fileHash != nil {
		fileHash.Write(rawBuffer)
	}
	return
}

func AssertUploadWithBlockSize(bufSize int32, code int, io1 *ios.Service, io2 *ios2.Service, fs *fss.Service2, t *testing.T) {

	_, c, err := UploadWithBlockSize(bufSize, io2, nil)
	if c != code {
		t.Fatal("UploadWithBlockSize(", bufSize, ") error:", err, " code:", c, " expected:", code)
	}
}

func Do(io1 *ios.Service, io2 *ios2.Service, fs *fss.Service2, t *testing.T) {

	{
		body := strings.NewReader("hello, world")

		fmt.Println("iotest2.Do")
		ret, code, err := io2.Mkblk(-3, body)
		if err != nil || code != 200 {
			t.Fatal("Mkblk(-3) error: err, code, ret", err, code, ret)
		}

		ret, code, err = io2.Mkblk(32, body)
		if err != nil || code != 200 {
			t.Fatal("Mkblk(32) error: err, code, ret", err, code, ret)
		}

		ret, code, err = io2.Mkblk(4194304, body)
		if err != nil || code != 200 {
			t.Fatal("Mkblk(4194304) error: err, code, ret", err, code, ret)
		}

		ret, code, err = io2.Mkblk(4194304+1, body)
		if code != 400 {
			t.Fatal("Mkblk(4194304 + 1) error: err, code, ret", err, code, ret)
		}
	}

	//mkblk 10 bytes
	{
		body := strings.NewReader("1234567890")
		ret, code, err := io2.Mkblk(10, body)
		if err != nil || code != 200 {
			t.Fatal("Mkblk(10) error: err, code, ret", err, code, ret)
		}

		key, _ := base64.URLEncoding.DecodeString(ret.Checksum)
		//		result []byte, code int, err os.Error
		results, code, err := io1.Query(key)
		if err != nil || code != 200 || results[0] != 1 {
			t.Fatal("Query(10) error: err, code, ret", err, code, ret)
		}
	}

	//mkblk 10 bytes, but put 20 bytes
	{
		body := strings.NewReader("12345678901234567890")
		ret, code, err := io2.Mkblk(10, body)
		if code == 200 {
			t.Fatal("Mkblk(10, but put 20) error: err, code, ret", err, code, ret)
		}
	}

	//mkblk	30 bytes
	{
		body := strings.NewReader("123456789009876543211234567890")
		ret, code, err := io2.Mkblk(30, body)
		if err != nil || code != 200 {
			t.Fatal("Mkblk(30) error: err, code, ret", err, code, ret)
		}

		key, _ := base64.URLEncoding.DecodeString(ret.Checksum)
		results, code, err := io1.Query(key)
		if err != nil || code != 200 || results[0] != 1 {
			t.Fatal("Query(30) error: err, code, ret", err, code, ret)
		}
	}

	//bput
	{
		body := strings.NewReader("1234567890")
		ret, code, err := io2.Mkblk(25, body)
		if err != nil || code != 200 {
			t.Fatal("Mkblk(25) error: err, code, ret", err, code, ret)
		}

		body = strings.NewReader("098765432112345")
		ret2, code, err := io2.Bput(ret.Ctx, 10, body)
		if err != nil || code != 200 {
			t.Fatal("Bput(20) error: err, code, ret", err, code, ret2)
		}

		fullbody := make([]byte, 25)
		copy(fullbody[:], "1234567890098765432112345")
		sha1 := sha1.New()
		sha1.Write(fullbody)

		key, _ := base64.URLEncoding.DecodeString(ret2.Checksum)
		if !bytes.Equal(sha1.Sum(), key) {
			t.Fatal("Bput(20) key error: corret, wrong", sha1.Sum(), key)
		}

		results, code, err := io1.Query(key)
		if err != nil || code != 200 || results[0] != 1 {
			t.Fatal("Query(25) error: err, code, ret", err, code, ret)
		}
	}

	//test uploading blocks
	{
		AssertUploadWithBlockSize(-1, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(0, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(30, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(512, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(4*1024-1, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(4*1024, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(4*1024+1, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(1024*1024-1, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(1024*1024, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(1024*1024+1, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(4*1024*1024, 200, io1, io2, fs, t)
		AssertUploadWithBlockSize(4*1024*1024+1, 400, io1, io2, fs, t)
	}

	//test uploading files
	{
		fileHash := sha1.New()

		ctx0, code, err := UploadWithBlockSize(4194304, io2, fileHash)
		if code != 200 {
			t.Fatal("UploadWithBlockSize(", 4194304, ") error:", err, " code:", code)
		}

		//resp, err := io2.Mkfile(0, "fs-put", rpc.EncodeURI("/f4194304a.txt") + "/editTime/0/perm/0", bytes.NewBuffer(ctx0))
		params := fss.FileMeta{EditTime: 0, Alt: "f4194304a-alt.txt", Base: "", Perm: 0}
		_, code, err = fs.Mkfile("/f4194304a.txt", bytes.NewBuffer(ctx0), &params)
		if err != nil || code != 200 {
			t.Fatal("io2.Mkfile(4194304a) failed:", code, err)
		}

		ctx1, code, err := UploadWithBlockSize(4048+5, io2, fileHash)
		if code != 200 {
			t.Fatal("UploadWithBlockSize(", 4048+5, ") error:", err, " code:", code)
		}

		//resp, err = io2.Mkfile(0, "fs-put", rpc.EncodeURI("/f4053.txt"), bytes.NewBuffer(ctx1))
		params2 := fss.FileMeta{EditTime: 0, Alt: "f4053-alt.txt", Base: "", Perm: 0}
		_, code, err = fs.Mkfile("/f4053.txt", bytes.NewBuffer(ctx0), &params2)
		if err != nil || code != 200 {
			t.Fatal("io2.Mkfile(4053) failed:", code, err)
		}

		b := bytes.NewBuffer(nil)
		b.Write(ctx0)
		b.Write(ctx1)
		//resp, err = io2.Mkfile(0, "fs-put", rpc.EncodeURI("/f4194304+4053.txt") + "/editTime/2424/perm/0", b)
		params3 := fss.FileMeta{EditTime: 0, Alt: "f4194304+4053-alt.txt", Base: "", Perm: 0}
		ret3, code, err := fs.Mkfile("/f4194304+4053.txt", b, &params3)
		if err != nil || code != 200 {
			t.Fatal("io2.Mkfile(4194304+4053) failed:", code, err)
		}

		ret3hash, _ := base64.URLEncoding.DecodeString(ret3.Hash)
		fsHash := sha1.New()
		fsHash.Write(ctx0)
		fsHash.Write(ctx1[0 : len(ctx1)-4])
		if !bytes.Equal(fsHash.Sum(), ret3hash[1:]) {
			t.Fatal("io2.Mkfile(4194304+4053) hash not right:", ret3hash, fsHash.Sum())
		}

		fsret, _, err := fs.Get("/f4194304+4053.txt", SID)

		var r *http.Response
		req, err := http.NewRequest("GET", fsret.URL, nil)
		if err != nil {
			t.Fatal("io2.Mkfile(4194304+4053) download url error:", err)
		}
		req.AddCookie(&http.Cookie{Name: "sid", Value: SID})

		r, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal("io2.Mkfile(4194304+4053) download request error:", err)
		}
		defer r.Body.Close()

		downloadedFileHash := sha1.New()
		_, err = io.Copy(downloadedFileHash, r.Body)
		if err != nil {
			t.Fatal("io2.Mkfile(4194304+4053) download calc hash error:", err)
		}

		if !bytes.Equal(fileHash.Sum(), downloadedFileHash.Sum()) {
			t.Fatal("io2.Mkfile(4194304+4053) download verify file hash error:", err)
		} else {
			//			log.Info("Test Mkfile(4194304+4053): Verify downloaded file succeeded")
			t.Log("Test Mkfile(4194304+4053): Verify downloaded file succeeded")
		}
	}
}
