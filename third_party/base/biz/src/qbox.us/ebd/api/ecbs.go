package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"qbox.us/cc/config"
	ebdcfg "qbox.us/ebdcfg/api"
	ecb "qbox.us/ecb/api"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/errors"
)

var ErrEmptyHosts = errors.New("empty hosts")

type rgetClient struct {
	ecbs  *ecb.Client
	mutex sync.RWMutex
	tr    http.RoundTripper
}

func newRgetClient(guid, ebdCfgHost string, reloadMs int, tr http.RoundTripper) (c *rgetClient, err error) {
	c = &rgetClient{tr: tr}
	reload := &config.ReloadingConfig{
		ConfName:   "ecbs.conf",
		RemoteLock: "ecbs.conf.lock",
		ReloadMs:   reloadMs,
		RemoteURL:  ebdcfg.EcbListUrl(ebdCfgHost, guid),
	}
	err = config.StartReloading(reload, c.onReload)
	if err == ErrEmptyHosts {
		err = nil
	}
	return
}

func (self *rgetClient) WriteTo(xl *xlog.Logger,
	w io.Writer, srgi *ecb.StripeRgetInfo) (n int64, err error) {

	self.mutex.RLock()
	ecbs := self.ecbs
	self.mutex.RUnlock()

	if ecbs == nil {
		return 0, ErrEmptyHosts
	}

	rc, err := ecbs.Rget(xl, srgi)
	if err != nil {
		return
	}
	defer rc.Close()
	n, err = io.Copy(w, rc)
	return
}

func (self *rgetClient) onReload(l rpc.Logger, data []byte) error {
	var ecbInfos []*ebdcfg.EcbInfo
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&ecbInfos)
	if err != nil {
		return err
	}

	hosts := make([]string, len(ecbInfos))
	for i, ecb := range ecbInfos {
		hosts[i] = ecb.Hosts[0]
	}

	if len(hosts) == 0 {
		return ErrEmptyHosts
	}
	ecbs, err := ecb.New(hosts, self.tr)
	if err != nil {
		return err
	}

	self.mutex.Lock()
	self.ecbs = &ecbs
	self.mutex.Unlock()
	return nil
}
