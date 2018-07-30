// +build go1.5

package fopg

import (
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api/dc"
	"qbox.us/cc/config"
	"qbox.us/dcutil"
)

type ReloadingData struct {
	LbConfig         lb.Config          `json:"lb_conf"`
	FailoverConfig   lb.Config          `json:"failover_conf"`
	LbTrConfig       lb.TransportConfig `json:"lb_tr_conf"`
	FailoverTrConfig lb.TransportConfig `json:"failover_tr_conf"`
	DCConf           dc.Config          `json:"dc_conf"`
	FopResConfig     FailoverConfig     `json:"fop_res_config"`
}

var shouldRetry = func(code int, err error) bool {
	if code == 570 {
		return true
	}
	return lb.ShouldRetry(code, err)
}

func (c *Client) onReload(l rpc.Logger, data []byte) error {
	xl := xlog.NewWith(l.ReqId())

	var conf ReloadingData
	err := config.LoadData(&conf, data)
	if err != nil {
		err = errors.Info(err, "config.LoadData").Detail(err)
		return err
	}

	mac := c.mac
	conf.LbConfig.ShouldRetry = shouldRetry
	conf.FailoverConfig.ShouldRetry = shouldRetry
	t1 := lb.NewTransport(&conf.LbTrConfig)
	clientTr := digest.NewTransport(mac, t1)
	t2 := lb.NewTransport(&conf.FailoverTrConfig)
	failoverTr := digest.NewTransport(mac, t2)
	client := lb.NewWithFailover(&conf.LbConfig, &conf.FailoverConfig, clientTr, failoverTr, nil)

	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.close()
	c.client = client
	if c.dcCache {
		dcTransport := dc.NewTransport(conf.DCConf.DialTimeoutMS, conf.DCConf.RespTimeoutMS, conf.DCConf.TransportPoolSize)
		cache := dc.NewClient(conf.DCConf.Servers, conf.DCConf.TryTimes, dcTransport)
		dcExt := dcutil.NewExt(dc.NewDiskCacheExt(cache))
		c.cacheExt = &dcExt
		c.close = newclose(cache, t1, t2)
	} else {
		c.close = newclose(nil, t1, t2)
	}
	c.fopResClient = NewFailoverClient(&conf.FopResConfig, clientTr, failoverTr, nil)
	c.mutex.Unlock()
	xl.Info("fopg client update success")

	return err
}

func (c *Client) loadClient() *lb.Client {
	c.mutex.RLock()
	ret := c.client
	c.mutex.RUnlock()
	return ret
}

func (c *Client) loadCache() *dcutil.CacheExt {
	c.mutex.RLock()
	ret := c.cacheExt
	c.mutex.RUnlock()
	return ret
}

func (c *Client) loadFopResClient() *FailoverClient {
	c.mutex.RLock()
	ret := c.fopResClient
	c.mutex.RUnlock()
	return ret
}
