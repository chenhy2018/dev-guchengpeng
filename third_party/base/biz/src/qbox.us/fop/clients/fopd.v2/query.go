package fopd

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
	v.Set("mime", ctx.MimeType)
	v.Set("cmdName", ctx.CmdName)
	v.Set("cmd", ctx.RawQuery)
	v.Set("cmds", ctx.RawQueries)
	v.Set("url", ctx.URL)
	v.Set("sp", ctx.StyleParam)
	v.Set("token", ctx.Token)
	v.Set("bucket", ctx.Bucket)
	v.Set("key", ctx.Key)
	v.Set("version", ctx.Version)
	v.Set("uid", strconv.FormatUint(uint64(ctx.Uid), 10))
	v.Set("mode", strconv.FormatUint(uint64(ctx.Mode), 10))
	if ctx.IsGlobal != 0 {
		v.Set("isGlobal", strconv.Itoa(ctx.IsGlobal))
	}
	if ctx.SourceURL != "" {
		v.Set("src", base64.URLEncoding.EncodeToString([]byte(ctx.SourceURL)))
	}
	if ctx.OutType != "" {
		v.Set("outType", string(ctx.OutType))
		v.Set("outRSBucket", ctx.OutRSBucket)
		v.Set("outRSKey", ctx.OutRSKey)
		v.Set("outRSDeleteAfterDays", strconv.Itoa(int(ctx.OutRSDeleteAfterDays)))
		v.Set("outDCKey", base64.URLEncoding.EncodeToString(ctx.OutDCKey)) // dc key 需要 base64 编码
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

	ctx = &fop.FopCtx{
		MimeType:   v.Get("mime"),
		CmdName:    v.Get("cmdName"),
		RawQuery:   v.Get("cmd"),
		RawQueries: v.Get("cmds"),
		URL:        v.Get("url"),
		StyleParam: v.Get("sp"),
		Token:      v.Get("token"),
		Bucket:     v.Get("bucket"),
		Key:        v.Get("key"),
		Version:    v.Get("version"),
		Fh:         fh,
	}

	if v.Get("isGlobal") != "" {
		ctx.IsGlobal, err = strconv.Atoi(v.Get("isGlobal"))
		if err != nil {
			err = errors.New("invalid isGlobal arg")
			return
		}
	}

	if v.Get("src") != "" {
		b, e := base64.URLEncoding.DecodeString(v.Get("src"))
		if e != nil {
			err = errors.New("invalid src")
			return
		}
		ctx.SourceURL = string(b)
	}

	uidStr := v.Get("uid")
	if uidStr != "" {
		uid, e := strconv.ParseUint(uidStr, 10, 32)
		if e != nil {
			err = errors.New("invalid uid")
			return
		}
		ctx.Uid = uint32(uid)
	}

	modeStr := v.Get("mode")
	if modeStr != "" {
		mode, e := strconv.ParseUint(v.Get("mode"), 10, 32)
		if e != nil {
			err = errors.New("invalid mode")
			return
		}
		ctx.Mode = uint32(mode)
	}

	outTypeStr := v.Get("outType")
	if outTypeStr != "" {
		ctx.OutType = outTypeStr
		ctx.OutRSBucket = v.Get("outRSBucket")
		ctx.OutRSKey = v.Get("outRSKey")
		deleteAfterDays := v.Get("outRSDeleteAfterDays")
		if deleteAfterDays != "" {
			d, e := strconv.Atoi(deleteAfterDays)
			if e != nil {
				err = errors.New("invalid deleteAfterDays")
				return
			}
			ctx.OutRSDeleteAfterDays = uint32(d)
		}
		ctx.OutDCKey, err = base64.URLEncoding.DecodeString(v.Get("outDCKey"))
		if err != nil {
			err = errors.New("invalid outDCKey")
			return
		}
	}

	return
}
