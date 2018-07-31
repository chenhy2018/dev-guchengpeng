package slave

import (
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/assert"
	"qbox.us/errors"
	"qbox.us/http/account.v2.1/digest_auth"
	jsonlog "qbox.us/http/audit/jsonlog.v1"
	"qbox.us/servend/account"
	"qbox.us/servend/proxy_auth"
)

var cfg = `
{
	"bind_host": "0.0.0.0:19060",
	"max_procs": 8,
	"debug_level": 0,

	"mc_hosts": ["localhost:11211"],
	"auth": {
		"qconfg": {
			"mc_hosts": ["127.0.0.1:11211"],
			"master_hosts": ["http://localhost:18500"], 
			"access_key": "4_odedBxmrAHiu4Y0Qp0HPG0NANCf6VAsAjWL_k9",
			"secret_key": "SrRuUVfDX6drVRvpyN8mv8Vcm9XnMZzlbDfvVfMe",
			"lc_expires_ms": 30000,
			"lc_duration_ms": 3000, 
			"lc_chan_bufsize": 16000
		}
	},
	"is_proxy":true,
	"proxy_config":{
		"async_retry_failed_interval_sec"	: 60,
		"confs_srv": {
			"z0.nb230": {
				"client":{
					"hosts":["http://127.0.0.1:19063"],
					"try_times":2
				},
				"failover":{
					"hosts":["http://127.0.0.1:19061"],
					"try_times":2
				},
				"client_tr":{
					"dial_timeout_ms":500
				},
				"failover_tr":{
					"dial_timeout_ms":2000
				}
			},
			"z0.nb231": {
				"client":{
					"hosts":["http://127.0.0.1:19062"],
					"try_times":2
				},
				"client_tr":{
					"dial_timeout_ms":500
				}
			}
		},
		"admin_ak":"4_odedBxmrAHiu4Y0Qp0HPG0NANCf6VAsAjWL_k9",
		"admin_sk":"SrRuUVfDX6drVRvpyN8mv8Vcm9XnMZzlbDfvVfMe",
		"mgo": {
			"host": "localhost",
			"db": "qbox_confs",
			"coll": "confs",
			"mode": "strong",
			"timeout": 1
		}
	},

	
	"uid_mgr": 74121669,

	"auditlog": {
		"logdir": "./run/auditlog/confs",
		"chunkbits": 29
	}
}
`

type ConfigMain struct {
	McHosts  []string `json:"mc_hosts"` // 互为镜像的Memcache服务
	BindHost string   `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	AuthConf digest_auth.Config `json:"auth"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`

	UidMgr uint32 `json:"uid_mgr"` // 只接受这个管理员发过来的请求

	IsProxy     bool        `json:"is_proxy"` //如果为true表示当前服务端作为一个代理节点使用
	ProxyConfig ProxyConfig `json:"proxy_config"`
}

func TestSrv(t *testing.T) {
	var conf ConfigMain
	err := json.Unmarshal([]byte(cfg), &conf)
	assert.Equal(t, err, nil)

	cfg := &Config{
		McHosts:     conf.McHosts,
		UidMgr:      conf.UidMgr,
		IsProxy:     conf.IsProxy,
		ProxyConfig: conf.ProxyConfig,
	}
	service, err := New(cfg)
	if err != nil {
		log.Fatal("qconfs.New failed:", errors.Detail(err))
	}

	router := &webroute.Router{Factory: wsrpc.Factory}

	srv := httptest.NewServer(router.Register(service))
	time.Sleep(time.Second * 2)

	doTestFailedCount(t, srv.URL)
}

func doTestFailedCount(t *testing.T, url string) {
	cAdminUser := &rpc.Client{
		Client: proxy_auth.NewClient(account.UserInfo{Uid: 74121669, Utype: account.USER_TYPE_ADMIN}, nil),
	}
	var ret FailedCountRet

	err := cAdminUser.GetCall(xlog.NewDummy(), &ret, url+"/failedcount?hours_before=0")
	assert.Equal(t, err, nil)
	var curCount = ret.Count

	doTestRefresh(t, url)
	time.Sleep(time.Microsecond * 500)

	err = cAdminUser.GetCall(xlog.NewDummy(), &ret, url+"/failedcount?hours_before=0")
	assert.Equal(t, err, nil)
	assert.Equal(t, ret.Count, curCount+2)
}

func doTestRefresh(t *testing.T, url string) {
	cAdminUser := &rpc.Client{
		Client: proxy_auth.NewClient(account.UserInfo{Uid: 74121669, Utype: account.USER_TYPE_ADMIN}, nil),
	}

	params := map[string][]string{}
	params["id"] = append(params["id"], "ID")
	err := cAdminUser.CallWithForm(xlog.NewDummy(), nil, url+"/refresh", params)
	assert.NotEqual(t, err, nil)
}
