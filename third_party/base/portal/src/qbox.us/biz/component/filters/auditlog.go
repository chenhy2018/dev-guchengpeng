package filters

import (
	"net/http"
	"time"

	qbytes "github.com/qiniu/bytes"
	"github.com/teapots/teapot"
	"qbox.us/biz/services.v2/account"
	"qbox.us/http/audit/teapotlog"
)

type AuditLogConfig struct {
	ModuleName string
	LogDir     string
	ChunkBits  uint
	BodyLimit  int
}

func UseAuditLog(tea *teapot.Teapot, auditLogCfg AuditLogConfig) interface{} {

	config := teapotlog.Config{
		LogDir:    auditLogCfg.LogDir,
		ChunkBits: auditLogCfg.ChunkBits,
		BodyLimit: auditLogCfg.BodyLimit,
	}

	al, err := teapotlog.Open(auditLogCfg.ModuleName, &config)
	if err != nil {
		tea.Logger().Errorf("create audit log failed. err:%v cfg:%v", err, config)
		panic("create audit log failed.")
	}

	tea.Logger().Infof("open teapot log success. cfg:%v", config)

	return func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {

		body := qbytes.NewWriter(make([]byte, al.Limit))

		var userInfo *account.UserInfo
		ctx.Find(&userInfo, "")

		startTime := time.Now().UnixNano()
		w1 := &teapotlog.ResponseWriter{
			ResponseWriter: rw.(teapot.ResponseWriter),
			Body:           body,
			Code:           200,
			Xlog:           al.Xlog,
			Mod:            al.Mod,
			StartT:         startTime,
		}

		ctx.ProvideAs(w1, (*http.ResponseWriter)(nil))
		ctx.Next()

		al.Log(userInfo, w1, req)
	}
}
