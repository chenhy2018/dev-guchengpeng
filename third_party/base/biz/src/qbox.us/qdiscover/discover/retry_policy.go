package discover

import (
	"errors"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
)

type RetryPolicy struct {
	Hosts       []string
	ShouldRetry func(err error) bool
}

func (r RetryPolicy) Run(call func(host string) error) (err error) {
	if len(r.Hosts) == 0 {
		return errors.New("no host available")
	}
	if r.ShouldRetry == nil {
		r.ShouldRetry = DefaultShouldRetry
	}
	for i, host := range r.Hosts {
		err = call(host)
		if err == nil || !r.ShouldRetry(err) {
			return
		}
		log.Warnf("RetryPolicy: call [%d][%s] failed: %v", i, host, err)
	}
	return
}

// 网络错误重试
func DefaultShouldRetry(err error) bool {
	if _, ok := err.(*rpc.ErrorInfo); ok {
		return false
	}
	return true
}
