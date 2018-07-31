package discover

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/brpc"
)

type Attrs map[string]interface{}

func (m Attrs) ToStruct(st interface{}) error {
	b, err := bson.Marshal(m)
	if err != nil {
		return err
	}
	return bson.Unmarshal(b, st)
}

type ServiceInfo struct {
	Addr       string    `bson:"addr"`
	Name       string    `bson:"name"`
	State      State     `bson:"state"`
	LastUpdate time.Time `bson:"lastUpdate"`
	Attrs      Attrs     `bson:"attrs,omitempty"`
	Cfg        Attrs     `bson:"cfg,omitempty"`
}

func (s *ServiceInfo) String() string {
	return fmt.Sprintf("Addr: %s, Name: %s, State: %s, LastUpdate: %s, Attrs: %v, Cfg: %v",
		s.Addr, s.Name, s.State, s.LastUpdate, s.Attrs, s.Cfg)
}

type QueryArgs struct {
	Node  string
	Name  []string
	State string
}

type RegisterArgs struct {
	Name  string `bson:"name"`
	Attrs Attrs  `bson:"attrs,omitempty"`
}

type CfgArgs struct {
	Key   string      `bson:"key"`
	Value interface{} `bson:"value"`
}

type CountRet struct {
	Count int `bson:"count"`
}

type ServiceListRet struct {
	Marker string         `bson:"marker,omitempty"`
	Items  []*ServiceInfo `bson:"items,omitempty"`
}

type Client struct {
	RetryPolicy
	Conn brpc.Client
}

func New(hosts []string, t http.RoundTripper) *Client {
	conn := brpc.Client{&http.Client{Transport: t}}
	policy := RetryPolicy{Hosts: hosts, ShouldRetry: DefaultShouldRetry}
	return &Client{policy, conn}
}

func (c *Client) ServiceRegister(l rpc.Logger, addr, name string, attrs Attrs) error {
	return c.Run(func(host string) error {
		body := &RegisterArgs{Name: name, Attrs: attrs}
		return c.Conn.CallWithBson(l, nil, host+"/service/register/addr/"+addr, body)
	})
}

func (c *Client) ServiceUnregister(l rpc.Logger, addr string) error {
	return c.Run(func(host string) error {
		return c.Conn.Call(l, nil, host+"/service/unregister/addr/"+addr)
	})
}

func (c *Client) ServiceEnable(l rpc.Logger, addr string) error {
	return c.Run(func(host string) error {
		return c.Conn.Call(l, nil, host+"/service/enable/addr/"+addr)
	})
}

func (c *Client) ServiceDisable(l rpc.Logger, addr string) error {
	return c.Run(func(host string) error {
		return c.Conn.Call(l, nil, host+"/service/disable/addr/"+addr)
	})
}

func (c *Client) ServiceGet(l rpc.Logger, ret interface{}, addr string) error {
	return c.Run(func(host string) error {
		return c.Conn.Call(l, ret, host+"/service/get/addr/"+addr)
	})
}

func (c *Client) ServiceCount(l rpc.Logger, names []string, state State) (count int, err error) {
	err = c.Run(func(host string) error {
		params := map[string][]string{
			"name":  names,
			"state": {string(state)},
		}
		var ret CountRet
		err = c.Conn.CallWithForm(l, &ret, host+"/service/count", params)
		count = ret.Count
		return err
	})
	return
}

func (c *Client) ServiceCountEx(l rpc.Logger, args *QueryArgs) (count int, err error) {
	err = c.Run(func(host string) error {
		params := map[string][]string{
			"node":  {args.Node},
			"name":  args.Name,
			"state": {args.State},
		}
		var ret CountRet
		err = c.Conn.CallWithForm(l, &ret, host+"/service/count", params)
		count = ret.Count
		return err
	})
	return
}

func (c *Client) ServiceList(l rpc.Logger, ret interface{}, names []string, state State, marker string, limit int) error {
	return c.Run(func(host string) error {
		params := map[string][]string{
			"name":   names,
			"state":  {string(state)},
			"marker": {marker},
			"limit":  {strconv.Itoa(limit)},
		}
		return c.Conn.CallWithForm(l, ret, host+"/service/list", params)
	})
}

func (c *Client) ServiceListEx(l rpc.Logger, ret interface{}, args *QueryArgs, marker string, limit int) error {
	return c.Run(func(host string) error {
		params := map[string][]string{
			"node":   {args.Node},
			"name":   args.Name,
			"state":  {args.State},
			"marker": {marker},
			"limit":  {strconv.Itoa(limit)},
		}
		return c.Conn.CallWithForm(l, ret, host+"/service/list", params)
	})
}

func (c *Client) ServiceListAll(l rpc.Logger, ret interface{}, names []string, state State) error {
	return c.Run(func(host string) error {
		params := map[string][]string{
			"name":  names,
			"state": {string(state)},
		}
		return c.Conn.CallWithForm(l, ret, host+"/service/listall", params)
	})
}

func (c *Client) ServiceListAllEx(l rpc.Logger, ret interface{}, args *QueryArgs) error {
	return c.Run(func(host string) error {
		params := map[string][]string{
			"node":  {args.Node},
			"name":  args.Name,
			"state": {args.State},
		}
		return c.Conn.CallWithForm(l, ret, host+"/service/listall", params)
	})
}

// --------------------------------------------------------
// Cfg
func (c *Client) ServiceSetCfg(l rpc.Logger, addr string, cfg *CfgArgs) error {
	return c.Run(func(host string) error {
		return c.Conn.CallWithBson(l, nil, host+"/service/setcfg/addr/"+addr, cfg)
	})
}
