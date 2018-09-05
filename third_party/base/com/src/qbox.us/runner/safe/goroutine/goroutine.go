package goroutine

import (
	"qbox.us/runner/safe"
)

// -------------------------------------------------------

type Instance struct{}

func (r Instance) Run(task func()) {
	go safe.Run(task)
}

// -------------------------------------------------------
