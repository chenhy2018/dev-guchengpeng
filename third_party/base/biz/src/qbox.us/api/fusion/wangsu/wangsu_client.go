package wangsu

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/qiniu/rpc.v3"

	"golang.org/x/net/context"
)

type Mac struct {
	User   string
	APIKey string
}

func NewTransport(mac Mac, transport http.RoundTripper) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Transport{
		Mac:          mac,
		RoundTripper: transport,
	}
}

type Transport struct {
	http.RoundTripper
	Mac
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	timeStr := time.Now().Format(time.RFC1123)
	req.Header.Set("Date", timeStr)
	password := Password(t.APIKey, timeStr)
	req.SetBasicAuth(t.User, password)
	return t.RoundTripper.RoundTrip(req)
}

func Password(apiKey string, rfc1123Time string) string {
	h := hmac.New(sha1.New, []byte(apiKey))
	io.WriteString(h, rfc1123Time)
	digest := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(digest)
}

type Client struct {
	host      string
	rpcClient rpc.Client
}

func New(host string, transport http.RoundTripper) (*Client, error) {
	rpcClient := rpc.Client{&http.Client{Transport: transport}}
	return &Client{host, rpcClient}, nil
}

func NewDefault(host string, mac Mac) (*Client, error) {
	t := rpc.NewTransport(&rpc.TransportConfig{
		DialTimeout: 2 * time.Second,
	})
	t = NewTransport(mac, t)
	return New(host, t)
}

func (c *Client) GetDomains(ctx context.Context, serviceType string, names []string, result interface{}) error {
	u := c.host + "/api/si/domain"
	query := url.Values{}
	if serviceType != "" {
		query.Set("serviceType", serviceType)
	}
	if len(names) > 0 {
		query.Set("domain", strings.Join(names, ";"))
	}
	u += "?" + query.Encode()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "deflate")

	resp, err := c.rpcClient.Do(ctx, req)
	if err != nil {
		return err
	}
	return rpc.CallRet(ctx, result, resp)
}

func (c *Client) GetDomainsDetail(ctx context.Context, names []string, result interface{}) error {
	u := c.host + "/api/si/domainconfigdetail"
	query := url.Values{}
	if len(names) > 0 {
		query.Set("domainList", strings.Join(names, ";"))
	}
	u += "?" + query.Encode()

	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "deflate")

	resp, err := c.rpcClient.Do(ctx, req)
	if err != nil {
		return err
	}
	return rpc.CallRet(ctx, result, resp)
}
