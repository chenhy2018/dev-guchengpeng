package taskpool2

import (
	"github.com/qiniu/osl/sync"
)

// -------------------------------------------------------

type Instance struct {
	mq chan func()
}

func New(workerCount int, mailBoxSize int) Instance {

	mq := make(chan func(), mailBoxSize)
	sema := sync.NewSemaphore(workerCount)
	go func() {
		for {
			sema.Lock()
			task := <-mq
			go func() {
				task()
				sema.Unlock()
			}()
		}
	}()
	return Instance{mq}
}

func (r Instance) Run(task func()) {

	r.mq <- task
}

func (r Instance) TryRun(task func()) bool {

	select {
	case r.mq <- task:
		return true
	default:
	}
	return false
}

// -------------------------------------------------------
