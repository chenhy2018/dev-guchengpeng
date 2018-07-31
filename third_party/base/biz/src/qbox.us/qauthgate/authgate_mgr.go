package qauthgate

import (
	"sync/atomic"
)

// ------------------------------------------------------------------------

func (p *Service) selectServer(host string, server string) (e *routeEntry, err error) {

	if route, ok := p.getRoute(host); ok {
		return route.selectServer(server)
	}
	return nil, ErrServiceNotFound
}

// ------------------------------------------------------------------------

type reloadArgs struct {
	Host string `json:"host"`
}

/*
POST /reload?host=<Host>
*/
func (p *Service) WspReload(args *reloadArgs) (err error) {

	return p.refreshRoute(args.Host)
}

// ------------------------------------------------------------------------

type queryArgs struct {
	Host   string `json:"host"`
	Server string `json:"server"`
}

type queryRet struct {
	Active int32 `json:"conn"`
}

/*
POST /query?host=<Host>&server=<Ip:Port>
*/
func (p *Service) WspQuery(args *queryArgs) (ret queryRet, err error) {

	e, err := p.selectServer(args.Host, args.Server)
	if err != nil {
		return
	}

	ret.Active = atomic.LoadInt32(&e.Active)
	return
}

// ------------------------------------------------------------------------

type enableArgs struct {
	Host   string `json:"host"`
	Server string `json:"server"`
	State  int32  `json:"state"`
}

/*
POST /enable?host=<Host>&server=<Ip:Port>&state=<Enabled>
*/
func (p *Service) WspEnable(args *enableArgs) (err error) {

	var disabled int32
	if args.State == 0 {
		disabled = 1
	}

	e, err := p.selectServer(args.Host, args.Server)
	if err != nil {
		return
	}

	atomic.StoreInt32(&e.Disabled, disabled)
	return
}

// ------------------------------------------------------------------------
