package api

import (
	"io"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	ebdtypes "qbox.us/ebd/api/types"
	"qbox.us/fh/fhver"
	"qbox.us/multiebd"
	"qbox.us/pfd/api/types"
	pfdcfg "qbox.us/pfdcfg/api"
	"qbox.us/pfdtracker/stater"
)

type Getter interface {
	Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error)
	GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error)
}

type Client struct {
	s   stater.Stater
	pfd Getter
	ebd multiebd.GetterChooser
}

func New(s stater.Stater, pfd Getter, ebd Getter) *Client {
	return NewWithChooser(s, pfd, multiebd.NewSingleEbd(ebd))
}

func NewWithChooser(s stater.Stater, pfd Getter, ebd multiebd.GetterChooser) *Client {
	return &Client{s: s, pfd: pfd, ebd: ebd}
}

func (self *Client) Get(l rpc.Logger,
	fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {

	xl := xlog.NewWith(l)
	fhi, err := ebdtypes.DecodeFh(fh)
	if err != nil {
		return
	}
	egid := types.EncodeGid(fhi.Gid)
	group, _, isECed, err := self.s.StateWithGroup(l, egid)
	if err != nil {
		return
	}
	if !isECed {
		rc, fsize, err = self.pfd.Get(l, fh, from, to)
		if err == nil {
			return
		}
		if code := httputil.DetectCode(err); code != 612 {
			return
		}
		// 再次确认是否EC
		group1, _, isECed, err1 := self.s.StateWithGroup(l, egid)
		if err1 != nil {
			xl.Errorf("pfd.Get 612, state got err: %v\n", err1)
			return
		}
		group = group1
		if !isECed {
			_, isECed, err1 := self.s.ForceUpdate(l, egid)
			if err1 != nil || !isECed {
				xl.Errorf("pfd.Get 612, ForceUpdate got err: %v, isECed: %v\n", err1, isECed)
				return
			}
		}
		xl.Debug("pfd.Get 612, and state is refreshed and ECed")
	}
	if fhver.FhVer(fh) == fhver.FhSha1bdEbd {
		rc, fsize, err = self.ebd.Choose(group).Get(l, fh, from, to)
		return
	}

	// ec状态，优先去ebd读
	// 如果失败了，再去pfd读
	rc, fsize, err = self.ebd.Choose(group).Get(l, fh, from, to)
	if err != nil {
		code := httputil.DetectCode(err)
		if code == 573 || code == 499 {
			return
		}
		xl.Infof("ebd.Get error(try pfd): %v\n", errors.Detail(err))
		rc, fsize, err = self.pfd.Get(l, fh, from, to)
	}
	return
}

func (self *Client) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	xl := xlog.NewWith(l)
	fhi, err := ebdtypes.DecodeFh(fh)
	if err != nil {
		return
	}
	egid := types.EncodeGid(fhi.Gid)

	group, _, isECed, err := self.s.StateWithGroup(xl, egid)
	if err != nil {
		xl.Errorf("state egid(%v) failed: %v\n", egid, err)
		return
	}
	if !isECed {
		return self.pfd.GetType(l, fh)
	}
	typ, err = self.ebd.Choose(group).GetType(l, fh)
	if err != nil { // ebd 读取失败要回 pfd 读
		xl.Errorf("ebd.GetType error", errors.Detail(err))
		typ, err = self.pfd.GetType(l, fh)
	}
	return
}
