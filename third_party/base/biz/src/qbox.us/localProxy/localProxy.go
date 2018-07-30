package localProxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/qiniu/log.v1"
)

//Some sdk cannot change request host different with request url (like prometheus),
//localProxy can do this.

type Config struct {
	//real address
	Addr string
	//proxy address
	ProxyAddr string
	//local listen addr, forward request to make host(Addr) different with req url(ProxyAddr)
	//sdk should send request to this address
	LocalAddr string
}

func Listen(config Config) error {
	if config.Addr == "" {
		return errors.New("addr cannot be empty")
	}
	if config.ProxyAddr == "" {
		return errors.New("proxyAddr cannot be empty")
	}
	if config.LocalAddr == "" {
		return errors.New("localAddr cannot be empty")
	}

	reverseProxy, err := newSingleHostProxy(config.Addr, config.ProxyAddr)
	if err != nil {
		return err
	}
	go func() {
		log.Fatalln(http.ListenAndServe(config.LocalAddr, reverseProxy))
	}()
	return nil
}

func newSingleHostProxy(host, proxyHost string) (*httputil.ReverseProxy, error) {

	if !strings.HasPrefix(proxyHost, "http://") {
		proxyHost = "http://" + proxyHost
	}
	proxyURL, err := url.Parse(proxyHost)
	if err != nil {
		return nil, err
	}
	p := http.ProxyURL(proxyURL)
	tp := &http.Transport{
		Proxy: p,
	}

	if !strings.HasPrefix(host, "http://") {
		host = "http://" + host
	}
	target, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	targetQuery := target.RawQuery

	director := func(req *http.Request) {
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	reverseProxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: tp,
	}
	return reverseProxy, nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
