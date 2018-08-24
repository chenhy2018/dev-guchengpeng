package gateapp

import (
	"io"

	"qiniu.com/apigate.v1/audit/jsonlog"
	"qiniu.com/apigate.v1/audit/multifilelog"
)

// --------------------------------------------------------------------

type AuditConfig struct {
	MultiLog   multifilelog.Config `json:"multilog"`
	BodyLimit  int                 `json:"bodylimit"`
	CloseAudit bool                `json:"closeaudit"`
}

func initAuditLog(cfg *AuditConfig) (al *jsonlog.Logger, closer io.Closer, err error) {

	if cfg.CloseAudit {
		wc := nopWriteCloser{}

		closer = wc
		al = jsonlog.New(wc, cfg.BodyLimit)
		return
	}

	w, err := multifilelog.Open(&cfg.MultiLog)
	if err != nil {
		return
	}
	closer = w

	al = jsonlog.New(w, cfg.BodyLimit)
	return
}

type nopWriteCloser struct {
}

func (c nopWriteCloser) Close() error {
	return nil
}

func (c nopWriteCloser) Log(mod string, msg []byte) error {
	return nil
}

// --------------------------------------------------------------------
