package filelog

import (
	"github.com/qiniu/filelog"
	"syscall"
)

// ----------------------------------------------------------

type Config struct {
	LogDir     string `json:"logdir"`
	NamePrefix string `json:"lognameprefix"`
	TimeMode   int64  `json:"timemode"` // timeMode 单位sec。保证可以平均切分；不能过小，大于1s; 不能过大，小于1day
	ChunkBits  uint   `json:"chunkbits"`
}

type Logger struct {
	w *filelog.Writer
}

func Open(cfg *Config) (r Logger, err error) {

	if cfg.LogDir == "" {
		return Logger{nil}, syscall.EINVAL
	}

	w, err := filelog.NewWriter(cfg.LogDir, cfg.NamePrefix, cfg.TimeMode, cfg.ChunkBits)
	if err != nil {
		return
	}
	return Logger{w}, nil
}

func (r Logger) Close() (err error) {

	return r.w.Close()
}

func (r Logger) Log(mod string, msg []byte) (err error) {

	msg = append(msg, '\n')
	_, err = r.w.Write(msg)
	return
}

// ----------------------------------------------------------

