// +build go1.5

package fopg

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"code.google.com/p/go.net/context"

	"sync"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api/dc"
	"qbox.us/cc/config"
	"qbox.us/dcutil"
	"qbox.us/fop"
)

var (
	NotCacheCmds = []string{"imageInfo", "exif", "imageAve", "qrcode", "barcode", "scanbar", "saveas", "pm3u8", "ufop", "avconvert", "vinfo"}
)

type Config struct {
	LbConfig         lb.Config          `json:"lb_conf"`
	LbTrConfig       lb.TransportConfig `json:"lb_tr_conf"`
	FailoverConfig   lb.Config          `json:"failover_conf"`
	FailoverTrConfig lb.TransportConfig `json:"failover_tr_conf"`
	AccessKey        string             `json:"access_key"`
	SecretKey        string             `json:"secret_key"`
	DcCache          bool               `json:"dc_cache"`
	FopResConfig     FailoverConfig     `json:"fop_res_config"` //  一般可不配置，采用默认配置

	DCConf dc.Config `json:"dc_conf"`

	Rcfg *config.ReloadingConfig `json:"reload_conf"`
}

type Client struct {
	mutex    sync.RWMutex
	client   *lb.Client
	mac      *digest.Mac
	cacheExt *dcutil.CacheExt
	dcCache  bool
	close    func() error

	fopResClient *FailoverClient
}

func NewClient(cfg *Config) (*Client, error) {

	cfg.LbConfig.ShouldRetry = shouldRetry
	cfg.FailoverConfig.ShouldRetry = shouldRetry
	mac := &digest.Mac{
		AccessKey: cfg.AccessKey,
		SecretKey: []byte(cfg.SecretKey),
	}
	t1 := lb.NewTransport(&cfg.LbTrConfig)
	clientTr := digest.NewTransport(mac, t1)
	t2 := lb.NewTransport(&cfg.FailoverTrConfig)
	failoverTr := digest.NewTransport(mac, t2)

	var (
		client *Client = &Client{mac: mac, dcCache: cfg.DcCache, close: newclose(nil, nil, nil)}
		err    error
	)
	if cfg.Rcfg != nil && len(cfg.Rcfg.ConfName) > 0 {
		err = config.StartReloading(cfg.Rcfg, client.onReload)
	} else {
		client.client = lb.NewWithFailover(&cfg.LbConfig, &cfg.FailoverConfig, clientTr, failoverTr, nil)
		if cfg.DcCache {
			dcTransport := dc.NewTransport(cfg.DCConf.DialTimeoutMS, cfg.DCConf.RespTimeoutMS, cfg.DCConf.TransportPoolSize)
			cache := dc.NewClient(cfg.DCConf.Servers, cfg.DCConf.TryTimes, dcTransport)
			dcExt := dcutil.NewExt(dc.NewDiskCacheExt(cache))
			client.cacheExt = &dcExt
			client.close = newclose(cache, t1, t2)
		} else {
			client.close = newclose(nil, t1, t2)
		}
		client.fopResClient = NewFailoverClient(&cfg.FopResConfig, clientTr, failoverTr, nil)
	}
	if err != nil {
		return nil, err
	}
	return client, nil
}

