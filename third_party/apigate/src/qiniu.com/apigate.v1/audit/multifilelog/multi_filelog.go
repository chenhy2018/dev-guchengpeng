package multifilelog

import (
	"github.com/qiniu/filelog"
	"os"
	"sync"
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
	writers map[string]*filelog.Writer
	cfg     Config
	mutex   sync.RWMutex
}

func Open(cfg *Config) (r *Logger, err error) {

	if cfg.LogDir == "" {
		return nil, syscall.EINVAL
	}

	_, err1 := os.Stat(cfg.LogDir)
	if os.IsNotExist(err1) {
		err = os.MkdirAll(cfg.LogDir, 0700)
		if err != nil {
			return
		}
	}

	writers := make(map[string]*filelog.Writer)
	r = &Logger{writers: writers, cfg: *cfg}
	r.cfg.LogDir += "/"
	return
}

func (r *Logger) Close() (err error) {

	for _, w := range r.writers {
		err1 := w.Close()
		if err1 != nil {
			err = err1
		}
	}
	return
}

func (r *Logger) getWriter(mod string) (w *filelog.Writer, err error) {
	r.mutex.RLock()
	w, ok := r.writers[mod]
	r.mutex.RUnlock()
	if ok {
		return
	}

	dir := r.cfg.LogDir + mod
	syscall.Mkdir(dir, 0777)
	w, err = filelog.NewWriter(dir, r.cfg.NamePrefix, r.cfg.TimeMode, r.cfg.ChunkBits)
	if err != nil {
		return
	}
	r.mutex.Lock()
	r.writers[mod] = w
	r.mutex.Unlock()
	return
}

func (r *Logger) Log(mod string, msg []byte) (err error) {

	w, err := r.getWriter(mod)
	if err != nil {
		return
	}

	msg = append(msg, '\n')
	_, err = w.Write(msg)
	return
}

// ----------------------------------------------------------
