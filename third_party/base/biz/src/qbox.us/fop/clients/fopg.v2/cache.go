// +build go1.5

package fopg

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"

	"qbox.us/cc"
	"qbox.us/fop"

	"github.com/qiniu/xlog.v1"
	"golang.org/x/net/context"
)

func (c *Client) decodeResponse(tpCtx context.Context, resp *http.Response) (r io.ReadCloser, length int64, metas map[string]string, err error) {
	r, length, mime, _, noCdnCache, err := c.ExtractResponse(tpCtx, resp)
	if err != nil {
		return
	}
	metas = make(map[string]string)
	metas["Content-Type"] = mime
	if noCdnCache {
		metas["X-Cdn-Cache-Control"] = "no-cache"
	}
	return
}

func encodeResponse(r io.ReadCloser, length int64, metas map[string]string) (resp *http.Response) {
	resp = &http.Response{ContentLength: length, Body: r, StatusCode: 200, Status: "200 Ok", Close: false}
	resp.Header = make(map[string][]string)
	resp.Header.Set("Content-Type", metas["Content-Type"])
	resp.Header.Set("X-Cdn-Cache-Control", metas["X-Cdn-Cache-Control"])
	resp.Header.Set(fop.OutHeader, fop.OutTypeStream)
	return
}

func generateCacheKey(fh []byte, fsize int64, query, styleParam string) []byte {

	key := "?fh=" + base64.URLEncoding.EncodeToString(fh) + "&fsize=" + strconv.FormatInt(fsize, 10)
	if len(query) > 0 {
		key += "&cmd=" + query
	}
	if styleParam != "" {
		key += "&sp=" + styleParam
	}
	h := sha1.New()
	io.WriteString(h, key)
	return h.Sum(nil)
}

func (c *Client) loadFromCache(xl *xlog.Logger, key []byte) (resp *http.Response, err error) {
	cacheExt := c.loadCache()
	r, length, metas, err := cacheExt.Get(xl, key)
	if err != nil {
		return
	}
	xl.Info("fopg hit cache")
	resp = encodeResponse(r, length, metas)
	return
}

func (c *Client) saveToCache(tpCtx context.Context, key []byte, resp *http.Response) (outresp *http.Response, err error) {
	xl := xlog.FromContextSafe(tpCtx)

	pr1, pw1 := io.Pipe()
	pr2, pw2 := io.Pipe()
	multiWriter := cc.OptimisticMultiWriter(pw1, pw2)
	r, length, metas, err := c.decodeResponse(tpCtx, resp)
	if err != nil {
		xl.Info("decode failed: %+v", resp)
		return
	}
	go func() {
		_, err = io.Copy(multiWriter, r)
		pw1.CloseWithError(err)
		pw2.Close()
		r.Close()
	}()
	go func(xl *xlog.Logger) {
		cacheExt := c.loadCache()
		err := cacheExt.Set(xl, key, pr2, length, metas)
		if err != nil {
			xl.Errorf("IOSystem.Fopproxy: s.dcExt.Set:", err)
		}
		pr2.CloseWithError(err)
	}(xl.Spawn())
	outresp = encodeResponse(pr1, length, metas)
	if resp.Header.Get("X-Resp-Code") != "" {
		outresp.Header.Set("X-Resp-Code", resp.Header.Get("X-Resp-Code"))
	}
	return
}
