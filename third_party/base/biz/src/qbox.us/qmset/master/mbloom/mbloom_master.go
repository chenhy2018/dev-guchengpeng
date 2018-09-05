package mbloom

import (
	"encoding/base64"
	"github.com/qiniu/errors"
	account "qbox.us/http/account.v2"
	. "qbox.us/qmset/proto"
	"syscall"
)

// ------------------------------------------------------------------------

type BaddArgs struct {
	Id     string   `json:"c"`
	Values []string `json:"v"`
}

func getValues(vals []string) (ret [][]byte, err error) {

	ret = make([][]byte, len(vals))
	for i, val := range vals {
		switch len(val) % 4 {
		case 2:
			val += "=="
		case 3:
			val += "="
		}
		b, err2 := base64.URLEncoding.DecodeString(val)
		if err2 != nil {
			err = ErrValueNotUrlsafeBase64
			return
		}
		ret[i] = b
	}
	return
}

//
// POST /badd?c=<GrpName>&v=<Value1>&v=<Value2>&v=...
//
func WspBadd(grps map[string]Flipper, args *BaddArgs, env *account.AdminEnv) (err error) {

	if grp, ok := grps[args.Id]; ok {
		if mbloomg, ok := grp.(*Filter); ok {
			bvals, err2 := getValues(args.Values)
			if err2 != nil {
				err = errors.Info(err2, "/badd", args.Id, args.Values).Detail(err2)
				return
			}
			mbloomg.Add(bvals)
			return nil
		}
		return ErrNotMbloom
	}
	return syscall.ENOENT
}

// ------------------------------------------------------------------------

type BchkArgs struct {
	Id     string   `json:"c"`
	Values []string `json:"v"`
}

//
// POST /bchk?c=<GrpName>&v=<Value1>&v=<Value2>&v=...
//
func WspBchk(grps map[string]Flipper, args *BchkArgs, env *account.AdminEnv) (idxs []int, err error) {

	if grp, ok := grps[args.Id]; ok {
		if mbloomg, ok := grp.(*Filter); ok {
			bvals, err2 := getValues(args.Values)
			if err2 != nil {
				err = errors.Info(err2, "/bchk", args.Id, args.Values).Detail(err2)
				return
			}
			return mbloomg.Exists(bvals), nil
		}
		err = ErrNotMbloom
		return
	}
	err = syscall.ENOENT
	return
}

// ------------------------------------------------------------------------
