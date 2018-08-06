package mmq

import (
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/xlog.v1"

	. "github.com/qiniu/ctype"
)

const (
	msgIdCType = URLSAFE_BASE64 | EQ
)

var (
	ErrMqNotFound   = httputil.NewError(612, "mq not found")
	ErrNothingTodo  = httputil.NewError(612, "nothing to do")
	ErrXidRequired  = httputil.NewError(400, "X-Id required")
	ErrInvalidXid   = httputil.NewError(400, "invalid X-Id")
	ErrInvalidMsgId = httputil.NewError(400, "invalid msgid: can only use urlsafe base64 characters")
)

// ------------------------------------------------------------------

type idArg struct {
	Id string `flag:"_"`
}

type msgArgs struct {
	MqId  string `flag:"_"`
	MsgId string `flag:"id"`
}

// ------------------------------------------------------------------

type MqConf struct {
	MqId     string `json:"id"`
	MqBufLen int    `json:"buflen"`  // 最大缓存任务数，超过这个任务数的时候，任务入队列会失败
	Expires  int    `json:"expires"` // 任务超时重做的超时时间，应该尽量大于任务的执行时间
	TryTimes int    `json:"try"`     // 任务超时重做的次数
}

type Config struct {
	Mqs []MqConf `json:"mqs"`
}

type Service struct {
	mqs  map[string]*Instance
	lock bool
}

func New(cfg *Config) *Service {

	mqs := make(map[string]*Instance)
	p := &Service{mqs: mqs}
	for _, conf := range cfg.Mqs {
		mq := NewInstance(conf.MqBufLen, uint32(conf.Expires), uint32(conf.TryTimes))
		mqs[conf.MqId] = mq
	}
	return p
}

// ------------------------------------------------------------------

//
// POST /put/<MqId>/id/<MsgId>
// Content-Type: application/octet-stream
//
// <MessageData>
//
func (p *Service) CmdPut_(args *msgArgs, env rpcutil.Env) (err error) {

	xl := xlog.New(env.W, env.Req)
	xl.Debug("mq.Put:", args.MsgId)

	if p.lock {
		return httputil.ErrGracefulQuit
	}

	if !IsType(msgIdCType, args.MsgId) {
		return ErrInvalidMsgId
	}

	mq, ok := p.mqs[args.MqId]
	if !ok {
		return ErrMqNotFound
	}

	msg, err := ioutil.ReadAll(env.Req.Body)
	if err != nil {
		return
	}

	return mq.TryPut(args.MsgId, msg)
}

// ------------------------------------------------------------------

//
// POST /get/<MqId>[/expires/<Expires>]
//
func (p *Service) CmdGet_(args *idArg, env rpcutil.Env) {

	xl := xlog.New(env.W, env.Req)
	xl.Debug("mq.Get")

	w := env.W
	mq, ok := p.mqs[args.Id]
	if !ok {
		httputil.Error(w, ErrMqNotFound)
		return
	}

	var cancel <-chan bool
	if cn, ok := getCloseNotifier(w); ok {
		cancel = cn.CloseNotify()
	}

	msgId, msg, err := mq.GetInf(cancel)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	h := w.Header()
	h.Set("X-Id", msgId)
	httputil.ReplyWith(w, 200, "application/octet-stream", msg)
}

func getCloseNotifier(w http.ResponseWriter) (cn http.CloseNotifier, ok bool) {

	v := reflect.ValueOf(w)
	v = reflect.Indirect(v)
	for v.Kind() == reflect.Struct {
		if fv := v.FieldByName("ResponseWriter"); fv.IsValid() {
			if cn, ok = fv.Interface().(http.CloseNotifier); ok {
				return
			}
			v = reflect.Indirect(fv)
		} else {
			break
		}
	}
	return
}

// ------------------------------------------------------------------

//
// POST /delete/<MqId>/id/<MsgId>
//
func (p *Service) CmdDelete_(args *msgArgs, env rpcutil.Env) (err error) {

	mq, ok := p.mqs[args.MqId]
	if !ok {
		return ErrMqNotFound
	}

	if args.MsgId == "" {
		args.MsgId = env.Req.Header.Get("X-Id")
	}
	return mq.Delete(args.MsgId, false)
}

// ------------------------------------------------------------------

//
// POST /stat/<MqId>
//
func (p *Service) CmdStat_(args *idArg) (ret StatInfo, err error) {

	mq, ok := p.mqs[args.Id]
	if !ok {
		err = ErrMqNotFound
		return
	}

	return mq.Stat(), nil
}

// ------------------------------------------------------------------

func (p *Service) Quit() {

	p.lock = true

	for p.cannotQuit() {
		time.Sleep(1e9)
	}
	os.Exit(0)
}

func (p *Service) cannotQuit() bool {

	for _, mq := range p.mqs {
		if mq.Len() != 0 {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------------
