package master

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"qbox.us/mgo2"
	"qbox.us/mockacc"
	"qbox.us/qconf/qconfapi"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/jsonrpc.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

func init() {
	log.SetOutputLevel(0)
}

var (
	MasterHost = "MHost"
	SlaveHost  = "SHost"
)

func doTestSetGet(t *testing.T, client *qconfapi.Client) {
	var err error
	xl := xlog.NewDummy()

	log.Println("SetProp")
	val := uint32(441375670)
	err = client.SetProp(xl, "key", "prop", val)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("GetProp")
	var ret struct {
		Prop uint32 `bson:"prop"`
	}
	err = client.Get(xl, &ret, "key", qconfapi.Cache_NoSuchEntry)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(ret.Prop)
	if ret.Prop != val {
		t.Fatal("ret.prop != val", ret.Prop, val)
	}

	err = client.Get(xl, &ret, "nosuchkey", qconfapi.Cache_NoSuchEntry)
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func doTestOldProtocol(t *testing.T, user *mockacc.UserInfo) {
	var err error
	xl := xlog.NewDummy()
	val := uint32(441370) // 441375670 will fail

	cli := digest.NewClient(&digest.Mac{
		AccessKey: user.AccessKey,
		SecretKey: []byte(user.SecretKey),
	}, nil)

	err = rpc.Client{cli}.CallWithJson(xl, nil, MasterHost+"/put", putArgs{"key", M{"$set": M{"prop": val}}})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := rpc.Client{cli}.PostWithForm(xl, MasterHost+"/get", map[string][]string{
		"id": {"key"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var ret struct {
		Prop uint32 `json:"prop"`
	}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		t.Fatal(err)
	}
	if ret.Prop != val {
		t.Fatal("ret.Prop != val", ret.Prop, val)
	}
}

var user = mockacc.GetSa()[0]

func TestQconfm(t *testing.T) {

	runMockSlave(t)
	runMaster(t)

	client := qconfapi.New(&qconfapi.Config{
		MasterHosts: []string{MasterHost},
		AccessKey:   user.AccessKey,
		SecretKey:   user.SecretKey,
	})

	doTestSetGet(t, client)
	doTestOldProtocol(t, &user)
}

func runMaster(t *testing.T) {
	c := mgo2.Open(&mgo2.Config{
		Host: "localhost",
		DB:   "qbox_conf_test",
		Coll: "conf",
		Mode: "strong",
	})
	conf := &Config{
		Coll:         c.Coll,
		MgrAccessKey: user.AccessKey,
		MgrSecretKey: user.SecretKey,
		SlaveHosts:   [][]string{[]string{SlaveHost}},
		AuthParser:   mockacc.NewParser(mockacc.GetSa()),
		UidMgr:       user.Uid,
	}

	service, err := New(conf)
	if err != nil {
		t.Fatal("qconfm.New failed:", errors.Detail(err))
	}

	mux := http.NewServeMux()
	factory := wsrpc.Factory.Union(bsonrpc.Factory).Union(jsonrpc.Factory)
	router := &webroute.Router{Factory: factory, Mux: mux}
	svr := httptest.NewServer(router.Register(service))
	MasterHost = svr.URL
}

// =========================================================

type MockSlave struct {
}

func newMockSlave() *MockSlave {
	return &MockSlave{}
}

func (m *MockSlave) DoRefresh(w http.ResponseWriter, req *http.Request) {
	log.Println("Slave refresh!!!!!!!!!!!!!!!!!")
	w.WriteHeader(200)
}

func runMockSlave(t *testing.T) {

	mux := http.NewServeMux()
	serviceSlave := newMockSlave()
	router := &webroute.Router{Mux: mux}
	router.Register(serviceSlave)
	svr := httptest.NewServer(mux)
	SlaveHost = svr.URL
}
