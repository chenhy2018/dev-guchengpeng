package rs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"mime/multipart"
	"os"

	. "github.com/qiniu/api/conf"
	"github.com/qiniu/rpc.v1"
)

const (
	typeDirect = "direct"
	typeCopy   = "copy"
)

type PutExtra struct {
	Params map[string]string //可选，用户自定义参数，必须以 "x:" 开头
	//若不以x:开头，则忽略
	MimeType string //可选，当为 "" 时候，服务端自动判断
	Crc32    uint32
	CheckCrc uint32
	// CheckCrc == 0: 表示不进行 crc32 校验
	// CheckCrc == 1: 对于 Put 等同于 CheckCrc = 2；对于 PutFile 会自动计算 crc32 值
	// CheckCrc == 2: 表示进行 crc32 校验，且 crc32 值就是上面的 Crc32 变量
}

type Part struct {
	// check fname first
	FileName string

	// then check R
	R io.Reader

	Crc32    uint32
	CheckCrc bool

	// finally, we use Key
	Key string

	// To == -1 means the end of file
	// [From, To)
	From, To int64
}

type part struct {
	Type string `json:"type"`

	Crc32 uint32 `json:"crc32"`

	StorageFile string `json:"storageFile"`
	Range       string `json:"range"`
}

type partArg struct {
	MimeType string `json:"mimeType"`
	Parts    []part `json:"parts"`
}

func PutParts(l rpc.Logger, ret interface{}, uptoken, key string, hasKey bool, parts []Part, extra *PutExtra) error {

	if extra == nil {
		extra = &PutExtra{}
	}
	arg := partArg{MimeType: extra.MimeType}

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	err := writeMultipartOfPart(writer, uptoken, key, hasKey, extra)

	for i, p := range parts {
		if p.FileName != "" {
			f, err := os.Open(p.FileName)
			if err != nil {
				return err
			}
			err = addDirectFile(writer, &arg, i, p.CheckCrc, p.Crc32, f)
			f.Close()
			if err != nil {
				return err
			}
		} else if p.R != nil {
			err = addDirectFile(writer, &arg, i, p.CheckCrc, p.Crc32, p.R)
			if err != nil {
				return err
			}
		} else {
			err := addCopyFile(&arg, i, p.Key, p.From, p.To)
			if err != nil {
				return err
			}
		}
	}

	argB, err := json.Marshal(arg)
	if err != nil {
		return err
	}
	err = writer.WriteField("parts", string(argB))
	if err != nil {
		return err
	}
	writer.Close()

	contentType := writer.FormDataContentType()
	return rpc.DefaultClient.CallWith64(l, ret, UP_HOST+"/parts", contentType, &b, int64(b.Len()))
}

func addDirectFile(writer *multipart.Writer, arg *partArg, idx int, checkCrc bool, crc32v uint32, r io.Reader) (err error) {

	fieldName := fmt.Sprintf("part-%d", idx)
	w, err := writer.CreateFormFile(fieldName, fieldName)
	if err != nil {
		return
	}

	if checkCrc && crc32v == 0 {
		rs, ok := r.(io.ReadSeeker)
		if !ok {
			err = errors.New("r should be io.ReadSeeker if generate crc32 in sdk")
			return
		}
		crch := crc32.NewIEEE()
		_, err = io.Copy(crch, rs)
		if err != nil {
			err = errors.New("io.Copy failed:" + err.Error())
			return
		}
		crc32v = crch.Sum32()
		_, err = rs.Seek(0, os.SEEK_SET)
		if err != nil {
			err = errors.New("rs.Seek failed:" + err.Error())
			return
		}
	}

	_, err = io.Copy(w, r)
	if err != nil {
		err = errors.New("io.Copy failed:" + err.Error())
		return
	}

	arg.Parts = append(arg.Parts, part{Type: typeDirect, Crc32: crc32v})
	return
}

func addCopyFile(arg *partArg, idx int, key string, from, to int64) (err error) {

	var rangeStr string
	if to == -1 || to > from {
		rangeStr = fmt.Sprintf("%d-%d", from, to)
	} else {
		return errors.New("invalid from&to argument")
	}

	arg.Parts = append(arg.Parts, part{Type: typeCopy, StorageFile: key, Range: rangeStr})
	return
}

func writeMultipartOfPart(writer *multipart.Writer, uptoken, key string, hasKey bool, extra *PutExtra) (err error) {

	// token
	if err = writer.WriteField("token", uptoken); err != nil {
		return
	}

	// key
	if hasKey {
		if err = writer.WriteField("key", key); err != nil {
			return
		}
	}

	// extra.Params
	if extra.Params != nil {
		for k, v := range extra.Params {
			err = writer.WriteField(k, v)
			if err != nil {
				return
			}
		}
	}
	return
}