func newclose(cache *dc.Client, tr1, tr2 http.RoundTripper) func() error {
	return func() error {
		if cache != nil {
			err := cache.Close()
			if err != nil {
				return err
			}
		}
		if tr1 != nil {
			if c, ok := tr1.(io.Closer); ok {
				err := c.Close()
				if err != nil {
					return err
				}
			}
		}
		if tr2 != nil {
			if c, ok := tr2.(io.Closer); ok {
				err := c.Close()
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func newRequestWithCtx(ctx context.Context, method, urlStr string, body io.ReaderAt) (req *lb.Request, err error) {
	req, err = lb.NewRequest(method, urlStr, body)
	if err != nil {
		return
	}
	req = req.WithContext(ctx)
	return
}

//fopg resp header中新增加X-Resp-Code字段；dora只保证非2xx错误一定会有此字段
func (c *Client) Op(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (
	resp *http.Response, err error) {

	u := "/op?" + EncodeQuery(fh, fsize, ctx)
	xl := xlog.FromContextSafe(tpCtx)
	xl.Debug("fopg.Op:", u)
	var key []byte
	var cacheErr error
	var needCache bool
	if c.dcCache && !hasPrefixArr(ctx.RawQuery, NotCacheCmds) {
		key = generateCacheKey(fh, fsize, ctx.RawQuery, ctx.StyleParam)
		resp, cacheErr = c.loadFromCache(xl, key)
		needCache = true
		if resp != nil {
			xl.Debugf("fopg client dcCache resp header: %v", resp.Header)
		}
	}
	if !needCache || cacheErr != nil {
		var req *lb.Request
		req, err = newRequestWithCtx(tpCtx, "POST", u, nil)
		if err != nil {
			return
		}
		resp, err = c.loadClient().DoWithCtx(req)
		if err != nil {
			if tpCtx.Err() != nil {
				err = tpCtx.Err()
			}
			return
		}
		if resp != nil {
			xl.Debugf("fopg client resp header: %v", resp.Header)
		}
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		if resp != nil {
			xl.Debugf("fopg err resp header: %v", resp.Header)
		}
		return
	}
	if needCache && cacheErr != nil {
		resp, err = c.saveToCache(tpCtx, key, resp)
		if err != nil {
			return
		}
		if resp != nil {
			xl.Debugf("fopg saveToCache outresp header: %v", resp.Header)
		}
	}
	return
}

func NeedCache(resp *http.Response) bool {
	return resp.Header.Get("X-Cache-Control") != "no-cache"
}

func NeedCdnCache(resp *http.Response) bool {
	return resp.Header.Get("X-Cdn-Cache-Control") != "no-cache"
}

func (c *Client) Nop(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (
	resp *http.Response, err error) {

	u := "/nop?" + EncodeQuery(fh, fsize, ctx)
	xl := xlog.FromContextSafe(tpCtx)
	xl.Debug("fopg.Nop:", u)
	var req *lb.Request
	req, err = newRequestWithCtx(tpCtx, "POST", u, nil)
	if err != nil {
		return
	}
	resp, err = c.loadClient().DoWithCtx(req)
	if err != nil {
		if tpCtx.Err() != nil {
			err = tpCtx.Err()
		}
		return
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}

	return
}

func (c *Client) RangeOp(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx, rangeHeader string) (
	resp *http.Response, err error) {

	u := "/rop?" + EncodeQuery(fh, fsize, ctx)
	xl := xlog.FromContextSafe(tpCtx)
	xl.Debug("fopg.RangeOp:", u)

	req, err := newRequestWithCtx(tpCtx, "GET", u, nil)
	if err != nil {
		return
	}
	req.Header.Set("Range", rangeHeader)

	resp, err = c.loadClient().DoWithCtx(req)
	if err != nil {
		if tpCtx.Err() != nil {
			err = tpCtx.Err()
		}
		return
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
	}
	return
}

func (c *Client) PersistentOp(tpCtx context.Context, ret interface{}, fh []byte, fsize int64, ctx *fop.FopCtx) (err error) {
	u := "/pop?" + EncodeQuery(fh, fsize, ctx)
	xl := xlog.FromContextSafe(tpCtx)
	xl.Debug("fopg.PersistentOp:", u)

	var req *lb.Request
	req, err = newRequestWithCtx(tpCtx, "POST", u, nil)
	if err != nil {
		return
	}
	resp, err := c.loadClient().DoWithCtx(req)
	if err != nil {
		if tpCtx.Err() != nil {
			err = tpCtx.Err()
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}
	return json.NewDecoder(resp.Body).Decode(ret)
}

func (c *Client) PersistentOpEx(tpCtx context.Context, ret interface{}, fh []byte, fsize int64, ctx *fop.FopCtx) (_ *http.Response, err error) {
	u := "/pop?" + EncodeQuery(fh, fsize, ctx)
	xl := xlog.FromContextSafe(tpCtx)
	xl.Debug("fopg.PersistentOpEx:", u)

	var req *lb.Request
	req, err = newRequestWithCtx(tpCtx, "POST", u, nil)
	if err != nil {
		return
	}
	resp, err := c.loadClient().DoWithCtx(req)
	if err != nil {
		if tpCtx.Err() != nil {
			err = tpCtx.Err()
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return resp, err
	}
	return resp, json.NewDecoder(resp.Body).Decode(ret)
}

func (c *Client) List(tpCtx context.Context) (fops []string, err error) {
	err = c.loadClient().CallWithCtx(tpCtx, &fops, "/list")
	return
}

func (c *Client) ExtractResponse(tpCtx context.Context, resp *http.Response) (r io.ReadCloser, length int64, mime string, needRetry, noCdnCache bool, err error) {
	xl := xlog.FromContextSafe(tpCtx)

	outType := fop.GetOutType(resp)
	if resp.Header.Get("X-Cdn-Cache-Control") == "no-cache" {
		noCdnCache = true
	}
	switch outType {
	case fop.OutTypeDC:
		var out fop.DCOut
		err = fop.DecodeAddrOut(&out, resp)
		resp.Body.Close()
		if err != nil {
			err = errors.Info(err, "ExtractResponse: DecodeAddrOut DCOut")
			return
		}
		xl.Debugf("ExtractResponse: dcAddrOut, dcHost:%s, dcKey:%s", out.Host, base64.URLEncoding.EncodeToString(out.Key))
		var metas map[string]string

		urlDC := out.Host + "/get/" + base64.URLEncoding.EncodeToString(out.Key)

		reqObj, err1 := http.NewRequest("POST", urlDC, nil)
		if err1 != nil {
			err = errors.Info(err1, "ExtractResponse: failed to construct dc request")
			return
		}
		req := &FailoverRequest{
			req: reqObj,
			ctx: tpCtx,
		}
		resp2, err2 := c.loadFopResClient().DoCtx(req)
		if err2 != nil {
			err = errors.Info(err2, "ExtractResponse: dcutil.GetWithHost")
			return
		}
		r, _, length, metas, err = decodeMeta(xl, resp2.Body, int(resp2.ContentLength))
		mime = metas["mime"]
	case fop.OutTypeTmp:
		var out fop.TmpOut
		err = fop.DecodeAddrOut(&out, resp)
		resp.Body.Close()
		if err != nil {
			err = errors.Info(err, "ExtractResponse: DecodeAddrOut TmpOut")
			return
		}
		xl.Debugf("ExtractResponse: tmpAddrOut, tmpURL:%s, mimeType:%s", out.URL, out.MimeType)
		mime = out.MimeType

		reqObj, err1 := http.NewRequest("POST", out.URL, nil)
		if err1 != nil {
			err = errors.Info(err1, "ExtractResponse: failed to construct tmp request")
			return
		}
		req := &FailoverRequest{
			req: reqObj,
			ctx: tpCtx,
		}
		resp2, err2 := c.loadFopResClient().DoCtx(req)
		if err2 != nil {
			err = errors.Info(err2, "ExtractResponse: rpcClient.Get tmpOut.URL")
			return
		}
		if resp2.StatusCode != http.StatusOK {
			err = rpc.ResponseError(resp2)
			resp2.Body.Close()
			return
		}
		r, length = resp2.Body, resp2.ContentLength
	default:
		xl.Debug("ExtractResponse: streamOut")
		r, length, mime = resp.Body, resp.ContentLength, resp.Header.Get("Content-Type")
	}
	return
}
