package foptest

import (
	"github.com/qiniu/xlog.v1"
	"qbox.us/dcutil"
	"qbox.us/fop"
	"qbox.us/fop/localcache"
	"qbox.us/mockacc"
	"qbox.us/servend/account"
)

type env struct {
	xl  *xlog.Logger
	acc account.InterfaceEx
}

func NewEnv() fop.Env {

	return &env{
		xl:  xlog.NewDummy(),
		acc: mockacc.New(mockacc.GetSa()),
	}
}

func (p *env) Xlogger() *xlog.Logger {

	return p.xl
}

func (p *env) Xdc() dcutil.Interface {

	return nil
}

func (p *env) TempDir() string {

	return ""
}

func (p *env) Acc() account.InterfaceEx {
	return p.acc
}

func (p *env) LocalCache() fop.LocalCache {
	lc, _ := localcache.NewLocalCache(nil)
	return lc
}
