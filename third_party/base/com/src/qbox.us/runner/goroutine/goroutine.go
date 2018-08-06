package goroutine

// -------------------------------------------------------

type Instance struct{}

func (r Instance) Run(task func()) {
	go task()
}

// -------------------------------------------------------
