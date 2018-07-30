package discoverd_test

import (
	"runtime"
	"testing"
	"time"

	"qbox.us/mgo3"
	"qbox.us/qdiscover/discover"
	. "qbox.us/qdiscover/discoverd"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

const heartbeatMissSecs = 1

func cleanTestEnv(cfg *Config) {
	scoll := cfg.ServiceColl
	u := mgo3.Dail(scoll.Host, scoll.Mode, scoll.SyncTimeoutInS)
	u.DB(scoll.DB).C(scoll.Coll).DropCollection()
}

func checkStatus(t *testing.T, who string, s *ServiceManager, addr string, state discover.State, attrs discover.Attrs) {
	info, err := s.Get(addr)
	assert.NoError(t, err, who)
	assert.Equal(t, info.State, state, who)
	if attrs != nil {
		assert.Equal(t, info.Attrs, attrs, who)
	}
}

func checkCount(t *testing.T, s *ServiceManager, args *QueryArgs, expectCount int) (count int) {
	count, err := s.Count(args)
	assert.NoError(t, err, "Count")
	assert.Equal(t, count, expectCount, "Count")
	return
}

func checkList(
	t *testing.T, s *ServiceManager, args *QueryArgs, marker string, limit int,
	expectCount int) (infos []*discover.ServiceInfo, marker2 string) {

	infos, marker2, err := s.List(args, marker, limit)
	assert.NoError(t, err, "List")
	assert.Equal(t, len(infos), expectCount, "List")
	return
}

func checkListAll(
	t *testing.T, s *ServiceManager, args *QueryArgs,
	expectCount int) (infos []*discover.ServiceInfo) {

	_, _, line, _ := runtime.Caller(1)
	infos, err := s.ListAll(args)
	assert.NoError(t, err, "ListAll %v", line)
	assert.Equal(t, len(infos), expectCount, "ListAll, line %v", line)
	return
}

func newQueryArgs(node string, name []string, state discover.State) *QueryArgs {
	return &QueryArgs{
		Node:  node,
		Name:  name,
		State: string(state),
	}
}

func TestServiceManager(t *testing.T) {
	cfg := &Config{
		ServiceColl: mgo3.Config{
			Host: "localhost",
			DB:   "servicemgrTest",
			Coll: "service",
			Mode: "strong",
		},
		HeartbeatMissSecs:           heartbeatMissSecs,
		OnlineCacheReloadIntervalMs: 100000000000000000,
	}
	cleanTestEnv(cfg)

	s, err := NewServiceManager(cfg)
	assert.NoError(t, err, "NewServiceManager")

	// test register enable disable get unregister

	imageName := "image"
	imageAddr := "192.168.1.10:8000"
	imageAttrs := discover.Attrs{
		"processing": 100,
		"cmds":       []interface{}{"imageView", "imageMogr", "imageInfo"},
	}

	err = s.Register(imageAddr, imageName, imageAttrs)
	assert.NoError(t, err, "Register")
	checkStatus(t, "Register", s, imageAddr, discover.StatePending, imageAttrs)

	imageAttrs["processing"] = 1000
	err = s.Register(imageAddr, imageName, imageAttrs)
	assert.NoError(t, err, "Register")
	checkStatus(t, "Register", s, imageAddr, discover.StatePending, imageAttrs)

	err = s.Enable(imageAddr)
	assert.NoError(t, err, "Enable")
	checkStatus(t, "Enable", s, imageAddr, discover.StateEnabled, nil)

	err = s.Disable(imageAddr)
	assert.NoError(t, err, "Disable")
	checkStatus(t, "Disable", s, imageAddr, discover.StateDisabled, nil)

	err = s.Unregister(imageAddr)
	assert.NoError(t, err, "Unregister")

	_, err = s.Get(imageAddr)
	assert.Equal(t, err, ErrNoSuchEntry, "Get")

	// multi addrs, test count and list

	ffmpegName := "ffmpeg"
	ffmpegAddr := "192.168.1.100:9000"
	ffmpegAttrs := discover.Attrs{
		"processing": 100,
		"cmds":       []interface{}{"avinfo", "avthumb"},
	}

	err = s.Register(ffmpegAddr, ffmpegName, ffmpegAttrs)
	assert.NoError(t, err, "Register")
	checkStatus(t, "Register", s, ffmpegAddr, discover.StatePending, ffmpegAttrs)

	err = s.Register(imageAddr, imageName, imageAttrs)
	assert.NoError(t, err, "Register")
	checkStatus(t, "Register", s, imageAddr, discover.StatePending, imageAttrs)

	// 不存在的服务名
	args := newQueryArgs("", []string{"nosuchname"}, discover.StatePending)
	checkCount(t, s, args, 0)
	checkListAll(t, s, args, 0)

	// 不存在的 node
	args = newQueryArgs("nosuncnode", []string{""}, discover.StatePending)
	checkCount(t, s, args, 0)
	checkListAll(t, s, args, 0)

	// 仅 image 服务
	args = newQueryArgs("", []string{"image"}, discover.StatePending)
	checkCount(t, s, args, 1)
	infos := checkListAll(t, s, args, 1)
	checkStatus(t, "List", s, infos[0].Addr, discover.StatePending, imageAttrs)

	// 仅 image node
	args = newQueryArgs("192.168.1.10", []string{}, discover.StatePending)
	checkCount(t, s, args, 1)
	infos = checkListAll(t, s, args, 1)
	checkStatus(t, "List", s, infos[0].Addr, discover.StatePending, imageAttrs)

	// 不指定服务名，测试 list 的 marker 和 limit 参数，因为结果是根据 addr 排过序，所以 ffmpeg 在前，image 在后。

	args = newQueryArgs("", []string{}, discover.StatePending)
	checkCount(t, s, args, 2)

	args = newQueryArgs("", []string{}, "")
	infos, marker := checkList(t, s, args, "", 1, 1)
	assert.Equal(t, marker, imageAddr)
	checkStatus(t, "List", s, infos[0].Addr, discover.StatePending, ffmpegAttrs)

	args = newQueryArgs("", []string{"image", "ffmpeg"}, discover.StatePending)
	infos, marker = checkList(t, s, args, marker, 1, 1)
	assert.Equal(t, marker, "")
	checkStatus(t, "List", s, infos[0].Addr, discover.StatePending, imageAttrs)

	// online offline 判断

	s.ReloadOnlineCache()
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOnline), 0)
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOffline), 0)

	err = s.Enable(imageAddr)
	assert.NoError(t, err, "Enable")
	err = s.Enable(ffmpegAddr)
	assert.NoError(t, err, "Enable")

	s.ReloadOnlineCache()
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOnline), 2)
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOffline), 0)

	// 心跳丢失
	time.Sleep(time.Duration(heartbeatMissSecs) * time.Second)

	s.ReloadOnlineCache()
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOnline), 0)
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOffline), 2)

	// 恢复 image 心跳
	err = s.Register(imageAddr, imageName, imageAttrs)
	assert.NoError(t, err, "Register")
	checkStatus(t, "Register", s, imageAddr, discover.StateEnabled, imageAttrs)

	s.ReloadOnlineCache()
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOnline), 1)
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOffline), 1)

	// 恢复 ffmpeg 心跳
	err = s.Register(ffmpegAddr, ffmpegName, ffmpegAttrs)
	assert.NoError(t, err, "Register")
	checkStatus(t, "Register", s, ffmpegAddr, discover.StateEnabled, ffmpegAttrs)

	s.ReloadOnlineCache()
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOnline), 2)
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOffline), 0)

	// disable 掉 image
	err = s.Disable(imageAddr)
	assert.NoError(t, err, "Enable")

	s.ReloadOnlineCache()
	infos = checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOnline), 1)
	checkStatus(t, "Register", s, infos[0].Addr, discover.StateEnabled, ffmpegAttrs)
	infos = checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOffline), 1)
	checkStatus(t, "Register", s, infos[0].Addr, discover.StateDisabled, imageAttrs)

	// diable 掉 ffmpeg
	err = s.Disable(ffmpegAddr)
	assert.NoError(t, err, "Enable")

	s.ReloadOnlineCache()
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOnline), 0)
	checkListAll(t, s, newQueryArgs("", []string{}, discover.StateOffline), 2)
}

