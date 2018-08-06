package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

const (
	StatusWaitPutHeader = 670
	StatusWaitPutData   = 671
)

func (c Client) GidsWithDgid(l rpc.Logger, dgid uint32) (gidInfos []GidInfo, err error) {
	u := fmt.Sprintf("%s/gids/%d", c.Host, dgid)
	err = c.GetClient.Call(l, &gidInfos, u)
	return
}

func (c Client) GidsWithActive(l rpc.Logger, dgid uint32) (gidInfos []GidInfo, err error) {
	u := fmt.Sprintf("%s/gids/%d?withActive=1", c.Host, dgid)
	err = c.GetClient.Call(l, &gidInfos, u)
	return
}

// find next valid raw file header off in [from, ...)
func (c Client) RawFileNext(l rpc.Logger, egid string, from int64) (off int64, err error) {
	u := fmt.Sprintf("%s/rawfile/%s/next/%d", c.Host, egid, from)
	var ret struct {
		Off int64 `json:"off"`
	}
	err = c.PostClient.Call(l, &ret, u)
	if err != nil {
		return
	}
	off = ret.Off
	return
}

// try rawfile from offset
func (c Client) RawFileOk(l rpc.Logger, egid string, off int64) (fsize int64, err error) {
	u := fmt.Sprintf("%s/rawfile/%s/ok/%d", c.Host, egid, off)
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return
	}
	resp, err := c.GetClient.Do(l, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fsize = -1
	if xfsize := resp.Header.Get("X-Fsize"); xfsize != "" { // can get fsize if is ErrStatusWaitPutData
		fsize, err = strconv.ParseInt(xfsize, 10, 64)
		if err != nil {
			return
		}
	}
	if resp.StatusCode/100 != 2 {
		err = rpc.ResponseError(resp)
	}
	return
}

// return file header & data
func (c Client) RawFileAt(l rpc.Logger, egid string, off int64) (rc io.ReadCloser, n int64, err error) {
	u := fmt.Sprintf("%s/rawfile/%s/at/%d", c.Host, egid, off)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	resp, err := c.GetClient.DoWithCrcCheck(l, req)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	rc, n = resp.Body, resp.ContentLength
	return
}

// write data & header
func (c Client) RawFilePutAt(l rpc.Logger, egid string, off int64, r io.Reader, rlen, fsize int64, fid uint64, flag uint8) error {
	if rlen == 0 {
		//r = nil
	}
	u := fmt.Sprintf("%s/rawfile/%s/putat/%d/fsize/%d/fid/%d/flag/%d", c.Host, egid, off, fsize, fid, flag)
	return c.PostClient.CallWith64(l, nil, u, "application/octet-stream", r, rlen)
}
