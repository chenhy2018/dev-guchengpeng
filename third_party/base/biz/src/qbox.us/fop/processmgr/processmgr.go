package processmgr

import (
	"os"
	"sync"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/xlog.v1"
)

type container struct {
	p *os.Process
	q chan bool
}

var counter = make(map[*xlog.Logger]*container)
var lock sync.RWMutex

func processMon(xl *xlog.Logger, quit chan bool, ctx context.Context) {

	select {
	case <-quit:

	case <-ctx.Done():
		lock.RLock()
		con := counter[xl]
		lock.RUnlock()
		if con != nil {
			con.p.Kill()
			xl.Infof("Pid: %d killed", con.p.Pid)
		}
	}

}

func Add(xl *xlog.Logger, p *os.Process, ctx context.Context) {
	quit := make(chan bool, 1)
	lock.Lock()
	counter[xl] = &container{p, quit}
	lock.Unlock()
	go processMon(xl, quit, ctx)
}

func Del(xl *xlog.Logger) {
	lock.Lock()
	quit := counter[xl].q
	delete(counter, xl)
	lock.Unlock()
	quit <- true
}
