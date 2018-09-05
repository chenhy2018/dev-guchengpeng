package mqafop

import (
	"sync"

	"qbox.us/qmq/mqutil"
	"qbox.us/qmq/mqutil/pmq"
)

// ------------------------------------------------------------------

type Instance struct {
	mq     *pmq.MessageChan
	lockMq *mqutil.MessageChan
	done   chan bool

	allMsgs map[string]*mqutil.Message
	mutex   sync.Mutex

	expires  uint32
	tryTimes uint32
}

// ------------------------------------------------------------------
