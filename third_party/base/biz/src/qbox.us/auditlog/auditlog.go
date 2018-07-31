package auditlog

import (
	"qbox.us/audit"
	"qbox.us/audit/logh"
	"qbox.us/servend/account"
	alogh "qbox.us/servend/account/logh"
)

// ----------------------------------------------------------

type Config struct {
	Logger    logh.Logger
	Hosts     []string
	BodyLimit int32
}

// ----------------------------------------------------------
// NOTE: 这个包已经迁移到 qbox.us/http/audit/jsonlog

func New(serviceType string, cfg *Config, acc account.Interface) *logh.Instance {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}
	if cfg.Logger == nil {
		if len(cfg.Hosts) == 0 {
			panic("NewAuditLog failed: invalid config")
		}
		cfg.Logger = audit.New(cfg.Hosts)
	}

	dec := alogh.Decoder{Account: acc}
	return logh.New(serviceType, cfg.Logger, dec, int(cfg.BodyLimit))
}

// ----------------------------------------------------------

func NewExt(serviceType string, cfg *Config, acc account.InterfaceEx) *logh.Instance {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}
	if cfg.Logger == nil {
		if len(cfg.Hosts) == 0 {
			panic("NewAuditLog failed: invalid config")
		}
		cfg.Logger = audit.New(cfg.Hosts)
	}

	dec := alogh.ExtDecoder{Account: acc}
	return logh.New(serviceType, cfg.Logger, dec, int(cfg.BodyLimit))
}

// ----------------------------------------------------------
