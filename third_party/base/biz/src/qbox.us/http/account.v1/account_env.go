package account

import (
	"qbox.us/http/account.v2"
	"qbox.us/servend/account"
)

// ---------------------------------------------------------------------------
// [[DEPRECATED]]

type Config struct {
	Acc account.InterfaceEx
}

type Manager struct {
	account.OldParserEx
}

func (p *Manager) InitAccountEx(cfg *Config) {

	p.Account = cfg.Acc
}

// ---------------------------------------------------------------------------
// [[DEPRECATED]]

type Env struct {
	v2.Env
}

type AdminEnv struct {
	v2.AdminEnv
}

type SudoerEnv struct {
	v2.SudoerEnv
}

type ParentUserEnv struct {
	v2.ParentUserEnv
}

// ---------------------------------------------------------------------------
