package sessions

import (
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache.20160421/memcache"
	"gopkg.in/mgo.v2/bson"

	"qbox.us/biz/utils.v2"
)

type mcStore struct {
	Sid       string                 `bson:"-"`
	Values    map[string]interface{} `bson:"values"`
	CreatedAt time.Time              `bson:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at"`
}

func newMcStore(sid string, params ...map[string]interface{}) *mcStore {
	store := &mcStore{
		Sid:       sid,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if len(params) > 0 {
		store.Values = params[0]
	}

	if store.Values == nil {
		store.Values = make(map[string]interface{})
	}
	return store
}

type mcData struct {
	store    *mcStore
	provider *McProvider

	changed bool // flag when values has changed
	closed  bool // flag when data destroy
}

var _ SessionStore = new(mcData)

func newMcData(p *McProvider, store *mcStore, changed bool) *mcData {
	sess := &mcData{
		store:    store,
		provider: p,
		changed:  changed,
	}

	if store.Values == nil {
		store.Values = make(map[string]interface{})
		sess.changed = true
	}

	return sess
}

// get session id
func (m *mcData) Sid() string {
	return m.store.Sid
}

func (m *mcData) Set(key string, value interface{}) {
	m.changed = true
	m.store.Values[key] = value
}

func (m *mcData) Get(key string) *utils.Value {
	if v, ok := m.store.Values[key]; ok {
		return utils.ValueTo(v)
	}
	return utils.ValueTo(nil)
}

func (m *mcData) Delete(key string) {
	m.changed = true
	delete(m.store.Values, key)
}

func (m *mcData) Has(key string) bool {
	_, ok := m.store.Values[key]
	return ok
}

// duplicate all values
func (m *mcData) Values() map[string]interface{} {
	values := make(map[string]interface{}, len(m.store.Values))
	for key, value := range m.store.Values {
		values[key] = value
	}
	return values
}

// clear all values in session
func (m *mcData) Clean() {
	m.changed = true
	m.store.Values = make(map[string]interface{})
}

// save session values to store
func (m *mcData) Flush() error {
	// has destroy
	if m.closed {
		return nil
	}

	// no changes
	if !m.changed {
		return nil
	}

	m.changed = false
	return m.provider.save(m.store)
}

// destory session values in store
func (m *mcData) Destroy() error {
	m.closed = true
	m.store.Values = make(map[string]interface{})
	return m.provider.Destroy(m.store.Sid)
}

func (m *mcData) Touch() (err error) {
	expireAt := int(time.Now().Unix()) + m.provider.config.SessionExpire

	key := fmt.Sprintf("%s_%s", m.provider.keyPrefix, m.store.Sid)
	err = m.provider.mc.Touch(key, int32(expireAt))
	return
}

type McProvider struct {
	config    *Config
	mc        *memcache.Client
	keyPrefix string
}

var _ SessionProvider = new(McProvider)

func NewMcProvider(config Config, mc *memcache.Client, keyPrefix string) (sess SessionProvider) {
	provider := new(McProvider)
	provider.config = &config
	provider.mc = mc
	provider.keyPrefix = keyPrefix

	return provider
}

func (p *McProvider) Create(sid string, params ...map[string]interface{}) (sess SessionStore, err error) {
	key := fmt.Sprintf("%s_%s", p.keyPrefix, sid)
	store := newMcStore(sid, params...)

	value, err := bson.Marshal(store)
	if err != nil {
		return
	}

	err = p.mc.Add(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(p.config.SessionExpire),
	})
	if err != nil {
		// key 存在
		if err == memcache.ErrNotStored {
			err = ErrDuplicateSid
		}
		return
	}

	sess = newMcData(p, store, len(params) > 0)

	return
}

func (p *McProvider) Read(sid string) (sess SessionStore, err error) {
	key := fmt.Sprintf("%s_%s", p.keyPrefix, sid)
	store := &mcStore{}

	item, err := p.mc.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			err = ErrNotFoundSession
		}
		return
	}

	err = bson.Unmarshal(item.Value, store)
	if err != nil {
		return
	}

	store.Sid = sid

	sess = newMcData(p, store, false)
	return
}

// 为现有的session数据更换sid
func (p *McProvider) Regenerate(old string, sid string) (sess SessionStore, err error) {
	keyOld := fmt.Sprintf("%s_%s", p.keyPrefix, old)
	keyNew := fmt.Sprintf("%s_%s", p.keyPrefix, sid)

	// 验证旧的sid
	item, err := p.mc.Get(keyOld)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			err = ErrNotFoundSession
		}
		return
	}

	store := &mcStore{}
	err = bson.Unmarshal(item.Value, store)
	if err != nil {
		return
	}
	store.Sid = sid
	store.UpdatedAt = time.Now()

	newVlaue, err := bson.Marshal(store)
	if err != nil {
		return
	}

	err = p.mc.Add(&memcache.Item{
		Key:        keyNew,
		Value:      newVlaue,
		Expiration: int32(p.config.SessionExpire),
	})
	if err != nil {
		if err == memcache.ErrNotStored {
			err = ErrDuplicateSid
		}
		return
	}

	// 删除旧的key
	p.Destroy(keyOld)

	return p.Read(sid)
}

func (p *McProvider) Destroy(sid string) (err error) {
	key := fmt.Sprintf("%s_%s", p.keyPrefix, sid)
	err = p.mc.Delete(key)
	return
}

func (p *McProvider) GC() (err error) {
	//  不需要
	return
}

func (p *McProvider) Config() *Config {
	config := *(p.config)
	return &config
}

func (p *McProvider) save(store *mcStore) (err error) {
	key := fmt.Sprintf("%s_%s", p.keyPrefix, store.Sid)
	store.UpdatedAt = time.Now()

	value, err := bson.Marshal(store)
	if err != nil {
		return
	}

	err = p.mc.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(p.config.SessionExpire),
	})

	return
}
