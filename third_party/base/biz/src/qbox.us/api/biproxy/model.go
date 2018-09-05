package biproxy

import (
	"labix.org/v2/mgo/bson"
	"strconv"
	"time"
)

type RespDB struct {
	Name   string      `json:"name"`
	Tables []RespTable `json:"tables"`
}

type RespTable struct {
	Name   string      `json:"name"`
	Desc   string      `json:"desc"`
	Fields []RespField `json:"fields"`
}

type RespField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Desc string `json:"desc"`
}

type ReqQuery struct {
	DB    string `json:"db"`
	Query string `json:"query"`
}

type RespQuery struct {
	Name string        `json:"name"`
	Data []interface{} `json:"data"`
}

type Status string

const (
	Success Status = "success"
	Failed  Status = "failed"
	Running Status = "running"
)

type DbType string

const (
	MYSQL DbType = "mysql"
	HIVE  DbType = "hive"
)

type Second int64

func (self Second) ToString() string {
	return strconv.FormatInt(int64(self), 10)
}

type QueryCache struct {
	Id          bson.ObjectId `bson:"_id" json:"id"`
	DbType      DbType        `json:"dbtype"`
	Sql         string        `json:"sql"`
	Data        []RespQuery   `json:"data"`
	Status      Status        `json:"status"`
	Detail      string        `json:"detail"`
	Persistence bool          `json:"persistent"`
	Name        string        `json:"name"`
	Desc        string        `json:"desc"`
	CreateAt    Second        `json:"createat"`
	UpdateAt    Second        `json:"updateat"`
}

type RespSQL struct {
	Id       bson.ObjectId `bson:"_id" json:"id"`
	DbType   DbType        `json:"dbtype"`
	Sql      string        `json:"sql"`
	Name     string        `json:"name"`
	Desc     string        `json:"desc"`
	CreateAt Second        `json:"createat"`
	UpdateAt Second        `json:"updateat"`
}

type ReqId struct {
	Id string `json:"id"`
}

type ReqPersistence struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type ReqSQLList struct {
	Index  int    `json:"index"`
	Limit  int    `json:"limit"`
	SortBy string `json:"sortby"`
}

//qjob
type ReqTaskQuery struct {
	Day    *string `json:"day"`
	Status *string `json:"status"`
}

type ReqOpQuery struct {
	Task string `json:"task"`
	Day  string `json:"day"`
}

type ReqTask struct {
	Task string `json:"task"`
	Day  string `json:"day"`
}

type ReqOperation struct {
	Day string `json:"day"`
	Key string `json:"key"`
}

type TaskStatus struct {
	Task   string    `json:"task" bson:"task"`
	Status string    `json:"status" bson:"status"`
	Day    string    `json:"day" bson:"day"`
	Round  int       `json:"round" bson:"round"`
	Key    string    `json:"key" bson:"key"`
	Begin  time.Time `json:"begin" bson:"begin"`
	End    time.Time `json:"end" bson:"end"`
}

type TaskOperation struct {
	Depth int      `json:"depth" bson:"depth"`
	Task  string   `json:"task" bson:"task"`
	Check bool     `json:"check" bson:"check"`
	After []string `json:"after" bson:"after"`
	Op    string   `json:"op" bson:"op"`
}

type Ops struct {
	Name   string       `json:"name"`
	Status []TaskStatus `json:"status"`
}
