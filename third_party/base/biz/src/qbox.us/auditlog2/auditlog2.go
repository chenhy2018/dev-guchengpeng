package auditlog2

import (
	"qbox.us/audit/largefile"
	"qbox.us/audit/logh"
	"qbox.us/servend/account"
	alogh "qbox.us/servend/account/logh"
)

// ----------------------------------------------------------

type Config struct {
	LogFile   string
	ChunkBits byte
	NoXlog    byte
	Reserved  int16
	BodyLimit int32
}

type Instance struct {
	*logh.Instance
	*largefile.Logger
}

// ----------------------------------------------------------
// NOTE: 这个包已经迁移到 qbox.us/http/audit/jsonlog

func Open(serviceType string, cfg *Config, acc account.Interface) (r Instance, err error) {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}
	if cfg.ChunkBits == 0 {
		cfg.ChunkBits = 22
	}

	logger, err := largefile.Open(cfg.LogFile, uint(cfg.ChunkBits))
	if err != nil {
		return
	}

	var dec logh.Decoder
	if acc != nil {
		dec = alogh.Decoder{Account: acc}
	}
	handler := logh.NewEx(serviceType, logger, dec, int(cfg.BodyLimit), cfg.NoXlog == 0)
	return Instance{handler, logger}, nil
}

func OpenExt(serviceType string, cfg *Config, acc account.InterfaceEx) (r Instance, err error) {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}
	if cfg.ChunkBits == 0 {
		cfg.ChunkBits = 22
	}

	logger, err := largefile.Open(cfg.LogFile, uint(cfg.ChunkBits))
	if err != nil {
		return
	}

	var dec logh.Decoder
	if acc != nil {
		dec = alogh.ExtDecoder{Account: acc}
	}
	handler := logh.NewEx(serviceType, logger, dec, int(cfg.BodyLimit), cfg.NoXlog == 0)
	return Instance{handler, logger}, nil
}

// ----------------------------------------------------------
