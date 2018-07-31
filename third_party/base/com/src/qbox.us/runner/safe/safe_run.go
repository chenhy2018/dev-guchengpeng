package safe

import (
	"github.com/qiniu/log.v1"
	"runtime/debug"
)

// -------------------------------------------------------

func Run(task func()) {

	defer func() {
		err := recover()
		if err != nil {
			log.Printf("WARN: panic fired in %v.panic - %v\n", task, err)
			log.Println(string(debug.Stack()))
		}
	}()

	task()
}

// -------------------------------------------------------
