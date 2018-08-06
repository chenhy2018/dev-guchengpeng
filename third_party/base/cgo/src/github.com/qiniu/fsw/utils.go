package fsw

import (
	"github.com/qiniu/log.v1"
)

// -------------------------------------------------------------------------

type Config struct {
	OnError        func(err error)
	EventsLimit    int
	WaitingToClose int
}

// -------------------------------------------------------------------------

const (
	DefaultEventsLimit = 1024
)

func logError(err error) {
	log.Error("Watcher failed:", err)
}

func validateCfg(pw **Config) {

	w := *pw
	if w == nil {
		*pw = &Config{logError, DefaultEventsLimit, 0}
		return
	}

	if w.OnError == nil {
		w.OnError = logError
	}

	if w.EventsLimit < 8 {
		w.EventsLimit = DefaultEventsLimit
	}
}

// -------------------------------------------------------------------------
