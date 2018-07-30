package mgocon

import (
	"time"

	"github.com/teapots/teapot"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"qbox.us/biz/utils.v2/log"
	"qbox.us/mgo2"
)

const (
	slowQueryTime = 20 * time.Millisecond
)

type ConnectorInvoker func(func(*mgo.Collection) error) error

type Connector struct {
	dbName MongoDB
	Col    bson.M
	Invoke ConnectorInvoker
}

func NewConnector(dbName MongoDB, col bson.M) Connector {
	return Connector{
		dbName: dbName,
		Col:    col,
		Invoke: func(callback func(*mgo.Collection) error) error {
			var db *mgo2.Database
			var c *mgo.Collection

			getDB := MongoDBs[dbName]
			if getDB != nil {
				db = getDB()
			}

			if db == nil {
				return ErrNotConnected
			}

			now := time.Now()

			c = mgo2.CopyCollection(db.C(col))
			defer func() {
				execute := time.Since(now)
				if execute > slowQueryTime {
					log.X.Info("slow_mgo_query:", execute, teapot.LineOpt{Depth: 3})
				}

				mgo2.CloseCollection(c)
			}()

			ret := callback(c)
			return parseMgoError(ret)
		},
	}
}

func (d *Connector) Insert(model interface{}) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.Insert(model)
	})
	return
}

func (d *Connector) Upsert(selector interface{}, model interface{}) (err error) {
	err = d.Invoke(func(c *mgo.Collection) (err error) {
		_, err = c.Upsert(selector, model)
		return
	})
	return
}

func (d *Connector) UpdateId(id bson.ObjectId, model interface{}) (err error) {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidId
	}
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.UpdateId(id, model)
	})
	return
}

func (d *Connector) UpdateIdSet(id bson.ObjectId, change bson.M) (err error) {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidId
	}
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.Update(bson.M{
			"_id": id,
		}, bson.M{
			"$set": change,
		})
	})
	return
}

func (d *Connector) Update(query, change bson.M) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.Update(query, change)
	})
	return
}

func (d *Connector) UpdateSet(query, change bson.M) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.Update(query, bson.M{
			"$set": change,
		})
	})
	return
}

func (d *Connector) UpdateAll(query, change bson.M) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		_, er := c.UpdateAll(query, change)
		return er
	})
	return
}

func (d *Connector) Find(query bson.M, model interface{}) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.Find(query).One(model)
	})
	return
}

func (d *Connector) FindId(id bson.ObjectId, model interface{}) (err error) {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidId
	}
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.FindId(id).One(model)
	})
	return
}

func (d *Connector) FindAll(query bson.M, models interface{}, skip, limit int, sorts ...string) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.Find(query).Skip(skip).Limit(limit).Sort(sorts...).All(models)
	})
	return
}

func (d *Connector) Distinct(query bson.M, models interface{}, key string) (err error) {
	err = d.Invoke(func(c *mgo.Collection) (er error) {
		return c.Find(query).Distinct(key, models)
	})
	return
}

// atomic update object and return old object
func (d *Connector) FindAndModify(query, change bson.M, model interface{}) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		_, er := c.Find(query).Apply2(mgo.Change{
			Update: change,
		}, model)
		return er
	})
	return
}

func (d *Connector) Remove(query bson.M) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.Remove(query)
	})
	return
}

func (d *Connector) RemoveId(id bson.ObjectId) (err error) {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidId
	}
	err = d.Invoke(func(c *mgo.Collection) error {
		return c.RemoveId(id)
	})
	return
}

func (d *Connector) RemoveAll(query bson.M) (err error) {
	err = d.Invoke(func(c *mgo.Collection) error {
		_, er := c.RemoveAll(query)
		return er
	})
	return
}

func (d *Connector) Count(query bson.M) (cnt int, err error) {
	err = d.Invoke(func(c *mgo.Collection) (er error) {
		cnt, er = c.Find(query).Count()
		return er
	})
	return
}

func parseMgoError(err error) error {
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}

	if mgo.IsDup(err) {
		return ErrDuplicateKey
	}

	return err
}
