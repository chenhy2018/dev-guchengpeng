package collector

import (
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type Client struct {
	urls []string
}

func NewClient(urls []string) Client {
	if urls == nil || len(urls) == 0 {
		urls = []string{
			"127.0.0.1:3444",
		}
	}
	return Client{
		urls: urls,
	}
}

func (r Client) Post(body io.Reader, n int) error {
	url := r.urls[rand.Intn(len(r.urls))]
	resp, err := DefaultClient.Post("http://"+url+"/v1/set", "text/plain", body)
	if err != nil {
		// log.Printf("[ERROR] -probe- client post failed: %s\n", err)
		return err
	}
	defer resp.Body.Close()
	// TODO !200, timeout   retry
	return nil
}

////////////////////////////////////////////////////////////////////////////////

var DefaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   time.Millisecond * 100, // 连不上就算了
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 500 * time.Millisecond,
}

var DefaultClient *http.Client = &http.Client{
	Transport: DefaultTransport,
	Timeout:   time.Second,
}
