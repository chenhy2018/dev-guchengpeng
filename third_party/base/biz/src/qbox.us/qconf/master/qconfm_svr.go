package master

import (
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"
	"labix.org/v2/mgo"
	account "qbox.us/http/account.v2"
	"qbox.us/qconf/master/qrefresher"
	. "qbox.us/qconf/master/qrefresher/proto"
)

var ErrUnacceptable = httputil.NewError(401, "bad token: unacceptable")
var ErrNoSuchEntry = httputil.NewError(612, "no such entry")

// ------------------------------------------------------------------------

type M map[string]interface{}

type keyArg struct {
	Id string `json:"id"`
}

type Config struct {
	Coll         *mgo.Collection
	MgrAccessKey string
	MgrSecretKey string     // 向 slave 发送指令时的帐号
	SlaveHosts   [][]string // [[idc1_slave1, idc1_slave2], [idc2_slave1, idc2_slave2], ...]

	AuthParser account.AuthParser

	UidMgr uint32 // 只接受这个管理员发过来的请求
}

type Service struct {
	account.Manager
	refresher Refresher
	Config
}

// ------------------------------------------------------------------------

func New(cfg *Config) (p *Service, err error) {

	p = &Service{Config: *cfg}

	if len(cfg.SlaveHosts) != 0 {
		p.refresher = qrefresher.New(cfg.MgrAccessKey, cfg.MgrSecretKey, cfg.SlaveHosts)
	} else {
		p.refresher = NilRefresher
	}

	p.InitAccount(cfg.AuthParser)
	return
}

func NewEx(cfg *Config, refresher Refresher) (p *Service, err error) {
	p = &Service{Config: *cfg, refresher: refresher}

	p.InitAccount(cfg.AuthParser)
	return
}

// ------------------------------------------------------------------------

/*
	POST /putb
	Content-Type: application/bson

	{id: <Id>, chg: <Change>}
*/
type putbArgs struct {
	Id     string      `bson:"id"`
	Change interface{} `bson:"chg"`
}

func (p *Service) BbrpcPutb(doc *putbArgs, env *account.AdminEnv) (err error) {

	log := xlog.New(env.W, env.Req)

	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}

	_, err = p.Coll.UpsertId(doc.Id, doc.Change)
	if err != nil {
		log.Error("qconf.Put: Upsert failed -", doc.Id, doc.Change, "Error:", err)
		return
	}

	return p.refresher.Refresh(log, doc.Id)
}

/*
	POST /insb
	Content-Type: application/bson

	{doc: <Doc>}
*/
type insbArgs struct {
	Doc interface{} `bson:"doc"`
}

func (p *Service) BbrpcInsb(args *insbArgs, env *account.AdminEnv) (err error) {

	log := xlog.New(env.W, env.Req)

	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}

	err = p.Coll.Insert(args.Doc)
	if err != nil {
		log.Error("qconf.Insert failed -", args.Doc, "Error:", err)
	}
	return
}

// 过时协议
/*
	POST /put
	Content-Type: application/json

	{id: <Id>, chg: <Change>}
*/
type putArgs struct {
	Id     string      `json:"id"`
	Change interface{} `json:"chg"`
}

func (p *Service) RpcPut(doc *putArgs, env *account.AdminEnv) (err error) {

	log := xlog.New(env.W, env.Req)

	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}

	_, err = p.Coll.UpsertId(doc.Id, doc.Change)
	if err != nil {
		log.Error("qconf.Put: Upsert failed -", doc.Id, doc.Change, "Error:", err)
		return
	}

	return p.refresher.Refresh(log, doc.Id)
}

// ------------------------------------------------------------------------

func (p *Service) WspRm(args *keyArg, env *account.AdminEnv) (err error) {

	log := xlog.New(env.W, env.Req)

	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}

	err = p.Coll.RemoveId(args.Id)
	if err != nil {
		log.Error("qconf.Remove failed -", args.Id, "Error:", err)
		return
	}

	return p.refresher.Refresh(log, args.Id)
}

// ------------------------------------------------------------------------

func (p *Service) WspRefresh(args *keyArg, env *account.AdminEnv) (err error) {

	log := xlog.New(env.W, env.Req)

	return p.refresher.Refresh(log, args.Id)
}

// ------------------------------------------------------------------------

var all = M{"_id": 0}

func (p *Service) WbrpcGetb(args *keyArg, env *account.AdminEnv) (doc M, err error) {

	if env.Uid != p.UidMgr {
		return nil, ErrUnacceptable
	}

	err = p.Coll.Find(M{"_id": args.Id}).Select(all).One(&doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = ErrNoSuchEntry
		}
		log := xlog.New(env.W, env.Req)
		log.Error("qconf.Get: Find failed -", args.Id, "Error:", err)
	}
	return
}

// 过时协议
func (p *Service) WspGet(args *keyArg, env *account.AdminEnv) (doc M, err error) {

	if env.Uid != p.UidMgr {
		return nil, ErrUnacceptable
	}

	err = p.Coll.Find(M{"_id": args.Id}).Select(all).One(&doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = ErrNoSuchEntry
		}
		log := xlog.New(env.W, env.Req)
		log.Error("qconf.Get: Find failed -", args.Id, "Error:", err)
	}
	return
}

// ------------------------------------------------------------------------
