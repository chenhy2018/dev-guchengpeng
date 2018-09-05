package mgocon

import (
	"sync"
	"time"

	"qbox.us/biz/utils.v2/log"
	"qbox.us/mgo2"
)

var (
	firstTried = map[MongoDB]bool{}
	cachedDBs  = map[MongoDB]*mgo2.Database{}
)

func ConnectDB(mongo MongoDB, url string) {
	var mutex sync.Mutex
	MongoDBs[mongo] = func() *mgo2.Database {
		db := cachedDBs[mongo]
		if db != nil {
			return db
		}

		if firstTried[mongo] {
			// 第一次失败的话，可能证明数据库产生了问题
			// 所以后续的连接使用异步，不阻塞客户端，把错误抛出去
			go func() {
				mutex.Lock()
				defer mutex.Unlock()

				// 锁解开以后判断是否已连接
				if cachedDBs[mongo] != nil {
					return
				}

				connectDB(mongo, url)
			}()

			db := cachedDBs[mongo]
			if db != nil {
				return db
			}
			return nil
		}

		// 初次连接，全局锁住等待
		mutex.Lock()
		defer mutex.Unlock()

		// 初次已跑完
		if firstTried[mongo] {
			return cachedDBs[mongo]
		}

		db = connectDB(mongo, url)

		// 连接完，写入标志
		firstTried[mongo] = true

		return db
	}
}

func connectDB(mongo MongoDB, url string) *mgo2.Database {
	db, err := mgo2.NewDatabaseWithTimeoutNoFatal(url, "strong", 3*time.Second)
	if err != nil {
		log.X.Error("mgo connect:", mongo.String(), err)
	}
	cachedDBs[mongo] = db
	return db
}
