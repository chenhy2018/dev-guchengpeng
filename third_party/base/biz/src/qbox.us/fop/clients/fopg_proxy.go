package clients

import (
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
	"qbox.us/errors"
)

type FopgProxy struct {
	fops map[string]bool // 假设做了字节alignment，修改操作是atomic的
	Fopg
}

var ErrUnknownFop = errors.New("unknown fop")

func NewFopgProxy(hosts []string, retryTimes, updateFopsDuration int, transport http.RoundTripper) *FopgProxy {

	p := &FopgProxy{
		fops: make(map[string]bool),
	}
	p.init(hosts, retryTimes, transport)
	if updateFopsDuration > 0 {
		p.Refresh(nil)
		go p.updateFops(updateFopsDuration)
	}
	return p
}

func (p *FopgProxy) Refresh(xl rpc.Logger) (err error) {

	fops, err := p.Fopg.List(xl)
	if err != nil {
		errors.Info(err, "FopsGetter.Refresh").Detail(err).Warn()
		return
	}

	m := make(map[string]bool, len(fops))
	for _, fop := range fops {
		m[fop] = true
	}
	p.fops = m
	return nil
}

func (p *FopgProxy) updateFops(updateFopsDuration int) {

	dur := time.Duration(updateFopsDuration) * 1e9
	for {
		time.Sleep(dur)
		p.Refresh(nil)
	}
}

func (p *FopgProxy) Has(fop string) bool {
	return p.fops[fop]
}
