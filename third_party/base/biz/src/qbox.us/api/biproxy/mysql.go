package biproxy

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleMySQL struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMySQL(host string, client *rpc.Client) *HandleMySQL {
	return &HandleMySQL{host, client}
}

func (r HandleMySQL) TableAll(logger rpc.Logger) (resp []RespDB, err error) {
	err = r.Client.Call(logger, &resp, r.Host+"/mysql/table/all")
	return
}

func (r HandleMySQL) AsynQuery(logger rpc.Logger, req ReqQuery) (id string, err error) {
	value := url.Values{}
	value.Add("db", req.DB)
	value.Add("query", req.Query)
	err = r.Client.Call(logger, &id, r.Host+"/mysql/asyn/query?"+value.Encode())
	return
}

func (r HandleMySQL) QueryCache(logger rpc.Logger, req ReqId) (cache QueryCache, err error) {
	value := url.Values{}
	value.Add("id", req.Id)
	err = r.Client.Call(logger, &cache, r.Host+"/mysql/query/cache?"+value.Encode())
	return
}

func (r HandleMySQL) PersistenceCache(logger rpc.Logger, req ReqPersistence) (err error) {
	value := url.Values{}
	value.Add("id", req.Id)
	value.Add("name", req.Name)
	value.Add("desc", req.Desc)
	err = r.Client.Call(logger, nil, r.Host+"/mysql/persistence/cache?"+value.Encode())
	return
}

func (r HandleMySQL) DelQueryCache(logger rpc.Logger, req ReqId) (err error) {
	value := url.Values{}
	value.Add("id", req.Id)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/mysql/del/query/cache", map[string][]string(value))
	return
}

func (r HandleMySQL) SqlList(logger rpc.Logger, req ReqSQLList) (sqls []RespSQL, err error) {
	value := url.Values{}
	value.Add("index", strconv.FormatInt(int64(req.Index), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("sortby", req.SortBy)
	err = r.Client.Call(logger, &sqls, r.Host+"/mysql/sql/list?"+value.Encode())
	return
}
