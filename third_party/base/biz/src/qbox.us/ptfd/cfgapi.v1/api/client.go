package api

import (
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

type DgInfo struct {
	Hosts    []string `json:"hosts"`
	Dgids    []uint32 `json:"dgids"`
	Guid     string   `json:"guid"`
	Writable bool     `json:"writable"`
	Repair   bool     `json:"repair"`
	Idc      string   `json:"idc"`
}

type Client struct {
	lb.Client
}

func shouldRetry(code int, err error) bool {
	if code/100 == 5 {
		return true
	}
	return lb.ShouldRetry(code, err)
}

func New(hosts []string, tr http.RoundTripper) (c *Client, err error) {

	cfg := &lb.Config{
		Hosts:              hosts,
		ShouldRetry:        shouldRetry,
		FailRetryIntervalS: -1,
		TryTimes:           uint32(len(hosts)),
	}
	lc := lb.New(cfg, tr)
	if err != nil {
		return
	}
	c = &Client{Client: *lc}
	return
}

// -----------------------------------------------------------------------------

func (p *Client) Dgs(l rpc.Logger, guid string) (dgs []DgInfo, err error) {

	err = p.CallWithForm(l, &dgs, "/v1/ptfd/dgs", map[string][]string{
		"guid": {guid},
	})
	return
}

func (p *Client) IdcDgs(l rpc.Logger, guid, idc string) (dgs []DgInfo, err error) {

	err = p.CallWithForm(l, &dgs, "/v1/ptfd/dgs", map[string][]string{
		"guid": {guid},
		"idc":  {idc},
	})
	return
}

func (p *Client) Dg(l rpc.Logger, guid, host string) ([]string, []uint32, error) {

	var dgs []DgInfo
	err := p.CallWithForm(l, &dgs, "/v1/ptfd/dgs", map[string][]string{
		"guid": {guid},
		"host": {host},
	})
	if err == nil {
		return dgs[0].Hosts, dgs[0].Dgids, nil
	}
	return nil, nil, err
}

// deprecated: please use HostsIdc
func (p *Client) Hosts(l rpc.Logger, guid string, dgid uint32) (hosts []string, err error) {

	hosts, _, err = p.HostsIdc(l, guid, dgid)
	return
}

func (p *Client) HostsIdc(l rpc.Logger, guid string, dgid uint32) (hosts []string, idc string, err error) {

	var dgs []DgInfo
	err = p.CallWithForm(l, &dgs, "/v1/ptfd/dgs", map[string][]string{
		"guid": {guid},
		"dgid": {strconv.FormatUint(uint64(dgid), 10)},
	})
	if err == nil {
		hosts = dgs[0].Hosts
		idc = dgs[0].Idc
	}
	return
}

func (p *Client) AddDg(l rpc.Logger, guid string, hosts []string, dgids []uint32, idc string) error {

	return p.CallWithJson(l, nil, "/v1/ptfd/adddg", map[string]interface{}{
		"guid":  guid,
		"hosts": hosts,
		"dgids": dgids,
		"idc":   idc,
	})
}

func (p *Client) DeleteDg(l rpc.Logger, guid string, host string) error {

	return p.CallWithJson(l, nil, "/v1/ptfd/deletedg", map[string]interface{}{
		"guid": guid,
		"host": host,
	})
}

func (p *Client) UpdateDg(l rpc.Logger, guid string, host string,
	nhosts []string, ndgids []uint32, writable, hasWritable bool, idc string) error {

	m := map[string]interface{}{
		"guid": guid,
		"host": host,
	}
	if len(nhosts) > 0 {
		m["newhosts"] = nhosts
	}
	if len(ndgids) > 0 {
		m["newdgids"] = ndgids
	}
	if hasWritable {
		m["writable"] = writable
		m["haswritable"] = true
	}
	if idc != "" {
		m["idc"] = idc
	}
	return p.CallWithJson(l, nil, "/v1/ptfd/updatedg", m)
}
