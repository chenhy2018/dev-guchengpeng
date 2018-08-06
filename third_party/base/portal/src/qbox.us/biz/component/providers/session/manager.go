package session

import (
	"qbox.us/biz/component/sessions"
)

func SessionManager(config sessions.Config, provider sessions.SessionProvider) *sessions.SessionManager {
	manager := sessions.NewSessionManager(provider)
	return manager
}
