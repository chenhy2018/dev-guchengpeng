/*
Modifyied by Konglingtao in 2015/7/23
To:
	Remove UC's cache to maintain consistency.
*/
package configuration

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/qiniu/xlog.v1"
	"labix.org/v2/mgo"
)

// GrpBucketInfo is the grp value of bucket info
var GrpBucketInfo = "pubinfo"

// use collection: qbox_uc.uc
type OldInstanceNC struct {
	*mgo.Collection
	Refresher Refresher
}

// use collection: qbox_bucket.bucket
type NewInstanceNC struct {
	*mgo.Collection
	Refresher Refresher
}

func OldNC(c *mgo.Collection, refresher Refresher) (p *OldInstanceNC, err error) {
	index := mgo.Index{
		Key:    []string{"grp", "key"},
		Unique: true,
	}
	err = c.EnsureIndex(index)
	if err != nil {
		return
	}
	p = &OldInstanceNC{c, refresher}
	return
}

// index has ensured
func NewNC(c *mgo.Collection, refresher Refresher) (p *NewInstanceNC, err error) {
	p = &NewInstanceNC{c, refresher}
	return
}

func (p *OldInstanceNC) Put(xl *xlog.Logger, grp, key, val string) (err error) {
	_, err = p.Upsert(M{"grp": grp, "key": key}, M{"grp": grp, "key": key, "val": val})
	if err != nil {
		xl.Warn("Write Database Error:", err, " uc:"+grp+"key:"+key)
	} else {
		p.refresh(xl, grp, key)
	}
	return
}

func (p *NewInstanceNC) Put(xl *xlog.Logger, uid uint32, tbl string, val string) (err error) {
	query := M{"uid": uid, "tbl": tbl}
	update := M{"$set": M{"val": val}}
	err = p.Update(query, update)
	if err != nil {
		xl.Warn("Write Database Error:", err, " uid:", uid, "tbl:", tbl)
	} else {
		p.refresh(xl, GrpBucketInfo, MakeUCKey(uid, tbl))
	}
	return
}

func (p *OldInstanceNC) Get(xl *xlog.Logger, grp, key string) (val string, err error) {

	var e entry
	err = p.Find(M{"grp": grp, "key": key, "fetched": nil}).One(&e)
	if err != nil {
		xl.Warn("Get Database Error:", err, " uc:"+grp+"key:"+key)
	}
	val = e.Val
	return
}

// return err=nil if bucket not set uc
func (p *NewInstanceNC) Get(xl *xlog.Logger, uid uint32, tbl string) (val string, err error) {

	var e struct {
		Val string `json:"val"`
	}
	query := M{"uid": uid, "tbl": tbl, "drop": 0}
	err = p.Find(query).Select(M{"val": 1}).One(&e)
	if err != nil {
		xl.Warn("Get Database Error:", err, "uid:", uid, "tbl:", tbl)
	}
	val = e.Val
	return
}

func (p *OldInstanceNC) Group(xl *xlog.Logger, grp string) (items []GroupItem, err error) {

	var entrylist []entry

	err = p.Find(M{"grp": grp, "fetched": nil}).All(&entrylist)
	if err != nil {
		xl.Warn(" Get Group Error:", err, " uc:"+grp)
		return
	}

	n := len(entrylist)
	items = make([]GroupItem, n)
	for i := 0; i < n; i++ {
		items[i] = GroupItem{entrylist[i].Key, entrylist[i].Val}
	}
	return
}

type ret struct {
	Uid uint32 `bson:"uid"`
	Tbl string `bson:"tbl"`
	Val string `bson:"val"`
}

// Warning: should not call this in product
// list all bucket uc Val except unset
func (p *NewInstanceNC) Group(xl *xlog.Logger) (items []GroupItem, err error) {

	var rets []ret
	err = p.Find(M{"drop": 0}).All(&rets)
	if err != nil {
		xl.Warn("Get Group Error:", err)
		return
	}
	for _, ret := range rets {
		if ret.Val != "" {
			key := MakeUCKey(ret.Uid, ret.Tbl)
			item := GroupItem{Key: key, Val: ret.Val}
			items = append(items, item)
		}
	}
	return
}

func (p *OldInstanceNC) Delete(xl *xlog.Logger, grp, key string) (err error) {

	err = p.Remove(M{"grp": grp, "key": key})
	if err != nil {
		xl.Warn("Delete Database Error:", err, " uc:"+grp+"key:"+key)
	}
	if err == nil || err == mgo.ErrNotFound {
		p.refresh(xl, grp, key)
	}
	return
}

func (p *NewInstanceNC) Delete(xl *xlog.Logger, uid uint32, tbl string) (err error) {
	query := M{"uid": uid, "tbl": tbl}
	update := M{"$unset": M{"val": 1}}
	err = p.Update(query, update)
	if err != nil {
		xl.Warn("Write Database Error:", err, " uid:", uid, "tbl:", tbl)
	} else {
		p.refresh(xl, GrpBucketInfo, MakeUCKey(uid, tbl))
	}
	return
}

// Refresh refresh caches about uc.
func Refresh(xl *xlog.Logger, refresher Refresher, grp, key string) {
	if refresher != nil {
		var keys = []string{
			"grp:" + grp,
			"uc:" + grp + ":" + key,
			"bucketInfo:" + key,
		}
		var wg sync.WaitGroup
		for _, key := range keys {
			wg.Add(1)
			go func(xl *xlog.Logger) {
				defer wg.Done()
				if err := refresher.Refresh(xl, key); err != nil {
					xl.Warn("refresh failed id:", key)
				}
			}(xl.Spawn())
		}
		wg.Wait()
	}
}

func (p *OldInstanceNC) refresh(xl *xlog.Logger, grp, key string) {
	Refresh(xl, p.Refresher, grp, key)
}

func (p *NewInstanceNC) refresh(xl *xlog.Logger, grp, key string) {
	Refresh(xl, p.Refresher, grp, key)
}

func ParseUCKey(key string) (uid uint32, tbl string, err error) {

	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		err = errors.New("key cannot split into 2")
		return
	}
	uid64, err := strconv.ParseUint(parts[0], 36, 32)
	if err != nil {
		return
	}
	uid = uint32(uid64)
	tbl = parts[1]
	return
}

func MakeUCKey(uid uint32, tbl string) (key string) {

	key = strconv.FormatUint(uint64(uid), 36) + ":" + tbl
	return
}
