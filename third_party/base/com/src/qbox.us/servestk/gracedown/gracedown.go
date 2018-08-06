package gracedown

import (
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"qbox.us/servestk"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/osl/signal"
)

// 当剩下 <= left 需要冷却，总冷却时间超过 expire 秒则直接 exit
type TimeoutRule struct {
	Left   int `json:"left"`
	Expire int `json:"expire"`
}

type Config struct {
	TimeoutRules []TimeoutRule `json:"timeout_rules"`
	On           int           `json:"on"` // on == 0 时 gracedown 不起作用
}

type Status struct {
	nprocess int32
	locked   int32
	Config

	beforeExit func()
}

func NewStatus(cfg *Config) *Status {
	s := &Status{Config: *cfg}
	go signal.WaitForInterrupt(s.Down)
	return s
}

func (s *Status) SetBeforeExitFunc(f func()) {
	s.beforeExit = f
}

func (s *Status) Down() {
	if s.On == 0 {
		log.Info("gracedown: off, exit directly")
		if s.beforeExit != nil {
			s.beforeExit()
		}
		os.Exit(0)
		return
	}

	atomic.StoreInt32(&s.locked, 1)

	log.Info("gracedown: cooling down ...")

	var n int32
	var beginTime = time.Now().Unix()
	for {
		nprocess := atomic.LoadInt32(&s.nprocess)
		if nprocess == 0 {
			log.Info("gracedown: cool down ok")
			if s.beforeExit != nil {
				s.beforeExit()
			}
			os.Exit(0)
		}
		if nprocess != n {
			// 记录 nprocess，用来和下一次的 nprocess 比较，如果不一样才打印，以免打印一堆 nprocess。
			n = nprocess
			log.Info("gracedown: nprocess", nprocess)
		}
		if rule, ok := s.isTimeout(nprocess, beginTime); ok {
			log.Warn("gracedown: timeout with rule:", rule, nprocess)
			if s.beforeExit != nil {
				s.beforeExit()
			}
			os.Exit(0)
		}
		time.Sleep(1e9)
	}
}

func (s *Status) isTimeout(nprocess int32, beginTime int64) (TimeoutRule, bool) {
	now := time.Now().Unix()
	for _, rule := range s.TimeoutRules {
		if nprocess <= int32(rule.Left) {
			if now-beginTime > int64(rule.Expire) {
				return rule, true
			}
		}
	}
	return TimeoutRule{}, false
}

func (s *Status) Handler() servestk.Handler {
	return func(w http.ResponseWriter, req *http.Request, f func(w http.ResponseWriter, req *http.Request)) {
		if atomic.LoadInt32(&s.locked) == 1 {
			httputil.Error(w, httputil.ErrGracefulQuit)
			return
		}
		atomic.AddInt32(&s.nprocess, 1)
		defer atomic.AddInt32(&s.nprocess, -1)
		f(w, req)
	}
}
