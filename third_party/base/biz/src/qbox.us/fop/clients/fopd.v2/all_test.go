package fopd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qbox.us/fop"
)

var fopCtx = fop.FopCtx{
	CmdName:    "rawQuery",
	RawQuery:   "rawQuery",
	StyleParam: "styleParam",
	URL:        "url",
	Token:      "token",
	MimeType:   "mimeType",
	Mode:       2,
	Uid:        100,
	Bucket:     "bucket",
	Key:        "key",
	Fh:         []byte("fh"),
}

var (
	host200     string
	host599     string
	hostTimeout string
	hostErr     string = "errHost"
)

var (
	err599 = httputil.NewError(599, "server 599")
)

func init() {
	log.SetOutputLevel(0)
	svr200 := new200svr()
	host200 = svr200.URL
	svr599 := new599svr()
	host599 = svr599.URL
	svrTimeout := newTimeoutSvr()
	hostTimeout = svrTimeout.URL
	FailRecoverIntervalSecs = 2
}

func new200svr() *httptest.Server {
	handler200 := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("helloworld"))
	}
	return httptest.NewServer(http.HandlerFunc(handler200))
}

func new570svr() *httptest.Server {
	handler500 := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(570)
	}
	return httptest.NewServer(http.HandlerFunc(handler500))
}

func new599svr() *httptest.Server {
	handler599 := func(w http.ResponseWriter, req *http.Request) {
		httputil.Error(w, err599)
	}
	return httptest.NewServer(http.HandlerFunc(handler599))
}

func newTimeoutSvr() *httptest.Server {
	handlerTimeout := func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("helloworld"))
	}
	return httptest.NewServer(http.HandlerFunc(handlerTimeout))
}

