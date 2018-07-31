// +build go1.5

package fopg

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"

	"qbox.us/fop"
)

func EncodeQuery(fh []byte, fsize int64, ctx *fop.FopCtx) string {
	v := url.Values{}
	v.Set("fh", base64.URLEncoding.EncodeToString(fh))
	v.Set("fsize", strconv.FormatInt(fsize, 10))
	v.Set("cmd", ctx.RawQuery)
	v.Set("uid", strconv.FormatUint(uint64(ctx.Uid), 10))
	v.Set("url", ctx.URL)
	v.Set("mime", ctx.MimeType)
	v.Set("bucket", ctx.Bucket)
	v.Set("key", ctx.Key)
	v.Set("version", ctx.Version)
	v.Set("pipelineId", ctx.PipelineId)
	v.Set("acceptAddrOut", ctx.AcceptAddrOut)
	v.Set("sp", ctx.StyleParam)
	v.Set("token", ctx.Token)
	if ctx.IsGlobal != 0 {
		v.Set("isGlobal", strconv.Itoa(ctx.IsGlobal))
	}
	if ctx.Mode != 0 {
		v.Set("mode", strconv.FormatUint(uint64(ctx.Mode), 10))
		v.Set("force", strconv.FormatUint(uint64(ctx.Force), 10))
	}
	return v.Encode()
}

func DecodeQuery(s string) (fh []byte, fsize int64, ctx *fop.FopCtx, err error) {
	v, err := url.ParseQuery(s)
	if err != nil {
		return
	}
	fh, err = base64.URLEncoding.DecodeString(v.Get("fh"))
	if err != nil {
		err = errors.New("invalid fh")
		return
	}
	fsize, err = strconv.ParseInt(v.Get("fsize"), 10, 64)
	if err != nil {
		err = errors.New("invalid fsize")
		return
	}
	uid, err := strconv.ParseUint(v.Get("uid"), 10, 32)
	if err != nil {
		err = errors.New("invalid uid")
		return
	}

	ctx = &fop.FopCtx{
		RawQuery:      v.Get("cmd"),
		Uid:           uint32(uid),
		URL:           v.Get("url"),
		MimeType:      v.Get("mime"),
		Bucket:        v.Get("bucket"),
		Key:           v.Get("key"),
		Version:       v.Get("version"),
		PipelineId:    v.Get("pipelineId"),
		AcceptAddrOut: v.Get("acceptAddrOut"),
		StyleParam:    v.Get("sp"),
		Token:         v.Get("token"),
	}

	if v.Get("mode") != "" {
		mode, e := strconv.ParseUint(v.Get("mode"), 10, 32)
		if e != nil {
			err = errors.New("invalid mode")
			return
		}
		force, e := strconv.ParseUint(v.Get("force"), 10, 32)
		if e != nil {
			err = errors.New("invalid force")
			return
		}
		ctx.Mode = uint32(mode)
		ctx.Force = uint32(force)
	}
	return
}
