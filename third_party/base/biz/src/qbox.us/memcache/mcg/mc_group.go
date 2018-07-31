package mcg

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/log.v1"
)

// ----------------------------------------------------------------------------

type Node struct {
	Keys []string `json:"keys"`
	Host string   `json:"host"`
}

func New(nodes []Node) (p *memcache.Client, err error) {

	selector, err := newServerSelector(nodes)
	if err != nil {
		log.Error("mcg.New failed", nodes, err)
		return
	}

	return memcache.NewFromSelector(selector), nil
}

// ----------------------------------------------------------------------------
