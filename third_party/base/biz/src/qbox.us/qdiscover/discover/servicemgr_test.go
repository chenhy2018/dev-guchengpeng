package discover

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutputLevel(0)
}

type errorLister struct{}

func (el errorLister) ServiceListAllEx(l rpc.Logger, ret interface{}, args *QueryArgs) error {
	return errors.New("errorLister")
}

type mockLister struct {
	services []*ServiceInfo
}

func (ml *mockLister) ServiceListAllEx(l rpc.Logger, ret interface{}, args *QueryArgs) error {
	v := reflect.ValueOf(ret)
	if v.Kind() != reflect.Ptr {
		return errors.New("ret not a ptr")
	}
	v = v.Elem()

	list := ServiceListRet{Items: ml.services}
	v.Set(reflect.ValueOf(list))
	return nil
}

func TestServiceManager(t *testing.T) {
	os.RemoveAll(DefaultPersistDir)
	defer os.RemoveAll(DefaultPersistDir)

	cfg := Config{
		DiscoverHosts:     []string{"mock"},
		FetchIntervalSecs: 1,
		IsServicesChanged: isAddrsChanged,
	}
	_, err := newServiceManager(&cfg, errorLister{})
	assert.Error(t, err, "NewServiceManager")

	// prepare attrs
	attrs := map[string]interface{}{
		"processing": 100,
		"cmds":       []string{"imageView", "imageMogr"},
	}

	// prepare two services
	aservice := &ServiceInfo{
		Addr:  "192.168.1.101:8000",
		Name:  "image",
		State: StateOnline,
		Attrs: attrs,
	}
	bservice := &ServiceInfo{
		Addr:  "192.168.1.102:8000",
		Name:  "image",
		State: StateOnline,
		Attrs: attrs,
	}

	// new with no services
	lister := &mockLister{services: nil}
	s, err := newServiceManager(&cfg, lister)
	assert.NoError(t, err, "NewServiceManager")

	notify := s.ChangeNotify()

	// one service
	lister.services = []*ServiceInfo{aservice}
	<-notify
	services := s.Services()
	assert.Equal(t, services[0], aservice)

	// 模拟 discover 挂掉，测试 discoverdRecoverTime 是否生效
	DiscoverdRecoverTime = 3 * time.Second
	fetchInterval := time.Duration(cfg.FetchIntervalSecs) * time.Second
	s.client = errorLister{}
	time.Sleep(2 * fetchInterval) // 后台 fetch 会失败，认为 discover 挂了
	// 恢复 fetch，让 fetch 返回 2 个服务
	lister.services = []*ServiceInfo{aservice, bservice}
	s.client = lister

	<-notify
	services = s.Services()
	assert.Equal(t, services[0], aservice)
	assert.Equal(t, services[1], bservice)

	// no notify
	assertNoChange(t, notify, "")

	// new success using persist file
	s, err = newServiceManager(&cfg, errorLister{})
	assert.NoError(t, err, "NewServiceManager")
	assert.Equal(t, services[0], aservice)
	assert.Equal(t, services[1], bservice)
}

func TestServiceManager_Protect(t *testing.T) {
	os.RemoveAll(DefaultPersistDir)
	defer os.RemoveAll(DefaultPersistDir)

	cfg := Config{
		DiscoverHosts:        []string{"mock"},
		FetchIntervalSecs:    1,
		IsServicesChanged:    isAddrsChanged,
		MinServicesNum:       4,
		MaxServicesDescRatio: 0.5,
	}

	services := make([]*ServiceInfo, 10)
	for i := range services {
		services[i] = &ServiceInfo{
			Addr: fmt.Sprintf("192.168.1.1%02d", i),
		}
	}
	lister := &mockLister{services: services}

	s, err := newServiceManager(&cfg, lister)
	assert.NoError(t, err)
	notify := s.ChangeNotify()

	// 10 -> 4 MaxServicesDescRatio
	lister.services = services[:4]
	assertNoChange(t, notify, "MaxServicesDescRatio")

	// 10 -> 5 OK
	lister.services = services[:5]
	<-notify
	assert.Equal(t, s.Services(), lister.services)

	// 5 -> 3 MinServicesNum
	lister.services = services[:3]
	assertNoChange(t, notify, "MinServicesNum")

	// 5 -> 10 OK
	lister.services = services[:10]
	<-notify
	assert.Equal(t, s.Services(), lister.services)
}

func assertNoChange(t *testing.T, notify <-chan bool, msg string) {
	select {
	case <-notify:
		log.Fatal("should not happen:", msg)
	case <-time.After(1500 * time.Millisecond):
	}
}
