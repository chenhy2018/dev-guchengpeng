package discoverd_test

import (
	"net/http"
	"testing"

	"qbox.us/mgo3"
	"qbox.us/qdiscover/discover"
	. "qbox.us/qdiscover/discoverd"

	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/mockhttp.v1"
	"github.com/stretchr/testify/assert"
)

var (
	bindHost    = "discoverd"
	clientHosts = []string{"xxx://invalidhost", "http://" + bindHost}
)

func init() {
	log.SetOutputLevel(0)
	runService()
}

func runService() {
	cfg := &Config{
		ServiceColl: mgo3.Config{
			Host: "localhost",
			DB:   "discoversvrTest",
			Coll: "service",
			Mode: "strong",
		},
		HeartbeatMissSecs: heartbeatMissSecs,
	}
	cleanTestEnv(cfg)

	d, err := New(cfg)
	if err != nil {
		log.Fatal("New failed:", err)
	}

	mux := http.NewServeMux()
	router := &webroute.Router{Factory: bsonrpc.Factory, Mux: mux}
	router.Register(d)

	mockhttp.Bind(bindHost, mux)
}

func checkServiceStatus(t *testing.T, who string, dapi *discover.Client, addr string, state discover.State, attrs discover.Attrs) {
	var info discover.ServiceInfo
	err := dapi.ServiceGet(nil, &info, addr)
	assert.NoError(t, err, who)
	assert.Equal(t, info.State, state, who)
	assert.Equal(t, info.Attrs, attrs, who)
}

func TestDiscoverd_Service(t *testing.T) {
	var err error
	dapi := discover.New(clientHosts, nil)

	name := "image"
	addr := "127.0.0.1:9000"
	attrs := discover.Attrs{
		"cmds": []interface{}{"imageView", "imageMogr", "imageInfo"},
	}

	err = dapi.ServiceRegister(nil, "addr", name, nil)
	assert.Equal(t, err.Error(), ErrInvalidAddr.Error(), "Register")
	err = dapi.ServiceRegister(nil, addr, "", nil)
	assert.Equal(t, err.Error(), ErrInvalidName.Error(), "Register")

	err = dapi.ServiceRegister(nil, addr, name, attrs)
	assert.NoError(t, err, "ServiceRegister")
	checkServiceStatus(t, "ServiceRegister", dapi, addr, discover.StatePending, attrs)

	err = dapi.ServiceEnable(nil, addr)
	assert.NoError(t, err, "ServiceEnable")
	checkServiceStatus(t, "ServiceEnable", dapi, addr, discover.StateEnabled, attrs)

	err = dapi.ServiceDisable(nil, addr)
	assert.NoError(t, err, "ServiceDisable")
	checkServiceStatus(t, "ServiceDisable", dapi, addr, discover.StateDisabled, attrs)

	err = dapi.ServiceUnregister(nil, addr)
	assert.NoError(t, err, "ServiceUnregister")

	info := discover.ServiceInfo{}
	err = dapi.ServiceGet(nil, &info, addr)
	assert.Equal(t, err.Error(), ErrNoSuchEntry.Error())

	err = dapi.ServiceList(nil, nil, nil, discover.StatePending, "invalidMarker", 0)
	assert.Equal(t, err.Error(), ErrInvalidMarker.Error())

	name2 := "image"
	addr2 := "127.0.0.1:9900"

	// 使用 ":9000" 和 "0.0.0.0:9900" 注册来测试自动获取 RemoteAddr。
	err = dapi.ServiceRegister(nil, ":9000", name, attrs)
	assert.NoError(t, err, "ServiceRegister")
	err = dapi.ServiceRegister(nil, "0.0.0.0:9900", name2, attrs)
	assert.NoError(t, err, "ServiceRegister")

	count, err := dapi.ServiceCount(nil, nil, discover.StatePending)
	assert.NoError(t, err, "ServiceCount")
	assert.Equal(t, count, 2, "ServiceCount")

	// list all
	listRet := discover.ServiceListRet{}
	err = dapi.ServiceListAll(nil, &listRet, nil, discover.StatePending)
	assert.NoError(t, err, "ServiceList")
	assert.Equal(t, len(listRet.Items), 2, "ServiceList")
	log.Infof("listRet: %#v", listRet)

	// list all with node
	args := &discover.QueryArgs{
		Node:  "127.0.0.1",
		State: string(discover.StatePending),
	}
	listRet = discover.ServiceListRet{}
	err = dapi.ServiceListAllEx(nil, &listRet, args)
	assert.NoError(t, err, "ServiceList")
	assert.Equal(t, len(listRet.Items), 2, "ServiceList")
	log.Infof("listRet: %#v", listRet)

	// list one by one
	listRet1 := discover.ServiceListRet{}
	err = dapi.ServiceList(nil, &listRet1, []string{"image"}, "", "", 1)
	assert.NoError(t, err, "ServiceList")
	assert.Equal(t, listRet1.Marker, addr2, "ServiceList")
	assert.Equal(t, len(listRet1.Items), 1, "ServiceList")
	assert.Equal(t, listRet1.Items[0].State, discover.StatePending, "ServiceList")
	assert.Equal(t, listRet1.Items[0].Attrs, attrs, "ServiceList")

	listRet2 := discover.ServiceListRet{}
	err = dapi.ServiceList(nil, &listRet2, []string{}, "", listRet1.Marker, 1)
	assert.NoError(t, err, "ServiceList")
	assert.Equal(t, listRet2.Marker, "", "ServiceList")
	assert.Equal(t, len(listRet2.Items), 1, "ServiceList")
	assert.Equal(t, listRet2.Items[0].State, discover.StatePending, "ServiceList")
	assert.Equal(t, listRet2.Items[0].Attrs, attrs, "ServiceList")
}

