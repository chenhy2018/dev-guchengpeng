package digest_auth

import (
	"qbox.us/errors"
	"qbox.us/servend/account"
)

// ---------------------------------------------------------------------------------------

type Config struct {
	UCHost          string
	RefreshDuration int64
	Account         account.Interface
	DefaultUserId   string
}

// ---------------------------------------------------------------------------------------

func New(cfg *Config) (account.InterfaceEx, error) {

	if acc, ok := cfg.Account.(account.InterfaceEx); ok {
		return acc, nil
	}
	return nil, errors.EINVAL
}

// ---------------------------------------------------------------------------------------
