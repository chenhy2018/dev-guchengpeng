package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/qiniu/http/restrpc.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/log.v1"
	"qiniupkg.com/mockable.v1"
	mockhttp "qiniupkg.com/mockable.v1/net/http"

	"qbox.us/cc/config"
)

type Config struct {
	BindHost string `json:"bind_host"`
}

func main() {

	config.Init("f", "demo", "demoserver.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load", err)
	}
	log.SetOutputLevel(0)

	mockable.Init()

	router := restrpc.Router{}
	mux := router.Register(&Service{})
	log.Infof("serve @%v", conf.BindHost)
	err := mockhttp.ListenAndServe(conf.BindHost, mux)
	log.Fatal("ListenAndServe", err)
}

type Service struct {
}

type sizeArg struct {
	Size int64 `json:"size"`
}

func (p *Service) GetSize(args *sizeArg, env *rpcutil.Env) {

	b := make([]byte, args.Size)
	env.W.Write(b)
}

type getArg struct {
	Url string `json:"url"`
}

type getRet struct {
	TimeMs int64 `json:"timeMs"`
	Size   int64 `json:"size"`
}

func (p *Service) PostGet(args *getArg, env *rpcutil.Env) (ret getRet, err error) {

	start := time.Now()
	client := http.Client{Transport: mockhttp.DefaultTransport}
	resp, err := client.Get(args.Url)
	if err != nil {
		return
	}
	n, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return
	}
	duration := time.Since(start)
	ret.Size = n
	ret.TimeMs = int64(duration / time.Millisecond)
	return
}
