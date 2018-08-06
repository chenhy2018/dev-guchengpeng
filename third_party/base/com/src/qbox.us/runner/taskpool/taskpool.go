package taskpool

// -------------------------------------------------------

type Instance struct {
	mq chan func()
}

func New(workerCount int, mailBoxSize int) Instance {

	mq := make(chan func(), mailBoxSize)
	for i := 0; i < workerCount; i++ {
		go func() {
			for {
				task, ok := <-mq
				if !ok {
					break
				}
				task()
			}
		}()
	}
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

func (r Instance) Close() {
	close(r.mq)
}

// -------------------------------------------------------
