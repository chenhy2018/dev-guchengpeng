package biproxy

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleHive struct {
	Host   string
	Client *rpc.Client
}

func NewHandleHive(host string, client *rpc.Client) *HandleHive {
	return &HandleHive{host, client}
}

func (r HandleHive) TableAll(logger rpc.Logger) (resp []RespDB, err error) {
	err = r.Client.Call(logger, &resp, r.Host+"/hive/table/all")
	return
}

func (r HandleHive) AsynQuery(logger rpc.Logger, req ReqQuery) (id string, err error) {
	value := url.Values{}
	value.Add("db", req.DB)
	value.Add("query", req.Query)
	err = r.Client.Call(logger, &id, r.Host+"/hive/asyn/query?"+value.Encode())
	return
}

func (r HandleHive) QueryCache(logger rpc.Logger, req ReqId) (cache QueryCache, err error) {
	value := url.Values{}
	value.Add("id", req.Id)
	err = r.Client.Call(logger, &cache, r.Host+"/hive/query/cache?"+value.Encode())
	return
}

func (r HandleHive) PersistenceCache(logger rpc.Logger, req ReqPersistence) (err error) {
	value := url.Values{}
	value.Add("id", req.Id)
	value.Add("name", req.Name)
	value.Add("desc", req.Desc)
	err = r.Client.Call(logger, nil, r.Host+"/hive/persistence/cache?"+value.Encode())
	return
}

func (r HandleHive) DelQueryCache(logger rpc.Logger, req ReqId) (err error) {
	value := url.Values{}
	value.Add("id", req.Id)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/hive/del/query/cache", map[string][]string(value))
	return
}

func (r HandleHive) SqlList(logger rpc.Logger, req ReqSQLList) (sqls []RespSQL, err error) {
	value := url.Values{}
	value.Add("index", strconv.FormatInt(int64(req.Index), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("sortby", req.SortBy)
	err = r.Client.Call(logger, &sqls, r.Host+"/hive/sql/list?"+value.Encode())
	return
}
