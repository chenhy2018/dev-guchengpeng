package account

import (
	"errors"
	"net/http"
)

var ErrNoManager = errors.New("account manager not found")

// ---------------------------------------------------------------------------

type Auther interface {
	Auth(env *Env, req *http.Request) error
}

// ---------------------------------------------------------------------------

type Manager struct {
}

func (p *Manager) Auth(env *Env, req *http.Request) (err error) {
	env.Uid = 123
	return nil
}

// ---------------------------------------------------------------------------

type Env struct {
	W   http.ResponseWriter
	Req *http.Request
	Uid uint
}

func (p *Env) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) (err error) {

	if g, ok := rcvr.(Auther); ok {
		return g.Auth(p, req)
	}
	return ErrNoManager
}

func (p *Env) CloseEnv() {
}

// ---------------------------------------------------------------------------
