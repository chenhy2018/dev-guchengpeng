package api

import (
	"bytes"
	"testing"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

func TestFidInfosEncoding(t *testing.T) {

	olds := []FidInfo{
		{
			Fid:   111111111111111,
			Sid:   222222222222222,
			Soff:  33333333,
			Fsize: 444444444444444,
		},
		{
			Fid:   10,
			Sid:   20,
			Soff:  30,
			Fsize: 40,
		},
	}
	buf := bytes.NewBuffer(nil)
	err := EncodeFidInfos(buf, olds)
	assert.NoError(t, err)

	news, err := DecodeFidInfos(buf, 2)
	assert.NoError(t, err)
	assert.Equal(t, olds, news)
}

func TestFwdClient(t *testing.T) {
	ebdMasterSvr := &Service{t: t}
	router := webroute.Router{Mux: http.NewServeMux(), Factory: wsrpc.Factory}
	svr := httptest.NewServer(router.Register(ebdMasterSvr))
	client := FwdClient{rpc.DefaultClient}
	err := client.FwdMigrate(xlog.NewDummy(), svr.URL, 1, &[N+M]uint64{}, &[N+M]uint32{}, Infos)
	assert.NoError(t, err)
	err = client.FwdRecycle(xlog.NewDummy(), svr.URL, 2, &[N+M]uint64{}, &[N+M]uint32{}, Infos)
	assert.NoError(t, err)
}

var Infos = []FidInfo{
	{
		Fid:   111111111111111,
		Sid:   222222222222222,
		Soff:  33333333,
		Fsize: 444444444444444,
	},
	{
		Fid:   10,
		Sid:   20,
		Soff:  30,
		Fsize: 40,
	},
}

type Service struct {
	t *testing.T
}

type fwdMigrateArgs struct {
	Sid   uint64 `json:"sid"`
	Psect string `json:"psect"`
	Crc32 string `json:"crc32s"`
}

func (s *Service) WsFwdMigrate(args *fwdMigrateArgs, env *rpcutil.Env) error {
	req := env.Req
	assert.Equal(s.t, 0, (req.ContentLength-16)% FidInfoLen)
	n := int((req.ContentLength-16) / FidInfoLen)
	infos, err := DecodeFidInfos(req.Body, n)
	assert.NoError(s.t, err)
	assert.Equal(s.t, Infos, infos)
	return nil
}

type fwdRecycleArgs struct {
	Sid   uint64 `json:"sid"`
	Psect string `json:"psect"`
	Crc32 string `json:"crc32s"`
}

func (s *Service) WsFwdRecycle(args *fwdRecycleArgs, env *rpcutil.Env) error {
	req := env.Req
	assert.Equal(s.t, 0, (req.ContentLength-16)% FidInfoLen)
	n := int((req.ContentLength-16) / FidInfoLen)
	infos, err := DecodeFidInfos(req.Body, n)
	assert.NoError(s.t, err)
	assert.Equal(s.t, Infos, infos)
	return nil
}
