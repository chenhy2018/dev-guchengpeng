package bdgetter

import (
	"fmt"
	"sync"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"

	"qbox.us/api/dc"
	"qbox.us/bdgetter/cached"
	"qbox.us/cc/config"
)

const (
	MaxSsdSizeLimit     = 4 * 1024 * 1024
	DefaultSsdSizeLimit = 32 * 1024
)

type DcConfig struct {
	SsdConns     []dc.DCConn       `json:"dc_ssd_conns"`
	SataConns    []dc.DCConn       `json:"dc_sata_conns"`
	SsdTimeouts  dc.TimeoutOptions `json:"dc_ssd_timeouts"`
	SataTimeouts dc.TimeoutOptions `json:"dc_sata_timeouts"`
	SsdSizeLimit int64             `json:"dc_ssd_size_limit"`
	Switch       bool              `json:"dc_cache_get"`
}

// -----------------------------------------------------------------------------

func newCachedInfo(cfg *config.ReloadingConfig) (*cachedInfo, error) {
	ci := new(cachedInfo)
	err := config.StartReloading(cfg, ci.onReload)
	if err != nil {
		return nil, errors.Info(err, "newCached: config.StartReloading").Detail(err)
	}
	return ci, nil
}

type cachedInfo struct {
	ssd, sata    cached.Cached
	ssdCacheSize int64
	mutex        sync.RWMutex
}

func (p *cachedInfo) onReload(l rpc.Logger, data []byte) error {
	var dcConf DcConfig
	err := config.LoadData(&dcConf, data)
	if err != nil {
		return errors.Info(err, "newCached: config.LoadData").Detail(err)
	}

	if dcConf.SsdSizeLimit == 0 {
		dcConf.SsdSizeLimit = DefaultSsdSizeLimit
	}
	if dcConf.SsdSizeLimit > MaxSsdSizeLimit {
		return errors.New(fmt.Sprintf("SsdSizeLimit(%v) is larger then MaxSsdSizeLimit(%v)", dcConf.SsdSizeLimit, MaxSsdSizeLimit))
	}

	var dcSsdClient, dcSataClient cached.Cached
	if len(dcConf.SsdConns) > 0 {
		dcSsdClient = dc.NewWithTimeout(dcConf.SsdConns, &dcConf.SsdTimeouts)
	}
	if len(dcConf.SataConns) > 0 {
		dcSataClient = dc.NewWithTimeout(dcConf.SataConns, &dcConf.SataTimeouts)
	}

	p.mutex.Lock()
	p.ssd, p.sata = dcSsdClient, dcSataClient
	p.ssdCacheSize = dcConf.SsdSizeLimit
	if !dcConf.Switch {
		p.ssd, p.sata = nil, nil
	}
	p.mutex.Unlock()
	return nil
}

func (p *cachedInfo) Get() (ssdCached, sataCached cached.Cached, ssdCacheSize int64) {
	p.mutex.RLock()
	ssdCached, sataCached, ssdCacheSize = p.ssd, p.sata, p.ssdCacheSize
	p.mutex.RUnlock()
	return
}

// -----------------------------------------------------------------------------
