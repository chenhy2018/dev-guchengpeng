package mgoconv

import (
	"github.com/qiniu/errors"
	"launchpad.net/mgo"
)

type M map[string]interface{}

// ------------------------------------------------------------------------

func EnsureIndex(flagc *mgo.Collection) (err error) {

	return flagc.EnsureIndex(mgo.Index{Key: []string{"conv"}})
}

// ------------------------------------------------------------------------
// 本模块用于数据库发生非兼容性修改后，升级表结构之用。
// 基本思路是引入一个 conv 状态表，如果这个表里面的 flag 存在，则表示升级已经完成，否则代表升级没有做或者做失败了。

func Do(flagc *mgo.Collection, convName string, converter func() error) (err error) {

	flagEntry := M{"conv": convName}
	flag, err := flagc.Find(flagEntry).Count()
	if err != nil {
		err = errors.Info(err, "mgoconv.Do: flagc.Count failed").Detail(err)
		return
	}
	if flag > 0 {
		return nil
	}

	err = converter()
	if err != nil {
		err = errors.Info(err, "mgoconv.Do: convert failed").Detail(err)
		return
	}

	err = flagc.Insert(flagEntry)
	if err != nil {
		err = errors.Info(err, "mgoconv.Do: flagc.Insert failed").Detail(err)
	}
	return
}

// ------------------------------------------------------------------------