func TestDiscoverd_Service_Cfg(t *testing.T) {
	var err error
	var cfg *discover.CfgArgs
	dapi := discover.New(clientHosts, nil)

	name := "fopagentttttt"
	addr := "127.0.0.2:22222"
	attrs := discover.Attrs{
		"cmds": []interface{}{"imageView", "imageMogr", "imageInfo"},
	}

	// SetCfg should be failed when no such addr
	cfg = &discover.CfgArgs{Key: "xxxxx", Value: 0}
	err = dapi.ServiceSetCfg(nil, addr, cfg)
	assert.Equal(t, err.Error(), ErrNoSuchEntry.Error(), "SetCfg non-existed Addr")

	err = dapi.ServiceRegister(nil, addr, name, attrs)
	assert.NoError(t, err, "ServiceRegister")
	checkServiceStatus(t, "ServiceRegister", dapi, addr, discover.StatePending, attrs)

	// SetCfg should be failed when key is empty
	cfg = &discover.CfgArgs{Key: "   ", Value: 1}
	err = dapi.ServiceSetCfg(nil, addr, cfg)
	assert.Equal(t, err.Error(), ErrInvalidCfgKey.Error(), "SetCfg non-existed Addr")

	type Tcfg struct {
		T interface{} `json:"ccc" bson:"ccc"`
	}
	var result Tcfg
	var info discover.ServiceInfo

	// SetCfg int
	cfg = &discover.CfgArgs{Key: "ccc", Value: 1}
	err = dapi.ServiceSetCfg(nil, addr, cfg)
	assert.NoError(t, err, "SetCfg int")

	err = dapi.ServiceGet(nil, &info, addr)
	assert.NoError(t, err, "ServiceGet")
	err = info.Cfg.ToStruct(&result)
	assert.NoError(t, err, "ToStruct")
	assert.Equal(t, result.T, 1, "SetCfg int")

	// SetCfg string
	cfg = &discover.CfgArgs{Key: "ccc", Value: "testtest"}
	err = dapi.ServiceSetCfg(nil, addr, cfg)
	assert.NoError(t, err, "SetCfg string")

	err = dapi.ServiceGet(nil, &info, addr)
	assert.NoError(t, err, "ServiceGet")
	err = info.Cfg.ToStruct(&result)
	assert.NoError(t, err, "ToStruct")
	assert.Equal(t, result.T, "testtest", "SetCfg string")
}
