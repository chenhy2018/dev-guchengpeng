package digest_auth

import (
	qaccount "qbox.us/account"
	sacc "qbox.us/servend/account"
	digest_auth "qbox.us/servend/digest_auth/admin"
)

// ---------------------------------------------------------------------------

type Config digest_auth.Config

func New(cfg *Config) sacc.InterfaceEx {

	acc1 := qaccount.Account{}
	return digest_auth.New((*digest_auth.Config)(cfg), acc1)
}

func NewParser(cfg *Config) sacc.OldParserEx {

	return sacc.OldParserEx{New(cfg)}
}

// ---------------------------------------------------------------------------
