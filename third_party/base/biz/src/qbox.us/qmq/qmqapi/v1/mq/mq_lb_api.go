package mq

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	lbv3 "github.com/qiniu/rpc.v1/lb.v3"
	"golang.org/x/net/context"
	reqid "qiniupkg.com/x/reqid.v7"
)

type LBClient struct {
	clientW *lbv3.Client
	clientR *lb.Client
	clients map[string]Service
}

func NewTimeoutLBClient(hosts []string, mac *digest.Mac, dial, resp, retry time.Duration) (*LBClient, error) {
	timeoutTransport := &http.Transport{Proxy: http.ProxyFromEnvironment} // DefaultTransport
	timeoutTransport.Dial = func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, dial)
	}
	timeoutTransport.ResponseHeaderTimeout = resp
	t := digest.NewTransport(mac, timeoutTransport)
	clients := make(map[string]Service)
	for _, host := range hosts {
		clients[host] = NewWithHost(t, host)
	}
	lbCfg := &lb.Config{
		Hosts:              hosts,
		TryTimes:           uint32(len(hosts)),
		FailRetryIntervalS: 2, // 2s
	}
	clientR := lb.New(lbCfg, t)
	clientW := lbv3.New(&lbv3.Config{
		Http:           &http.Client{Transport: t},
		Hosts:          hosts,
		HostRetrys:     len(hosts),
		RetryTimeoutMs: int(retry / time.Millisecond),
	})

	c := &LBClient{
		clientW: clientW,
		clientR: clientR,
		clients: clients,
	}
	return c, nil
}

func (r *LBClient) Make(l rpc.Logger, mqId string, expires int) (err error) {

	url := "/make/" + mqId + "/expires/" + strconv.Itoa(expires)
	return r.clientR.Call(l, nil, url)
}

func (r *LBClient) Put(l rpc.Logger, mqId string, msg []byte) (msgId string, err error) {

	return r.PutEx(l, mqId, bytes.NewReader(msg), len(msg))
}

func (r *LBClient) PutString(l rpc.Logger, mqId string, msg string) (msgId string, err error) {

	return r.PutEx(l, mqId, strings.NewReader(msg), len(msg))
}

func (r *LBClient) PutEx(l rpc.Logger, mqId string, msgr io.ReaderAt, bytes int) (msgId string, err error) {

	ctx := reqid.NewContext(context.TODO(), l.ReqId())
	resp, err := r.clientW.PostWith(ctx, "/put/"+mqId, "application/octet-stream", msgr, bytes)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}
	msgId = resp.Header.Get("X-Id")
	return
}

func (r *LBClient) Get(l rpc.Logger, mqId string) (host string, msg []byte, msgId string, err error) {

	host, resp, err := r.clientR.PostWithHostRet(l, "/get/"+mqId, "application/x-www-form-urlencoded", nil, 0)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	msgId = resp.Header.Get("X-Id")

	length := resp.ContentLength
	if length <= 0 || length > (1<<22) {
		err = errors.New("invalid ContentLength")
		return
	}
	msg = make([]byte, length)
	_, err = io.ReadFull(resp.Body, msg)
	return
}

func (r *LBClient) Delete(l rpc.Logger, host string, mqId string, msgId string) (err error) {
	client := r.clients[host]
	req, err := http.NewRequest("POST", client.Host+"/delete/"+mqId, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-Id", msgId)
	resp, err := client.Conn.Do(l, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
	}
	return
}