func TestSortUniq(t *testing.T) {
	xl := xlog.NewWith("TestUniq")

	var in1 []*discover.ServiceInfo
	out1 := SortUniq(in1)
	assert.Equal(t, in1, out1)

	in2 := []*discover.ServiceInfo{
		{Addr: "0"},
	}
	out2 := SortUniq(in2)
	xl.Info(out2)
	assert.Equal(t, in2, out2)

	in3 := []*discover.ServiceInfo{
		{Addr: "2"},
		{Addr: "0"},
		{Addr: "1"},
	}
	out3 := SortUniq(in3)
	xl.Info(out3)
	expectOut3 := []*discover.ServiceInfo{
		{Addr: "0"},
		{Addr: "1"},
		{Addr: "2"},
	}
	assert.Equal(t, out3, expectOut3)

	in4 := []*discover.ServiceInfo{
		{Addr: "1"},
		{Addr: "2"},
		{Addr: "0"},
		{Addr: "1"},
		{Addr: "1"},
		{Addr: "2"},
		{Addr: "0"},
		{Addr: "2"},
		{Addr: "2"},
		{Addr: "0"},
	}
	out4 := SortUniq(in4)
	xl.Info(out4)
	expectOut4 := []*discover.ServiceInfo{
		{Addr: "0"},
		{Addr: "1"},
		{Addr: "2"},
	}
	assert.Equal(t, out4, expectOut4)
}

func TestServiceSetCfg(t *testing.T) {
	cfg := &Config{
		ServiceColl: mgo3.Config{
			Host: "localhost",
			DB:   "servicemgrTest",
			Coll: "service",
			Mode: "strong",
		},
		HeartbeatMissSecs: heartbeatMissSecs,
	}
	cleanTestEnv(cfg)

	s, err := NewServiceManager(cfg)
	assert.NoError(t, err, "NewServiceManager")

	imageName := "imageeee"
	imageAddr := "192.168.1.101:8888"
	imageAttrs := discover.Attrs{
		"processing": 100,
		"cmds":       []interface{}{"imageView", "imageMogr", "imageInfo"},
	}

	// SetCfg should be failed when no such addr
	args := &discover.CfgArgs{Key: "ttttt", Value: "t"}
	err = s.SetCfg(imageAddr, args)
	assert.Equal(t, err, ErrNoSuchEntry, "SetCfg")

	err = s.Register(imageAddr, imageName, imageAttrs)
	assert.NoError(t, err, "Register")
	checkStatus(t, "Register", s, imageAddr, discover.StatePending, imageAttrs)

	err = s.SetCfg(imageAddr, args)
	assert.NoError(t, err, "SetCfg")

	type Tcfg struct {
		T interface{} `bson:"ttttt"`
	}
	var result Tcfg
	info, err := s.Get(imageAddr)
	assert.NoError(t, err, "SetCfg")
	err = info.Cfg.ToStruct(&result)
	assert.NoError(t, err, "ToStruct")
	assert.Equal(t, result.T, "t", "SetCfg")
}
