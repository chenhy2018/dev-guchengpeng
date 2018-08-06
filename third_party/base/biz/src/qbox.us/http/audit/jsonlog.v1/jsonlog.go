package jsonlog

import (
	"github.com/qiniu/http/audit/jsonlog"
	"github.com/qiniu/reliable/log"
	"qbox.us/http/audit/jsonlog.v2"
	"qbox.us/servend/account"
)

// ----------------------------------------------------------

type Config v2.Config

func Open(module string, cfg *Config, acc account.InterfaceEx) (al *jsonlog.Logger, logf *log.Logger, err error) {

	return v2.Open(module, (*v2.Config)(cfg), account.OldParserEx{acc})
}

// ----------------------------------------------------------
