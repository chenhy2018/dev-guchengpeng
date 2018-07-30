package mq

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
	"github.com/stretchr/testify.v2/assert"
	. "qbox.us/api/conf"
	"qbox.us/auditlog2"
	"qbox.us/mockacc"
	admin_api "qbox.us/qmq/qmqapi/v1/admin_api/mq"
	api "qbox.us/qmq/qmqapi/v1/mq"
	"qbox.us/servend/account"
)

var Host string

func init() {
	log.SetOutputLevel(1)
}

func doTestMq(t *testing.T) {

	adminTrans := mockacc.MakeTransport(0, account.USER_TYPE_ADMIN)
	admin_mq := admin_api.New(Host, adminTrans)

	MQ_HOST = Host
	trans := mockacc.MakeTransport(0, 0)
	mq := api.NewEx(trans)

	msg1 := "hello.world"
	msg2 := "hello.golang"

	err := admin_mq.Make(nil, "0-prefop", 1)
	if err != nil {
		ts.Fatal(t, err)
	}

	_, err = mq.PutString(nil, "prefop", msg1)
	if err != nil {
		ts.Fatal(t, err)
	}

	_, err = mq.PutString(nil, "prefop", msg2)
	if err != nil {
		ts.Fatal(t, err)
	}

	buf1, _, err := mq.Get(nil, "prefop")
	if err != nil {
		ts.Fatal(t, err)
	}
	if string(buf1) != msg1 {
		ts.Fatal(t, "mq.Get failed", string(buf1))
	}

	buf2, _, err := mq.Get(nil, "prefop")
	if err != nil {
		ts.Fatal(t, err)
	}
	if string(buf2) != msg2 {
		ts.Fatal(t, "mq.Get failed", string(buf2))
	}

	_, _, err = mq.Get(nil, "prefop")
	if err == nil {
		ts.Fatal(t, "mq.Get failed")
	}

	time.Sleep(2 * time.Second)

	buf1, msgid1, err := mq.Get(nil, "prefop")
	if err != nil {
		ts.Fatal(t, err)
	}
	if string(buf1) != msg1 {
		ts.Fatal(t, "mq.Get failed", string(buf1))
	}

	err = mq.Delete(nil, "prefop", msgid1)
	if err != nil {
		ts.Fatal(t, err)
	}

	time.Sleep(2 * time.Second)

	buf2, _, err = mq.Get(nil, "prefop")
	if err != nil {
		ts.Fatal(t, err)
	}
	if string(buf2) != msg2 {
		ts.Fatal(t, "mq.Get failed", string(buf2))
	}

	_, _, err = mq.Get(nil, "prefop")
	if err == nil {
		ts.Fatal(t, "mq.Get failed")
	}

}

func mqRun(t *testing.T) *Service {
	root, err := ioutil.TempDir("", "mq-run")
	if err != nil {
		log.Fatal(err)
	}

	logdir, err := ioutil.TempDir("", "mq-run-audit")
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Account:  mockacc.Account{},
		DataPath: root,
		Config: auditlog2.Config{
			LogFile: logdir,
		},
		ChunkBits:     4,
		Expires:       1,
		SaveHours:     0,
		CheckInterval: 3,
	}
	r, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	err = r.RegisterHandlers(mux)
	if err != nil {
		t.Fatal(err)
	}
	svr := httptest.NewServer(mux)

	Host = svr.URL

	return r
}

func TestMq(t *testing.T) {
	r := mqRun(t)
	time.Sleep(1e9)
	doTestMq(t)
	doTestDeleteDataFiles(t, r)
}

func doTestDeleteDataFiles(t *testing.T, r *Service) {
	var names []string
	var mqs []*Instance
	r.mutex.RLock()
	len := len(r.mqs)
	fmt.Printf("%+v\n", r.mqs)
	for name, mq := range r.mqs {
		names = append(names, name)
		mqs = append(mqs, mq)
	}
	r.mutex.RUnlock()

	//mq instance: 0-prefop
	assert.Equal(t, 1, len)

	//初始化instance data files
	for _, name := range names {
		setDataFiles(path.Join(r.DataPath, name))
	}

	testSaveHours = func() {
		r.SaveHours = 0
	}
	chunkSize := 1 << r.ChunkBits
	count := r.deleteDataFiles(int64(chunkSize))
	assert.Equal(t, 3, count)
}

func setDataFiles(dataPath string) {
	content := "0123"
	for i := 0; i < 50; i++ {
		fileNname := strconv.FormatInt(int64(i), 36)
		fname := path.Join(dataPath, fileNname)
		err := ioutil.WriteFile(fname, []byte(content), 0744)
		if err != nil {
			log.Fatalln("ioutil.WriteFile", err)
		}
	}
}
