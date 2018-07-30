package clients

import (
	"encoding/base64"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"io"
	"net/http"
	gourl "net/url"
	"qbox.us/errors"
	"qbox.us/fop"
	"qbox.us/qconf/qconfapi"
	qufop "qbox.us/ufop"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var default_max_timeout int64 = 300

type UfopdConfig struct {
	Gates         []string         `json:"ufop_gates"`
	RetryTimes    int64            `json:"ufop_retry_times"`
	RetryInterval int64            `json:"ufop_retry_interval"`
	MaxTimeout    int64            `json:"ufop_max_timeout"`
	QconfCfg      *qconfapi.Config `json:"ufop_qconf"`
}

type Ufopd struct {
	lock  *sync.Mutex
	conns []*UfopdConn
	qconf *qconfapi.Client
	*UfopdConfig
}

func NewUfopd(config *UfopdConfig) *Ufopd {
	ufopd := &Ufopd{&sync.Mutex{}, make([]*UfopdConn, 0), qconfapi.New(config.QconfCfg), config}
	for _, gate := range config.Gates {
		conn := NewUfopdConn(gate, config.MaxTimeout)
		ufopd.conns = append(ufopd.conns, conn)
	}

	return ufopd
}

func (p *Ufopd) Op(xl *xlog.Logger, f io.Reader, fsize int64, fopCtx *fop.FopCtx, client *http.Client) (r *http.Response, err error) {

	cmd := strings.SplitN(fopCtx.RawQuery, "/", 2)[0]
	//for all gates, they must have share the same ufop command sets
	conn, err := p.pickConn(nil)
	if err != nil {
		err = errors.New("no gate available")
		return
	}

	var retry bool
	var i int64
	var uappinfo *qufop.UappInfo
	var ufopinfo *qufop.UfopInfo
	for i = 0; i < p.RetryTimes; i++ {
		ufopinfo, err = p.getUfopInfo(xl, cmd)
		if err != nil {
			break
		}
		uappinfo, err = p.getUappInfo(xl, ufopinfo.Uapp)
		if err != nil {
			break
		}
		r, retry, err = conn.Op(xl, f, fsize, fopCtx, client, uappinfo)
		if !retry {
			break
		} else {
			conn, err = p.pickConn(conn)
			if err != nil {
				break
			}
		}
	}
	return
}

func (p *Ufopd) HasCmd(xl *xlog.Logger, ufop string) bool {
	id := qufop.QconfUfopID(ufop)
	if err := p.qconf.Get(xl, &qufop.UfopInfo{}, id, qconfapi.Cache_NoSuchEntry); err != nil {
		//no matter what error is, we regard it as ufop not exist
		xl.Error("ufopd qconf:", err.Error())
		return false
	} else {
		return true
	}
}

func (p *Ufopd) PermmitUseUfop(xl *xlog.Logger, user uint32, ufop string) bool {
	ufopinfo, err := p.getUfopInfo(xl, ufop)
	if err != nil {
		xl.Error("ufopd.PermitUseUfop: getUfopInfo", err.Error())
		return false
	}
	uappinfo, err := p.getUappInfo(xl, ufopinfo.Uapp)
	if err != nil {
		xl.Error("ufopd.PermitUseUfop: getUappInfo", err.Error())
		return false
	}
	if uappinfo.Mode == 1 {
		return true
	} else if uappinfo.Mode == 0 && user == uappinfo.Uid {
		return true
	} else if uappinfo.Mode == 2 {
		for _, uid := range uappinfo.AccList {
			if uid == user {
				return true
			}
		}
	}
	return false
}

func (p *Ufopd) NeedCache(xl *xlog.Logger, ufop string) (need bool, err error) {
	var ufopinfo *qufop.UfopInfo
	ufopinfo, err = p.getUfopInfo(xl, ufop)
	if err != nil {
		return
	} else {
		if ufopinfo.DiskCache == qufop.ENABLE_DISKCACHE {
			need = true
		}
	}
	return
}

//----------------------------------------------------------------------------------------------------
func (p *Ufopd) getUfopInfo(xl *xlog.Logger, ufop string) (ui *qufop.UfopInfo, err error) {
	id := qufop.QconfUfopID(ufop)
	ui = &qufop.UfopInfo{}
	err = p.qconf.Get(xl, ui, id, qconfapi.Cache_NoSuchEntry)
	return
}

func (p *Ufopd) getUappInfo(xl *xlog.Logger, uapp string) (ui *qufop.UappInfo, err error) {
	id := qufop.QconfUappID(uapp)
	ui = &qufop.UappInfo{}
	err = p.qconf.Get(xl, ui, id, qconfapi.Cache_NoSuchEntry)
	return
}

func (p *Ufopd) pickConn(ignoreconn *UfopdConn) (*UfopdConn, error) {

	var min int64
	var idx int
	min, idx = -1, -1

	now := time.Now().Unix()
	for i, conn := range p.conns {
		n := conn.ProcessingNum()
		if min < 0 {
			idx = i
			min = n
		}
		if now-conn.LastFailTime() > p.RetryInterval {
			if n < min {
				if ignoreconn == nil || ignoreconn.gateaddr != conn.gateaddr {
					idx = i
					min = n
				}
			}
		}
	}
	if idx < 0 {
		return nil, errors.New("no conn avaiable")
	}

	return p.conns[idx], nil
}

//----------------------------------------------------------------------------------------------------
type UfopdConn struct {
	lastFailTime int64
	processNum   int64
	gateaddr     string
	client       *rpc.Client
	maxTimeout   int64
}

func NewUfopdConn(gate string, timeout int64) *UfopdConn {
	if timeout == 0 {
		timeout = default_max_timeout
	}
	conn := &UfopdConn{client: &rpc.Client{http.DefaultClient}, gateaddr: gate, maxTimeout: timeout}
	return conn
}

type UfopdOpRet struct {
	r     *http.Response
	retry bool
	err   error
}

func (conn *UfopdConn) Op(xl *xlog.Logger, f io.Reader, fsize int64, fopCtx *fop.FopCtx, c *http.Client, uappInfo *qufop.UappInfo) (r *http.Response, retry bool, err error) {

	atomic.AddInt64(&conn.processNum, 1)
	defer atomic.AddInt64(&conn.processNum, -1)

	var client *rpc.Client
	if c != nil {
		client = &rpc.Client{c}
	} else {
		client = &rpc.Client{http.DefaultClient}
	}

	query := ""
	//for qiniu style op or old fop which transplanted into ufop
	if uappInfo.Type == qufop.QINIU_UAPP {
		query += "op?"
		v := gourl.Values{}
		v.Set("fsize", strconv.FormatInt(fsize, 10))
		if fopCtx != nil {
			v.Set("cmd", fopCtx.RawQuery)
			v.Set("sp", fopCtx.StyleParam)
			v.Set("url", fopCtx.URL)
			v.Set("token", fopCtx.Token)
			if fopCtx.Mode != 0 {
				v.Set("mode", strconv.FormatUint(uint64(fopCtx.Mode), 10))
				v.Set("uid", strconv.FormatUint(uint64(fopCtx.Uid), 10))
				v.Set("bucket", fopCtx.Bucket)
				v.Set("key", fopCtx.Key)
				v.Set("fh", base64.URLEncoding.EncodeToString(fopCtx.Fh))
			}
		}
		query += v.Encode()
	} else {
		//normal ufop
		pos := strings.Index(fopCtx.RawQuery, "/")
		if pos == -1 {
			query = ""
		} else {
			query = fopCtx.RawQuery[pos+1:]
		}
	}

	url := conn.gateaddr + "/" + query
	req, err := http.NewRequest("POST", url, f)
	if err != nil {
		return
	}

	//gate需要用Host判断请求的uapp名称
	if len(uappInfo.Domains) > 0 {
		req.Host = uappInfo.Domains[0]
	} else {
		xl.Error("ufopd.Op: no domain found", uappInfo.UappName)
		err = errors.New("no domain found")
		return
	}
	//CL=-1 makes go transfer.go/WriteBody call WriteTo, not Read
	req.ContentLength = -1

	retCh := make(chan *UfopdOpRet)
	var resp *http.Response
	go func() {
		//可能因为超时而关闭了channel
		defer func() {
			if err := recover(); err != nil {
				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
		resp, err = client.Do(xl, req)
		if err != nil {
			conn.lastFailTime = time.Now().Unix()
			retCh <- &UfopdOpRet{r: nil, retry: true, err: err}
			return
		}

		if resp.StatusCode/100 != 2 {
			err = rpc.ResponseError(resp)
		}
		retCh <- &UfopdOpRet{r: resp, retry: false, err: err}
	}()

	//FIXME::long operation timeout
	//ufop的种类各不相同，没法设定各个ufop操作的超时，如果某个ufop需要的时间很久，应该
	//设计为异步返回，这样fopg就不会阻塞，但是这种异步请求不能管道化
	select {
	case <-time.After(time.Second * time.Duration(conn.maxTimeout)):
		close(retCh)
		return nil, false, errors.New("ufop request timeout")
	case ret := <-retCh:
		return ret.r, ret.retry, ret.err
	}
	return
}

func (conn *UfopdConn) LastFailTime() int64 {
	return atomic.LoadInt64(&conn.lastFailTime)
}

func (conn *UfopdConn) ProcessingNum() int64 {
	return atomic.LoadInt64(&conn.processNum)
}
