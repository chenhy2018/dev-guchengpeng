package session

import (
	"github.com/bradfitz/gomemcache.20160421/memcache"
	"qbox.us/biz/component/sessions"
)

func McSessionManager(
	config sessions.Config, mc *memcache.Client, keyPrefix string,
) *sessions.SessionManager {

	provider := sessions.NewMcProvider(config, mc, keyPrefix)
	manager := sessions.NewSessionManager(provider)
	return manager
}
