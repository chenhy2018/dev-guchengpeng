package session

import (
	"labix.org/v2/mgo"
	"qbox.us/biz/component/sessions"
)

func MgoSessionManager(config sessions.Config,
	invoker func(func(*mgo.Collection) error) error,
) *sessions.SessionManager {

	provider := sessions.NewMgoProvider(config, invoker)
	manager := sessions.NewSessionManager(provider)
	return manager
}
