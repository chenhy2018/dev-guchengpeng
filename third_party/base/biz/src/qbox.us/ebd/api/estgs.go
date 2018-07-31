package api

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/qiniu/rpc.v1"
	ebdcfg "qbox.us/ebdcfg/api"

	"qbox.us/cc/config"
	"qbox.us/ebdcfg/api/qconf/stgapi"
	"qbox.us/qconf/qconfapi"
)

type estgsClient struct {
	estgs stgapi.Client
	lc    map[string]string // guid+diskId -> host
	mutex sync.RWMutex
}

func newEstgsClient(guid, ebdCfgHost string, reloadMs int) (c *estgsClient, err error) {
	c = &estgsClient{
		estgs: stgapi.Client{qconfapi.New(&qconfapi.Config{
			MasterHosts: []string{ebdCfgHost},
		})},
	}
	reload := &config.ReloadingConfig{
		ConfName:   "estgs.conf",
		RemoteLock: "estgs.conf.lock",
		ReloadMs:   reloadMs,
		RemoteURL:  ebdcfg.DiskListUrl(ebdCfgHost, guid),
	}
	err = config.StartReloading(reload, c.onReload)
	return
}

func (self *estgsClient) Host(l rpc.Logger, guid string, diskId uint32) (host string, err error) {
	key := stgapi.MakeDiskId(guid, diskId)
	self.mutex.RLock()
	host, ok := self.lc[key]
	self.mutex.RUnlock()
	if !ok {
		host, err = self.estgs.Host(l, guid, diskId)
	}
	return
}

func (self *estgsClient) onReload(l rpc.Logger, data []byte) error {
	var disks []*ebdcfg.DiskInfo
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&disks)
	if err != nil {
		return err
	}

	lc := make(map[string]string, len(disks))
	for _, disk := range disks {
		lc[stgapi.MakeDiskId(disk.Guid, disk.DiskId)] = disk.Hosts[0]
	}

	self.mutex.Lock()
	self.lc = lc
	self.mutex.Unlock()
	return nil
}
