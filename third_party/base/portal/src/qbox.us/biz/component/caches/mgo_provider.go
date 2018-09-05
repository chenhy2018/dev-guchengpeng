package caches

import (
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/qiniu/log.v1"

	"qbox.us/biz/utils.v2"
)

type MgoConfig struct {
	AutoExpire     bool
	KeyField       string
	ValueField     string
	ExpiredAtField string
}

func NewMgoConfig() MgoConfig {
	config := MgoConfig{
		KeyField:       "key",
		ValueField:     "value",
		ExpiredAtField: "expired_at",
	}

	return config
}

type MgoProvider struct {
	config  *MgoConfig
	connect func(func(c *mgo.Collection) error) error
}

var _ CacheProvider = new(MgoProvider)

func NewMgoProvider(config MgoConfig, connect func(func(c *mgo.Collection) error) error) (sess CacheProvider) {
	provider := new(MgoProvider)
	provider.config = &config
	provider.connect = connect

	connect(func(c *mgo.Collection) (err error) {
		defer func() {
			if err != nil {
				log.Error("NewMgoProvider, connect", err)
			}
		}()

		err = c.Database.Session.Ping()
		if err != nil {
			log.Error("NewMgoProvider, Ping", err)
			return
		}

		// unique index of key field
		if err = c.EnsureIndex(mgo.Index{Key: []string{config.KeyField}, Name: config.KeyField, Unique: true}); err != nil {
			log.Error("NewMgoProvider, c.EnsureIndex", err)
			return
		}

		expireIndex := mgo.Index{
			Name: config.ExpiredAtField,
			Key:  []string{config.ExpiredAtField},
		}

		if config.AutoExpire {
			// 创建 mgo 自动过期的索引
			expireIndex.ExpireAfter = time.Second
		}

		// index of expired_at field
		if err = c.EnsureIndex(expireIndex); err != nil {
			log.Error("NewMgoProvider, c.EnsureIndex", err)
			return
		}

		return
	})
	return provider
}

func (p *MgoProvider) Get(key string) *utils.Value {
	values, ok := p.fetchValues(key)
	if !ok {
		return utils.ValueTo(nil)
	}

	value, _ := p.getValue(values)
	return utils.ValueTo(value)
}

func (p *MgoProvider) Set(key string, val interface{}, params ...int) (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		var timeout time.Duration
		if len(params) > 0 {
			switch {
			case params[0] > 0:
				timeout = time.Duration(params[0]) * time.Second
			case params[0] == 0:
				timeout = time.Hour * 24 * 9999
			}
		}

		if timeout <= 0 {
			timeout = DEFAULT_TIMEOUT
		}

		expiredAt := time.Now().Add(timeout)

		_, er := c.Upsert(bson.M{
			p.config.KeyField: key,
		}, bson.M{
			"$set": bson.M{
				p.config.KeyField:       key,
				p.config.ValueField:     val,
				p.config.ExpiredAtField: expiredAt,
			},
		})
		return er
	})
	return
}

func (p *MgoProvider) Delete(key string) (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		er := c.Remove(bson.M{
			p.config.KeyField: key,
		})
		return er
	})
	return
}

func (p *MgoProvider) Incr(key string, params ...int) (err error) {
	cnt := 1
	if len(params) > 0 {
		cnt = params[0]
	}

	_, ok := p.fetchValues(key)
	if !ok {
		return ERR_MISSED_KEY
	}

	err = p.connect(func(c *mgo.Collection) error {
		er := c.Update(bson.M{
			p.config.KeyField: key,
		}, bson.M{
			"$inc": bson.M{
				p.config.ValueField: cnt,
			},
		})
		return er
	})
	return
}

func (p *MgoProvider) Decr(key string, params ...int) (err error) {
	cnt := -1
	if len(params) > 0 {
		cnt = 0 - params[0]
	}

	_, ok := p.fetchValues(key)
	if !ok {
		return ERR_MISSED_KEY
	}

	err = p.connect(func(c *mgo.Collection) error {
		er := c.Update(bson.M{
			p.config.KeyField: key,
		}, bson.M{
			"$inc": bson.M{
				p.config.ValueField: cnt,
			},
		})
		return er
	})
	return
}

func (p *MgoProvider) Has(key string) bool {
	values, ok := p.fetchValues(key)
	if !ok {
		return false
	}

	_, ok = p.getValue(values)
	return ok
}

func (p *MgoProvider) Clean() (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		_, er := c.RemoveAll(nil)
		return er
	})
	return
}

func (p *MgoProvider) GC() (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		_, er := c.RemoveAll(bson.M{
			p.config.ExpiredAtField: bson.M{
				"$lte": time.Now(),
			},
		})
		return er
	})
	return
}

func (p *MgoProvider) fetchValues(key string) (values map[string]interface{}, ok bool) {
	p.connect(func(c *mgo.Collection) error {
		er := c.Find(bson.M{
			p.config.KeyField: key,
		}).One(&values)
		return er
	})
	return values, values != nil
}

func (p *MgoProvider) getValue(values map[string]interface{}) (value interface{}, ok bool) {
	if values == nil {
		return
	}

	// get expired time
	expiredAt, exists := values[p.config.ExpiredAtField].(time.Time)
	if exists {
		// not expired yet
		if time.Now().Before(expiredAt) {

			// get cached value
			value, ok = values[p.config.ValueField]
		}
	}
	return
}
