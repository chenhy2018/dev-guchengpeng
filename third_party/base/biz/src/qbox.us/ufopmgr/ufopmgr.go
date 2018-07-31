package ufopmgr

import (
	"sync"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/api/qconf/ufopg"
)

const (
	DefaultListInterval = 300 // 5min
)

type UfopLister interface {
	List(l rpc.Logger) (ufopg.ListRet, error)
}

type UfopItem struct {
	AclMode byte
	AclList map[uint32]bool
	Url     string
	Method  byte
}

type UfopMgr struct {
	UfopLister
	ListInterval time.Duration
	Cache        map[string]UfopItem
	mutex        sync.RWMutex
}

func NewUfopMgr(lister UfopLister, listInterval int64) *UfopMgr {

	if listInterval == 0 {
		listInterval = DefaultListInterval
	}
	interval := time.Duration(listInterval) * time.Second

	um := &UfopMgr{
		UfopLister:   lister,
		ListInterval: interval,
		Cache:        make(map[string]UfopItem),
	}
	go func() {
		um.doList()
		for {
			<-time.After(um.ListInterval)
			um.doList()
		}
	}()
	return um
}

func (um *UfopMgr) doList() {

	xl := xlog.NewDummy()
	ret, err := um.List(xl)
	if err != nil {
		xl.Error("ufop list failed -", err)
		return
	}
	xl.Debug("ufop acl info list got =>", ret.Entries)

	var aclList map[uint32]bool
	cache := make(map[string]UfopItem)

	for _, entry := range ret.Entries {
		if entry.AclMode != 0 {
			aclList = sliceToMap(entry.AclList)
		}
		cache[entry.Ufop] = UfopItem{
			AclMode: entry.AclMode,
			Url:     entry.Url,
			Method:  entry.Method,
			AclList: aclList,
		}
	}

	um.mutex.Lock()
	um.Cache = cache
	um.mutex.Unlock()
}

func (um *UfopMgr) IsValidUfop(fop string, uid uint32) bool {

	um.mutex.RLock()
	defer um.mutex.RUnlock()

	item, ok := um.Cache[fop]
	if !ok {
		return false
	}
	if item.AclMode == 0 {
		return true
	} else {
		_, ok = item.AclList[uid]
		return ok
	}
	return false
}

func (um *UfopMgr) GetUrlByFop(fop string) (url string, ifexist bool) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	item, ok := um.Cache[fop]
	if ok {
		url = item.Url
		ifexist = true
		return
	}
	return
}

func (um *UfopMgr) GetMethodByFop(fop string) (method byte, ifexist bool) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	item, ok := um.Cache[fop]
	if ok {
		method = item.Method
		ifexist = true
		return
	}
	return
}

func sliceToMap(s []uint32) (m map[uint32]bool) {
	if len(s) == 0 {
		return
	}

	m = make(map[uint32]bool)
	for _, z := range s {
		m[z] = true
	}
	return
}
