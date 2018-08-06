package mgo2

import (
	"strings"
)

import (
	"github.com/qiniu/log.v1"
	"labix.org/v2/mgo"
)

// ------------------------------------------------------------------------

var g_modes = map[string]int{
	"eventual":  0,
	"monotonic": 1,
	"mono":      1,
	"strong":    2,
}

func SetMode(s *mgo.Session, modeFriendly string, refresh bool) {

	mode, ok := g_modes[strings.ToLower(modeFriendly)]
	if !ok {
		log.Panic("invalid mgo mode")
	}
	switch mode {
	case 0:
		s.SetMode(mgo.Eventual, refresh)
	case 1:
		s.SetMode(mgo.Monotonic, refresh)
	case 2:
		s.SetMode(mgo.Strong, refresh)
	}
}

// ------------------------------------------------------------------------
