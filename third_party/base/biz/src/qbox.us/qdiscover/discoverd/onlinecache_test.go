package discoverd

import (
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo"
	"qbox.us/mgo3"
	"qbox.us/qdiscover/discover"
)

func TestSortByLastUpdate(t *testing.T) {

	now := time.Now()
	idxs := []int{3, 5, 0, 1, 2, 4}
	var infos []*discover.ServiceInfo
	for _, idx := range idxs {
		infos = append(infos, &discover.ServiceInfo{
			Name:       strconv.Itoa(idx),
			LastUpdate: now.Add(-time.Duration(idx)),
		})
	}
	sort.Sort(byLastUpdates(infos))
	for i, info := range infos {
		assert.Equal(t, strconv.Itoa(i), info.Name, "i:%v name:%v", i, info.Name)
		assert.Equal(t, now.Add(-time.Duration(i)), info.LastUpdate, "i:%v lastUpdate:%v", i, info.Name)
	}
}

type addrName struct {
	Addr string
	Name string
}

func diffIndex(svrs []addrName, services []*discover.ServiceInfo) int {

	n := len(svrs)
	if n != len(services) {
		return -2
	}
	for i := 0; i < n; i++ {
		if svrs[i].Addr != services[i].Addr || svrs[i].Name != services[i].Name {
			return i
		}
	}
	return -1
}

func TestServiceListOnline(t *testing.T) {

	mcfg := &mgo3.Config{
		Host: "localhost",
		DB:   "TestServiecListOnline",
		Coll: "service",
	}
	session := mgo3.Open(mcfg)
	session.DB.DropDatabase()

	s := &ServiceManager{
		session: session,
		colls:   []*mgo.Collection{session.Coll.Copy().Collection},
		Config: Config{
			HeartbeatMissSecs:           100,
			MinOnlineRatios:             map[string]float64{"image": 0.4, "ffmpeg": 0.7},
			MinOnlineRatioDefault:       0.6,
			OnlineCacheReloadIntervalMs: 500000000,
		},
	}

	err := s.ReloadOnlineCache()
	assert.NoError(t, err)
	infos, err := s.ListOnline("", []string{"image", "fopagent"})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(infos))

	svrs := []addrName{
		{"192.168.0.1:300", "fopagent"},
		{"192.168.0.1:100", "image"},
		{"192.168.0.1:101", "image"},
		{"192.168.0.1:102", "image"},

		{"192.168.0.2:300", "fopagent"},
		{"192.168.0.2:100", "image"},
		{"192.168.0.2:101", "image"},
		{"192.168.0.2:201", "ffmpeg"},
		{"192.168.0.2:202", "ffmpeg"},

		{"192.168.0.3:300", "fopagent"},
		{"192.168.0.3:200", "ffmpeg"},

		{"192.168.0.4:100", "xxx"},
	}
	for _, svr := range svrs {
		s.Register(svr.Addr, svr.Name, nil)
		s.Enable(svr.Addr)
	}
	s.Disable("192.168.0.1:101")

	err = s.ReloadOnlineCache()
	assert.NoError(t, err)

	infos, err = s.ListOnline("192.168.0.1", []string{"image", "ffmpeg"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex([]addrName{svrs[1], svrs[3]}, infos), -1)

	infos, err = s.ListOnline("192.168.0.2", []string{"image", "ffmpeg"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex(svrs[5:9], infos), -1)

	infos, err = s.ListOnline("192.168.0.3", []string{"image", "ffmpeg"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex(svrs[10:11], infos), -1)

	infos, err = s.ListOnline("", []string{"fopagent"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex([]addrName{svrs[0], svrs[4], svrs[9]}, infos), -1)

	s.HeartbeatMissSecs = 1
	time.Sleep(time.Second)

	for i, svr := range svrs {
		if i != 5 && i != 8 && i != 9 {
			s.Register(svr.Addr, svr.Name, nil)
		}
	}
	err = s.ReloadOnlineCache()
	assert.NoError(t, err)

	infos, err = s.ListOnline("192.168.0.1", []string{"image", "ffmpeg"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex([]addrName{svrs[1], svrs[3]}, infos), -1)

	infos, err = s.ListOnline("192.168.0.2", []string{"image", "ffmpeg"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex(svrs[6:9], infos), -1)

	infos, err = s.ListOnline("192.168.0.3", []string{"image", "ffmpeg"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex(svrs[10:11], infos), -1)

	infos, err = s.ListOnline("", []string{"fopagent"})
	assert.NoError(t, err)
	assert.Equal(t, diffIndex([]addrName{svrs[0], svrs[4]}, infos), -1)
}
