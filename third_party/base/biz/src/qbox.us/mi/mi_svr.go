package mi

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"qbox.us/api"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"strings"
	"syscall"
)

type InstConfig struct {
	IOHosts []string `json:"ios"`     // ["https://io1.qbox.us:80", "https://io2.qbox.us:80", ...]
	FSHost  string   `json:"fs"`      // "https://fs.qbox.us:80"
	ESHost  string   `json:"es"`      // "https://es.qbox.us:80"
	RSHost  string   `json:"rs"`      // "https://rs.qbox.us:80"
	IOHost  string   `json:"io"`      // "https://io.qbox.us:80"
	AccHost string   `json:"account"` // "https://account.qbox.us:80"
}

type Config struct {
	MaxProcs int                    `json:"max_procs"`
	BindHost string                 `json:"bind_host"`
	Settings map[string]*InstConfig `json:"settings"`
	Account  account.Interface
}

func (cfg *Config) LoadFromString(data string) (err error) {

	err = json.Unmarshal([]byte(data), &cfg.Settings)
	return
}

func (cfg *Config) Load(file string) (err error) {

	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&cfg.Settings)
	return
}

type MultiIS struct {
	Config
}

func New(cfg *Config) (p *MultiIS, err error) {
	p = &MultiIS{*cfg}
	return
}

func settings(req *http.Request) string {

	req.ParseForm()

	log.Println(req.URL.Path, "-", req.Form)

	if agent, ok := req.Form["agent"]; ok { // qbox-ios-0.5.11-dev
		parts := strings.Split(agent[0], "-")
		if len(parts) >= 4 {
			return parts[3]
		} else {
			return "pub"
		}
	}
	return ""
}

func (p *MultiIS) iotable(w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}
	_, err := account.GetAuth(p.Account, req)
	if err != nil {
		w.ReplyWithCode(api.BadToken)
		return
	}

	sett := settings(req)
	if cfg, ok := p.Settings[sett]; ok {
		w.ReplyWith(200, cfg.IOHosts)
	} else {
		w.ReplyWithCode(api.InvalidArgs)
	}
}

type hostsRet struct {
	FS string `json:"fs"`
	ES string `json:"es"`
	RS string `json:"rs"`
	IO string `json:"io"`
}

func (p *MultiIS) hosts(w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}
	_, err := account.GetAuth(p.Account, req)
	if err != nil {
		w.ReplyWithCode(api.BadToken)
		return
	}

	sett := settings(req)
	if cfg, ok := p.Settings[sett]; ok {
		w.ReplyWith(200, hostsRet{cfg.FSHost, cfg.ESHost, cfg.RSHost, cfg.IOHost})
	} else {
		w.ReplyWithCode(api.InvalidArgs)
	}
}

type initRet struct {
	Host string `json:"account"`
}

func (p *MultiIS) init(w1 http.ResponseWriter, req *http.Request) {

	w := rpc.ResponseWriter{w1}

	sett := settings(req)
	if cfg, ok := p.Settings[sett]; ok {
		w.ReplyWith(200, initRet{cfg.AccHost})
	} else {
		w.ReplyWithCode(api.InvalidArgs)
	}
}

func RegisterHandlers(mux *http.ServeMux, cfg *Config) error {

	if cfg.Account == nil {
		log.Println("cfg.Account == nil")
		return syscall.EINVAL
	}
	p, err := New(cfg)
	if err != nil {
		log.Println(err)
		return err
	}
	mux.HandleFunc("/iohosts", func(w http.ResponseWriter, req *http.Request) { p.iotable(w, req) })
	mux.HandleFunc("/hosts", func(w http.ResponseWriter, req *http.Request) { p.hosts(w, req) })
	mux.HandleFunc("/connect", func(w http.ResponseWriter, req *http.Request) { p.init(w, req) })
	return nil
}

func Run(addr string, cfg *Config) error {

	mux := http.DefaultServeMux
	err := RegisterHandlers(mux, cfg)
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, mux)
}
