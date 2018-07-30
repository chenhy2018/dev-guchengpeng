package discoverd

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"labix.org/v2/mgo/bson"

	"qbox.us/qdiscover/discover"

	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/log.v1"
)

var (
	ErrInvalidAddr   = httputil.NewError(400, "invalid addr")
	ErrInvalidMarker = httputil.NewError(400, "invalid marker")
	ErrInvalidName   = httputil.NewError(400, "invalid name")
	ErrInvalidBson   = httputil.NewError(400, "invalid bson encoding")
	ErrInvalidCfgKey = httputil.NewError(400, "invalid cfg key")
	ErrInvalidState  = httputil.NewError(400, "invalid state")
)

type addrArg struct {
	Addr string `flag:"addr"`
}

type QueryArgs struct {
	Node  string   `json:"node"`
	Name  []string `json:"name"`
	State string   `json:"state"`
}

type listArgs struct {
	Node   string   `json:"node"`
	Name   []string `json:"name"`
	State  string   `json:"state"`
	Marker string   `json:"marker"`
	Limit  int      `json:"limit"`
}

type Discoverd struct {
	s *ServiceManager
}

func New(cfg *Config) (d *Discoverd, err error) {
	s, err := NewServiceManager(cfg)
	if err != nil {
		return nil, err
	}
	return &Discoverd{s}, nil
}

func Run(cfg *Config, bindHost string, mux webroute.Mux) error {
	d, err := New(cfg)
	if err != nil {
		return err
	}

	router := &webroute.Router{Factory: bsonrpc.Factory, Mux: mux}
	router.Register(d)

	return http.ListenAndServe(bindHost, mux)
}

func (d *Discoverd) CmdpbrpcServiceRegister_(args *addrArg, env *rpcutil.Env) error {
	host, port, err := net.SplitHostPort(args.Addr)
	if err != nil {
		return ErrInvalidAddr
	}
	if host == "" || host == "0.0.0.0" {
		if host, _, err = net.SplitHostPort(env.Req.RemoteAddr); err != nil {
			return err
		}
	}
	addr := host + ":" + port

	b, err := ioutil.ReadAll(env.Req.Body)
	if err != nil {
		return err
	}

	var info discover.RegisterArgs
	err = bson.Unmarshal(b, &info)
	if err != nil {
		return ErrInvalidBson
	}
	if info.Name == "" {
		return ErrInvalidName
	}
	log.Debugf("register - host:%s, port:%s, name:%s, attrs:%#v", host, port, info.Name, info.Attrs)

	return d.s.Register(addr, info.Name, info.Attrs)
}

func (d *Discoverd) CmdpbrpcServiceUnregister_(args *addrArg, env *rpcutil.Env) error {
	host, port, err := net.SplitHostPort(args.Addr)
	if err != nil {
		return ErrInvalidAddr
	}
	log.Debugf("unregister - host:%s, port:%s", host, port)

	return d.s.Unregister(args.Addr)
}

func (d *Discoverd) CmdpbrpcServiceEnable_(args *addrArg, env *rpcutil.Env) error {
	host, port, err := net.SplitHostPort(args.Addr)
	if err != nil {
		return ErrInvalidAddr
	}
	log.Debugf("enable - host:%s, port:%s", host, port)

	err = d.s.Enable(args.Addr)
	if err != nil {
		return err
	}
	return d.s.SetLastChange(args.Addr, "enable")
}

func (d *Discoverd) CmdpbrpcServiceDisable_(args *addrArg, env *rpcutil.Env) error {
	host, port, err := net.SplitHostPort(args.Addr)
	if err != nil {
		return ErrInvalidAddr
	}
	log.Debugf("disable - host:%s, port:%s", host, port)
	err = d.s.Disable(args.Addr)
	if err != nil {
		return err
	}
	return d.s.SetLastChange(args.Addr, "disable")
}

func (d *Discoverd) CmdbrpcServiceGet_(args *addrArg, env *rpcutil.Env) (*discover.ServiceInfo, error) {
	host, port, err := net.SplitHostPort(args.Addr)
	if err != nil {
		return nil, ErrInvalidAddr
	}
	log.Debugf("get - host:%s, port:%s", host, port)

	return d.s.Get(args.Addr)
}

func (d *Discoverd) WbrpcServiceCount(args *QueryArgs, env *rpcutil.Env) (*discover.CountRet, error) {
	if !isValidState(args.State) {
		return nil, ErrInvalidState
	}
	args.Node = adjustNode(args.Node, env.Req)
	log.Debugf("count - %#v", *args)

	count, err := d.s.Count(args)
	if err != nil {
		return nil, err
	}
	return &discover.CountRet{count}, nil
}

func (d *Discoverd) WbrpcServiceListall(args *QueryArgs, env *rpcutil.Env) (*discover.ServiceListRet, error) {
	if !isValidState(args.State) {
		return nil, ErrInvalidState
	}
	args.Node = adjustNode(args.Node, env.Req)
	log.Debugf("listall - %#v", args)

	infos, err := d.s.ListAll(args)
	if err != nil {
		return nil, err
	}
	return &discover.ServiceListRet{Items: infos}, nil
}

func (d *Discoverd) WbrpcServiceList(args *listArgs, env *rpcutil.Env) (*discover.ServiceListRet, error) {
	if !isValidMarker(args.Marker) {
		return nil, ErrInvalidMarker

	}
	if !isValidState(args.State) {
		return nil, ErrInvalidState

	}
	args.Node = adjustNode(args.Node, env.Req)
	log.Debugf("list - %#v", *args)

	queryArgs := &QueryArgs{
		Node:  args.Node,
		Name:  args.Name,
		State: args.State,
	}
	infos, marker, err := d.s.List(queryArgs, args.Marker, args.Limit)
	if err != nil {
		return nil, err

	}
	return &discover.ServiceListRet{Marker: marker, Items: infos}, nil

}

func adjustNode(node string, req *http.Request) string {
	if node == "self" {
		node, _, _ = net.SplitHostPort(req.RemoteAddr)
	}
	return node
}

func isValidState(s string) bool {
	if s == "" {
		return true // 空 state 认为是合法的
	}
	_, ok := discover.ValidState(s)
	return ok
}

func isValidMarker(s string) bool {
	if s != "" {
		if _, _, err := net.SplitHostPort(s); err != nil {
			return false
		}
	}
	return true
}

// --------------------------------------------------------
// Cfg
func (d *Discoverd) CmdpbrpcServiceSetcfg_(args *addrArg, env *rpcutil.Env) error {
	host, port, err := net.SplitHostPort(args.Addr)
	if err != nil {
		return ErrInvalidAddr
	}

	b, err := ioutil.ReadAll(env.Req.Body)
	if err != nil {
		return err
	}

	var info discover.CfgArgs
	err = bson.Unmarshal(b, &info)
	if err != nil {
		return ErrInvalidBson
	}
	if strings.TrimSpace(info.Key) == "" {
		return ErrInvalidCfgKey
	}

	log.Debugf("setcfg - host:%s, port:%s, key:%s, value:%#v", host, port, info.Key, info.Value)

	err = d.s.SetCfg(args.Addr, &info)
	if err != nil {
		return err
	}
	return d.s.SetLastChange(args.Addr, "setcfg")
}
