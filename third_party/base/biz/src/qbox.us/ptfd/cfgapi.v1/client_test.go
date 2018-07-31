package cfgapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"qbox.us/ptfd/cfgapi.v1/api"

	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

type mockCfg struct {
	dgs     []api.DgInfo
	masters []string
	wg      *sync.WaitGroup
}

type dgsArgs struct {
	Dgid uint32 `json:"dgid"`
	Idc  string `json:"idc"`
}

func (p *mockCfg) WsDgs(args *dgsArgs, env *rpcutil.Env) ([]api.DgInfo, error) {

	xl := xlog.New(env.W, env.Req)
	xl.Infof("Service.WsDgs: args => %+v", args)
	if args.Dgid != 0 {
		for i, dg := range p.dgs {
			for _, dgid := range dg.Dgids {
				if args.Dgid == dgid {
					return p.dgs[i : i+1], nil
				}
			}
		}
		return nil, errors.New("mockCfg: dgid not found")
	}
	if args.Idc != "" {
		var dgs []api.DgInfo
		for _, dg := range p.dgs {
			if dg.Idc == args.Idc {
				dgs = append(dgs, dg)
			}
		}
		return dgs, nil
	}
	if p.wg != nil {
		p.wg.Done()
	}
	return p.dgs, nil
}

func (p *mockCfg) WsMasters(env *rpcutil.Env) (hosts []string, err error) {

	xl := xlog.New(env.W, env.Req)
	xl.Info("Service.WsMasters: request =>", env.Req.URL.Path)
	if p.wg != nil {
		p.wg.Done()
	}
	return p.masters, nil
}

func TestClient(t *testing.T) {

	dgs := []api.DgInfo{
		{Hosts: []string{"abc", "ABC", "AAA"}, Dgids: []uint32{1, 2}, Writable: true, Idc: "nb"},
		{Hosts: []string{"123", "456", "789"}, Dgids: []uint32{3, 4}, Writable: false, Idc: "nb"},
		{Hosts: []string{"def", "DEF"}, Dgids: []uint32{5, 6}, Writable: true, Idc: "nb"},
		{Hosts: []string{"yyy", "zzz"}, Dgids: []uint32{7, 8}, Writable: true, Idc: "hz"},
	}
	mcfg := &mockCfg{}
	router := webroute.Router{
		Factory:       wsrpc.Factory,
		PatternPrefix: "/v1/ptfd",
		Mux:           http.NewServeMux(),
	}
	mux := router.Register(mcfg)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client, err := api.New([]string{ts.URL}, nil)
	assert.NoError(t, err)

	cfg := &Config{
		Hosts:           []string{ts.URL},
		UpdateIntervalS: 1,
	}
	xl := xlog.NewDummy()
	p := &Client{
		Config: *cfg,
		client: client,
	}
	p.updateDgs(xl)

	_, _, _, err = p.HostsIdc(xl, 100)
	assert.Error(t, err)

	_, _, err = p.Actives(xl, "nb")
	assert.Error(t, err)

	mcfg.dgs = dgs
	mcfg.masters = []string{"http://192.168.1.2", "http://192.168.1.3"}

	idxs := []int{0, 1, 2, 0}
	hosts := []string{"123", "456", "789"}
	for i, idx := range idxs {
		rhosts, ridx, ridc, err := p.HostsIdc(xl, 3)
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, idx, ridx, "%v", i)
		assert.Equal(t, hosts, rhosts, "%v", i)
		assert.Equal(t, "nb", ridc)
	}

	idxs = []int{0, 1, 0}
	hosts = []string{"abc", "def"}
	for i, idx := range idxs {
		rhosts, ridx, err := p.Actives(xl, "nb")
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, idx, ridx, "%v", i)
		assert.Equal(t, hosts, rhosts, "%v", i)
	}

	mcfg.dgs[1].Hosts = []string{"123", "456"}
	mcfg.dgs[1].Writable = true

	go p.loopUpdate()

	var wg sync.WaitGroup
	wg.Add(4)
	mcfg.wg = &wg
	wg.Wait()

	idxs = []int{1, 0, 1}
	hosts = []string{"123", "456"}
	for i, idx := range idxs {
		rhosts, ridx, ridc, err := p.HostsIdc(xl, 3)
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, idx, ridx, "%v", i)
		assert.Equal(t, hosts, rhosts, "%v", i)
		assert.Equal(t, "nb", ridc)
	}

	idxs = []int{1, 2, 0, 1}
	hosts = []string{"abc", "123", "def"}
	for i, idx := range idxs {
		rhosts, ridx, err := p.Actives(xl, "nb")
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, idx, ridx, "%v", i)
		assert.Equal(t, hosts, rhosts, "%v", i)
	}

	hosts = []string{"yyy"}
	rhosts, _, err := p.Actives(xl, "hz")
	assert.NoError(t, err)
	assert.Equal(t, hosts, rhosts)

	// test Repair flag
	mcfg.dgs[2].Repair = true
	mcfg.wg = nil
	p.updateDgs(xl)
	idxs = []int{1, 0, 1}
	hosts = []string{"abc", "123"}
	for i, idx := range idxs {
		rhosts, ridx, err := p.Actives(xl, "nb")
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, idx, ridx, "%v", i)
		assert.Equal(t, hosts, rhosts, "%v", i)
	}
}
