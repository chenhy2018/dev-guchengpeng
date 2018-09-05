package respool

// -------------------------------------------------------

type Instance struct {
	mq chan func(res interface{})
}

func New(resources []interface{}, mailBoxSize int) Instance {

	mq := make(chan func(interface{}), mailBoxSize)
	for _, res := range resources {
		go func(res interface{}) {
			for {
				task := <-mq
				task(res)
			}
		}(res)
	}
	return Instance{mq}
}

func (r Instance) Run(task func(res interface{})) {

	r.mq <- task
}

func (r Instance) TryRun(task func(res interface{})) bool {

	select {
	case r.mq <- task:
		return true
	default:
	}
	return false
}

// -------------------------------------------------------