func newOutSvr() *httptest.Server {
	handler := func(w http.ResponseWriter, req *http.Request) {
		_, _, ctx, _ := DecodeQuery(req.URL.RawQuery)
		if ctx.OutType != "" {
			m := map[string]interface{}{
				"out":   ctx.OutType,
				"dcKey": ctx.OutDCKey,
				"rsKey": ctx.OutRSKey,
			}
			httputil.Reply(w, 200, m)
			return
		}
		b, _ := ioutil.ReadAll(req.Body)
		if len(b) == 0 {
			w.Write([]byte("helloworld"))
		} else {
			w.Write(b)
		}
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func newSlowSvr() *httptest.Server {
	handler := func(w http.ResponseWriter, req *http.Request) {
		xl := xlog.NewWithReq(req)
		xl.Info("handle: start")
		defer xl.Info("handle: done")

		time.Sleep(1e9) // slow request

		_, _, ctx, _ := DecodeQuery(req.URL.RawQuery)
		if ctx.OutType != "" {
			m := map[string]interface{}{
				"out":   ctx.OutType,
				"dcKey": ctx.OutDCKey,
				"rsKey": ctx.OutRSKey,
			}
			httputil.Reply(w, 200, m)
			return
		}
		b, _ := ioutil.ReadAll(req.Body)
		if len(b) == 0 {
			w.Write([]byte("helloworld"))
		} else {
			w.Write(b)
		}
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestConn(t *testing.T) {
	tpCtx := xlog.NewContextWith(context.Background(), "TestConn")

	conn := NewConn(hostErr, nil)
	_, err := conn.Op2(tpCtx, []byte("fh"), 0, &fopCtx)
	assert.Error(t, err)
	assert.True(t, conn.LastFailedTime() > 0)
	assert.Equal(t, conn.ProcessingNum(), 0)

	_, err = conn.Op(tpCtx, strings.NewReader("file"), 4, &fopCtx)
	assert.Error(t, err)

	conn = NewConn(host200, nil)
	_, err = conn.Op2(tpCtx, []byte("fh"), 0, &fopCtx)
	assert.NoError(t, err)
	assert.Equal(t, conn.LastFailedTime(), 0)

	_, err = conn.Op(tpCtx, strings.NewReader("file"), 4, &fopCtx)
	assert.NoError(t, err)

	conn = NewConn(host599, nil)
	_, err = conn.Op2(tpCtx, []byte("fh"), 0, &fopCtx)
	assert.Equal(t, err.Error(), err599.Error())

	_, err = conn.Op(tpCtx, strings.NewReader("file"), 4, &fopCtx)
	assert.Equal(t, err.Error(), err599.Error())
}

func TestSelector0(t *testing.T) {
	tpCtx := xlog.NewContextWith(context.Background(), "TestSelector0")
	conns := []*ConnInfo{
		{Conn: NewConn(host200, nil)},
		{Conn: NewConn(hostErr, nil)},
		{Conn: NewConn(host599, nil)},
	}
	s := &lbSelector0{Conns: conns}
	conn, _ := s.PickConn()
	assert.Equal(t, conn.Host, hostErr)
	conn, _ = s.PickConn()
	assert.Equal(t, conn.Host, host599)
	conn, _ = s.PickConn()
	assert.Equal(t, conn.Host, host200)
	conns[1].Conn.Op2(tpCtx, []byte("fh"), 0, &fopCtx)
	conn, _ = s.PickConn()
	assert.Equal(t, conn.Host, host599)
	time.Sleep(time.Duration(FailRecoverIntervalSecs) * time.Second) // wait to hostErr recover
	conn, _ = s.PickConn()
	assert.Equal(t, conn.Host, host200)
	conn, _ = s.PickConn()
	assert.Equal(t, conn.Host, hostErr)
}

func TestSelector1(t *testing.T) {
	tpCtx := xlog.NewContextWith(context.Background(), "TestSelector1")
	conns := []*ConnInfo{
		{Conn: NewConn(host200, nil)},
		{Conn: NewConn(hostTimeout, nil)},
	}
	s := &lbSelector1{Conns: conns}
	go func() { conns[1].Conn.Op2(tpCtx, []byte("fh"), 0, &fopCtx) }()
	time.Sleep(time.Second)
	conn, _ := s.PickConn()
	assert.Equal(t, conn.Host, host200)
	conn, _ = s.PickConn()
	assert.Equal(t, conn.Host, host200)
	conn, _ = s.PickConn()
	assert.Equal(t, conn.Host, host200)
}

func TestSelector_minLoadIndex(t *testing.T) {
	times := map[int]int64{}
	cases := []int64{1, 2, 1, 2, 1}
	for i := 0; i < 1000; i++ {
		index := minLoadIndex(cases)
		times[index]++
	}
	// index only can be 0,2,4
	// index value must > 250
	for index, v := range times {
		if index == 1 || index == 3 {
			t.Fatalf("index got %v, expect 0,2,4\n", index)
		}
		if v < 250 {
			t.Fatalf("not random selection, got %v, expect v > 250\n", v)
		}
	}
}

func TestSelector_shuffleConns(t *testing.T) {
	xl := xlog.NewWith("TestSelector_shuffleConns")
	conns := []*ConnInfo{
		{Conn: &Conn{Host: "localhost:12306"}},
		{Conn: &Conn{Host: "localhost"}},
		{Conn: &Conn{Host: "xxxx:12306"}},
		{Conn: &Conn{Host: "yyyy:12306"}},
		{Conn: &Conn{Host: "localhost:12306"}},
		{Conn: &Conn{Host: "yyyy:12306"}},
		{Conn: &Conn{Host: "xxxx:12306"}},
		{Conn: &Conn{Host: "localhost:12306"}},
		{Conn: &Conn{Host: "localhost:12306"}},
	}
	shuffleConns(conns)
	for _, c := range conns {
		xl.Info(c.Conn.Host)
	}
	for i := 1; i <= 6; i++ {
		if getIp(conns[i-1].Conn.Host) == getIp(conns[i].Conn.Host) {
			t.Error(i-1, i)
		}
	}
}

func TestClient_Retry570(t *testing.T) {
	svr570 := new570svr()
	host570 := svr570.URL
	defer svr570.Close()

	xl := xlog.NewWith("TestClient_Retry500")
	tpCtx := xlog.NewContext(context.Background(), xl)

	svrs := []HostInfo{
		HostInfo{
			Host: host570,
			Cmds: []string{"a"},
		},
	}
	cfg := Config{Servers: svrs}
	client, _ := NewClient(&cfg)

	ctx := fopCtx
	ctx.CmdName = "a"
	ctx.RawQuery = "a/1"
	xl.Info("Op: ", ctx.RawQuery)
	_, err := client.Op2(tpCtx, []byte("fh"), 0, &ctx)
	assert.Error(t, err)
	assert.Equal(t, client.Stats().GetItem(MakeStatsKey("a", 0, host570)).Failed.Get(), 2)
	assert.Equal(t, client.Stats().GetItem(MakeStatsKey("a", 0, host570)).Retry.Get(), 2)
}

func TestClient(t *testing.T) {
	svr1 := new200svr()
	host1 := svr1.URL
	svr2 := newTimeoutSvr()
	host2 := svr2.URL
	defer func() {
		svr1.Close()
		svr2.Close()
	}()

	svrs := []HostInfo{
		HostInfo{
			Host: host1,
			Cmds: []string{"a", "b"},
		},
		HostInfo{
			Host: hostErr,
			Cmds: []string{"a", "b"},
		},
		HostInfo{
			Host: host2,
			Cmds: []string{"c"},
		},
	}
	cfg := Config{
		Servers:  svrs,
		NotCache: []string{"b", "c"},
		Timeouts: map[string]int{
			"c": 1,
		},
		LoadBalanceMode: map[string]int{
			"b": 1,
		},
		DefaultTimeout: 3,
	}

	xl := xlog.NewWith("TestClient")
	tpCtx := xlog.NewContext(context.Background(), xl)
	client, _ := NewClient(&cfg)
	cmds := client.List()
	xl.Info("cmds:", cmds)
	assert.Equal(t, len(cmds), 3)
	for _, cmd := range client.List() {
		assert.Equal(t, client.HasCmd(cmd), true)
	}
	assert.Equal(t, client.NeedCache("a"), true)
	assert.Equal(t, client.NeedCache("b"), false)
	assert.Equal(t, client.NeedCache("c"), false)

	ctx := fopCtx
	ctx.CmdName = "a"
	ctx.RawQuery = "a/1"
	xl.Info("Op: ", ctx.RawQuery)
	_, err := client.Op2(tpCtx, []byte("fh"), 0, &ctx)
	assert.NoError(t, err)
	_, err = client.Op2(tpCtx, []byte("fh"), 0, &ctx)
	assert.NoError(t, err)
	assert.Equal(t, client.Stats().GetItem(MakeStatsKey("a", 0, hostErr)).Failed.Get(), 1)
	assert.Equal(t, client.Stats().GetItem(MakeStatsKey("a", 0, hostErr)).Retry.Get(), 1)

	ctx.CmdName = "b"
	ctx.RawQuery = "b/1"
	xl.Info("Op: ", ctx.RawQuery)
	_, err = client.Op2(tpCtx, []byte("fh"), 0, &ctx)
	assert.NoError(t, err)
	_, err = client.Op2(tpCtx, []byte("fh"), 0, &ctx)
	assert.NoError(t, err)

	ctx.CmdName = "c"
	ctx.RawQuery = "c/1"
	xl.Info("Op: ", ctx.RawQuery)
	_, err = client.Op2(tpCtx, []byte("fh"), 0, &ctx)
	assert.True(t, isResponseTimeout(err))
	_, err = client.Op2(tpCtx, []byte("fh"), 0, &ctx)
	assert.True(t, isResponseTimeout(err))
	assert.Equal(t, client.Stats().GetItem(MakeStatsKey("c", 0, host2)).Timeout.Get(), 2)
	assert.Equal(t, client.Stats().GetItem(MakeStatsKey("c", 0, host2)).Failed.Get(), 2)
	assert.Equal(t, client.Stats().GetItem(MakeStatsKey("c", 0, host2)).Retry.Get(), 0) // ResponseHeaderTimeout 不重试

	client.Close()
}

func TestPipe(t *testing.T) {
	svrOut := newOutSvr()
	hostOut := svrOut.URL
	defer svrOut.Close()

	tpCtx := xlog.NewContextWith(context.Background(), "TestPipe")

	svrs := []HostInfo{
		HostInfo{
			Host: hostOut,
			Cmds: []string{"a"},
		},
	}
	cfg := Config{Servers: svrs}
	client, _ := NewClient(&cfg)

	pipe := NewPipe(&PipeConfig{Fopd: client})

	tasks := newPipeTasks("a")
	resp, err := pipe.Exec(tpCtx, []byte("fh"), 0, tasks, fop.Out{})
	assert.NoError(t, err)
	assertStreamResponse(t, resp, "helloworld")

	tasks = newPipeTasks("a")
	out := fop.Out{Type: fop.OutTypeDC, DCKey: []byte("dckey")}
	resp, err = pipe.Exec(tpCtx, []byte("fh"), 0, tasks, out)
	assert.NoError(t, err)
	assertOutResponse(t, resp, out)

	tasks = newPipeTasks("a/1", "a/2", "a/3")
	resp, err = pipe.Exec(tpCtx, []byte("fh"), 0, tasks, fop.Out{})
	assert.NoError(t, err)
	assertStreamResponse(t, resp, "helloworld")

	tasks = newPipeTasks("a/1", "a/2", "a/3")
	out = fop.Out{Type: fop.OutTypeRS, RSKey: "rskey"}
	resp, err = pipe.Exec(tpCtx, []byte("fh"), 0, tasks, out)
	assert.NoError(t, err)
	assertOutResponse(t, resp, out)

	client.Close()
}

func TestPipeConnError(t *testing.T) {
	assert := assert.New(t)

	svrOut := newOutSvr()
	hostOut := svrOut.URL
	defer svrOut.Close()

	tpCtx := xlog.NewContextWith(context.Background(), "TestPipe")

	svrs := []HostInfo{
		HostInfo{
			Host: hostOut,
			Cmds: []string{"a"},
		},
		HostInfo{
			Host: "http://not_exsits",
			Cmds: []string{"b"},
		},
	}
	cfg := Config{Servers: svrs}
	client, _ := NewClient(&cfg)

	pipe := NewPipe(&PipeConfig{Fopd: client})

	tasks := newPipeTasks("a")
	resp, err := pipe.Exec(tpCtx, []byte("fh"), 0, tasks, fop.Out{})
	assert.NoError(err)
	assertStreamResponse(t, resp, "helloworld")

	tasks = newPipeTasks("a", "b")
	out := fop.Out{Type: fop.OutTypeDC, DCKey: []byte("dckey")}
	_, err = pipe.Exec(tpCtx, []byte("fh"), 0, tasks, out)
	assert.NotNil(err)

	tasks = newPipeTasks("a/1", "b/2", "a/3")
	_, err = pipe.Exec(tpCtx, []byte("fh"), 0, tasks, fop.Out{})
	assert.NotNil(err)

	client.Close()
}

func TestPipeUidCmdLimit(t *testing.T) {
	svr := newSlowSvr()
	host := svr.URL
	defer svr.Close()

	svrs := []HostInfo{
		HostInfo{
			Host: host,
			Cmds: []string{"a", "b"},
		},
	}
	cfg := Config{Servers: svrs}
	client, _ := NewClient(&cfg)

	pipe := NewPipe(&PipeConfig{client, 1})
	out := fop.Out{Type: fop.OutTypeRS, RSKey: "rskey"}

	/// cmd a, uid 1
	tasksCmdAUid1 := []*fop.FopCtx{
		&fop.FopCtx{CmdName: "a", Uid: 1, RawQuery: "a"},
	}
	ctxCmdAUid1 := xlog.NewContextWith(context.Background(), "1:a")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := pipe.Exec(ctxCmdAUid1, []byte("fh"), 1, tasksCmdAUid1, out) // 该请求会卡住
		assert.NoError(t, err)
		defer resp.Body.Close()
	}()

	time.Sleep(5e8) // wait go func()

	_, err := pipe.Exec(ctxCmdAUid1, []byte("fh"), 1, tasksCmdAUid1, out)
	assert.Error(t, err) // 此时超过限制报错
	assert.Equal(t, ErrOpsPerUidCmdOutOfLimit.Code, httputil.DetectCode(err))

	/// cmd b, uid 1
	tasksCmdBUid1 := []*fop.FopCtx{
		&fop.FopCtx{CmdName: "b", Uid: 1, RawQuery: "b"},
	}
	ctxCmdBUid1 := xlog.NewContextWith(context.Background(), "1:b")

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := pipe.Exec(ctxCmdBUid1, []byte("fh"), 1, tasksCmdBUid1, out) // 该请求会卡住
		assert.NoError(t, err)
		defer resp.Body.Close()
	}()

	// cmd a, uid 2
	tasksCmdAUid2 := []*fop.FopCtx{
		&fop.FopCtx{CmdName: "a", Uid: 2, RawQuery: "a"},
	}
	ctxCmdAUid2 := xlog.NewContextWith(context.Background(), "2:a")

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := pipe.Exec(ctxCmdAUid2, []byte("fh"), 1, tasksCmdAUid2, out) // 该请求会卡住
		assert.NoError(t, err)
		defer resp.Body.Close()
	}()

	wg.Wait()

	return
}

func newPipeTasks(cmds ...string) []*fop.FopCtx {
	var tasks []*fop.FopCtx
	for _, cmd := range cmds {
		task := fopCtx
		task.CmdName = strings.SplitN(cmd, "/", 2)[0]
		task.RawQuery = cmd
		tasks = append(tasks, &task)
	}
	return tasks
}

func assertStreamResponse(t *testing.T, resp *http.Response, expectedData string) {
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(b), expectedData)
}

func assertOutResponse(t *testing.T, resp *http.Response, out fop.Out) {
	defer resp.Body.Close()
	var ret struct {
		Out   string `json:"out"`
		RSKey string `json:"rsKey"`
		DCKey []byte `json:"dcKey"`
	}
	err := json.NewDecoder(resp.Body).Decode(&ret)
	assert.NoError(t, err)
	assert.Equal(t, ret.Out, out.Type)
	assert.Equal(t, ret.RSKey, out.RSKey)
	assert.Equal(t, string(ret.DCKey), string(out.DCKey))
}
