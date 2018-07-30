package mockmq

import (
	"bytes"
	"io"
	"net/http"
	"qbox.us/api"
	"qbox.us/errors"
	"github.com/qiniu/log.v1"
	"qbox.us/net/httputil"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Mesg struct {
	ctx         []byte
	lastGetTime int64
}

type Queue struct {
	msgs    []*Mesg
	freeIds []int
	expires int64
}

type MockMQ struct {
	mq map[string]*Queue
	sync.Mutex
}

func New() *MockMQ {
	return &MockMQ{mq: make(map[string]*Queue)}
}

// POST /make/<MQID>/expires/<ExpiresInSeconds>
//
// 200 OK
func (p *MockMQ) make(w http.ResponseWriter, req *http.Request) {

	log.Info("MockMQ.make")

	var err error

	querys := strings.Split(req.URL.Path[1:], "/")
	if len(querys) != 4 || querys[2] != "expires" {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	mqid := querys[1]

	expires, err := strconv.ParseInt(querys[3], 10, 64)
	if err != nil {
		err = errors.Info(err, "MockMQ.make: strconv.ParseInt failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	p.Lock()
	defer p.Unlock()

	_, ok := p.mq[string(mqid)]
	if ok {
		httputil.ReplyWithCode(w, 200)
		return
	}

	q := &Queue{msgs: make([]*Mesg, 0), freeIds: make([]int, 0), expires: expires}
	p.mq[string(mqid)] = q
	httputil.ReplyWithCode(w, 200)
}

// POST /put/<MQID>
// Content-Type: application/octet-stream
// Body: <Message>
//
// 200 OK
// X-Id: <MessageId>
func (p *MockMQ) put(w http.ResponseWriter, req *http.Request) {

	log.Info("MockMQ.put")

	var err error

	querys := strings.Split(req.URL.Path[1:], "/")
	if len(querys) != 2 {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	mqid := querys[1]

	if req.ContentLength < 0 {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	m := make([]byte, req.ContentLength)
	_, err = io.ReadFull(req.Body, m)
	if err != nil {
		err = errors.Info(err, "MockMQ.put: ioutil.ReadFull failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	p.Lock()
	defer p.Unlock()

	q, ok := p.mq[string(mqid)]
	if !ok {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	var id int

	lens := len(q.freeIds)
	if lens > 0 {
		id = q.freeIds[lens-1]
		q.freeIds = q.freeIds[:lens-1]
		q.msgs[id] = &Mesg{ctx: m}
	} else {
		id = len(q.msgs)
		q.msgs = append(q.msgs, &Mesg{ctx: m})
	}

	w.Header().Set("X-Id", strconv.Itoa(id))
	httputil.ReplyWithCode(w, 200)
}

// POST /get/<MQID>
//
// 200 OK
// Content-Type: application/octet-stream
// X-Id: <MessageId>
// Body: <Message>
func (p *MockMQ) get(w http.ResponseWriter, req *http.Request) {

	log.Info("MockMQ.get")

	querys := strings.Split(req.URL.Path[1:], "/")
	if len(querys) != 2 {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	mqid := querys[1]

	p.Lock()

	q, ok := p.mq[string(mqid)]
	if !ok {
		p.Unlock()
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	var id int = -1
	var ctx []byte

	// no efficiency
	for i, m := range q.msgs {
		if m != nil {
			now := time.Now().Unix()
			if now-m.lastGetTime < q.expires {
				continue
			}
			id = i
			m.lastGetTime = now
			ctx = m.ctx
			break
		}
	}

	p.Unlock()

	if id < 0 {
		httputil.ReplyWithCode(w, 404)
		return
	}

	w.Header().Set("X-Id", strconv.Itoa(id))
	w.Header().Set("Content-Length", strconv.Itoa(len(ctx)))
	io.Copy(w, bytes.NewReader(ctx))
}

// POST /delete/<MQID>
//
// 200 OK
func (p *MockMQ) del(w http.ResponseWriter, req *http.Request) {

	log.Info("MockMQ.del")

	var err error

	querys := strings.Split(req.URL.Path[1:], "/")
	if len(querys) != 2 {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	mqid := querys[1]

	idStr := req.Header.Get("X-Id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		err = errors.Info(err, "MockMQ.del: strconv.ParseInt failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	p.Lock()
	defer p.Unlock()

	q, ok := p.mq[string(mqid)]
	if !ok {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	if id < 0 || int(id) > len(q.msgs)-1 {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	q.msgs[id] = nil
	httputil.ReplyWithCode(w, 200)
}

func (p *MockMQ) RegisterHandlers(mux *http.ServeMux) {

	mux.HandleFunc("/make/", func(w http.ResponseWriter, req *http.Request) { p.make(w, req) })
	mux.HandleFunc("/get/", func(w http.ResponseWriter, req *http.Request) { p.get(w, req) })
	mux.HandleFunc("/put/", func(w http.ResponseWriter, req *http.Request) { p.put(w, req) })
	mux.HandleFunc("/delete/", func(w http.ResponseWriter, req *http.Request) { p.del(w, req) })
}

func (p *MockMQ) Run(addr string) (err error) {

	mux := http.NewServeMux()
	p.RegisterHandlers(mux)
	err = http.ListenAndServe(addr, mux)
	return
}
