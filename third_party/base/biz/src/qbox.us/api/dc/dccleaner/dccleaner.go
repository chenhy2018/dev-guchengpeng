package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"qbox.us/admin_api/rs"
	"qbox.us/api/dc"
	"qbox.us/api/one/domain"
	"qbox.us/cc/config"
	"qbox.us/fop"
	"qbox.us/qdiscover/discover"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	. "github.com/qiniu/api/conf"
)

// ----------------------------------------------------------
type Task struct {
	Uid    uint32
	Bucket string
	Key    string
	Cmd    string
}

func loadTaskFromUrl(c *domain.Client, rawUrl string) (task Task, err error) {
	urlRet, err := url.Parse(rawUrl)
	if err != nil {
		return
	}

	ret, ok := GetByDomainRets[urlRet.Host]
	if !ok {
		xl := xlog.NewDummy()
		ret, err = c.GetByDomain(xl, urlRet.Host)
		if err != nil {
			err = fmt.Errorf("reqid: %v, GetByomain failed, %v", xl.ReqId(), err)
			return
		}
		GetByDomainRets[urlRet.Host] = ret
	}

	task = Task{
		Uid:    ret.Uid,
		Bucket: ret.Tbl,
		Key:    urlRet.Path[1:],
		Cmd:    urlRet.RawQuery,
	}
	return
}

func loadTaskFromFields(fields []string) (task Task, err error) {
	if len(fields) != 4 {
		err = errors.New("len(fields) != 4")
		return
	}
	uid, err := strconv.Atoi(fields[0])
	if err != nil {
		err = fmt.Errorf("invalid uid: %#v", fields[0])
		return
	}
	task = Task{
		Uid:    uint32(uid),
		Bucket: fields[1],
		Key:    fields[2],
		Cmd:    fields[3],
	}
	return
}

func LoadTasks(domainClient *domain.Client, r io.Reader) (tasks []Task) {
	var task Task
	var err error
	scanner := bufio.NewScanner(r)
	//以'\n'为结束符读入一行
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 4 {
			task, err = loadTaskFromFields(fields)
			if err != nil {
				fmt.Fprintf(os.Stderr, "fields: %#v, skipped, loadTaskFromFields failed, %v\n", fields, err)
				continue
			}
		} else if len(fields) == 1 {
			task, err = loadTaskFromUrl(domainClient, line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "url: %#v, skipped, loadTaskFromUrl failed, %v\n", line, err)
				continue
			}
		} else {
			fmt.Fprintf(os.Stderr, "%#v, wrong format, skipped, should be <full url> or <uid bucket key cmd>\n", line)
			continue
		}
		tasks = append(tasks, task)
	}
	return
}

// ----------------------------------------------------------
type RSClient struct {
	Host string
	Conn rpc.Client
}

func NewRSClient(host string, t http.RoundTripper) *RSClient {
	client := &http.Client{Transport: t}
	return &RSClient{Host: host, Conn: rpc.Client{client}}
}

func (rs RSClient) EntryInfo(xl rpc.Logger, uid uint32, bucket, key string) (info rs.EntryInfoRet, err error) {
	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"bucket": {bucket},
		"key":    {key},
	}
	err = rs.Conn.CallWithForm(xl, &info, rs.Host+"/entryinfo", params)
	return
}

// -----------------------------------------------------------
func canonicalHTTPAddr(addr string) string {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	return addr
}

// -----------------------------------------------------------
type Config struct {
	DebugLevel  int             `json:"debug_level"`
	RsHost      string          `json:"rs_host"`
	OneHost     string          `json:"one_host"`
	DCConf      dc.Config       `json:"dc"`
	DiscoverDC  discover.Config `json:"discover_dc"`
	AdminAccess string          `json:"admin_access_key"`
	AdminSecret string          `json:"admin_secret_key"`
}

var GetByDomainRets map[string]domain.GetByDomainRet

type DCAttrs struct {
	Processing int64    `bson:"processing"`
	Keys       []string `bson:"keys"`
}

func main() {
	configFile := flag.String("f", "dccleaner.conf", "dcCleaner conf file")
	flag.Parse()

	var conf Config
	if err := config.LoadFile(&conf, *configFile); err != nil {
		fmt.Fprintf(os.Stderr, "config.Load failed: %v\n", err)
		os.Exit(1)
	}
	log.SetOutputLevel(conf.DebugLevel)

	if conf.RsHost == "" {
		conf.RsHost = RS_HOST
	}
	GetByDomainRets = make(map[string]domain.GetByDomainRet)

	adminTransport := digest.NewTransport(&digest.Mac{conf.AdminAccess, []byte(conf.AdminSecret)}, rpc.DefaultTransport)
	rsClient := NewRSClient(conf.RsHost, adminTransport)
	domainClient := domain.New(conf.OneHost, adminTransport)
	tasks := LoadTasks(&domainClient, os.Stdin)

	// dc
	if len(conf.DiscoverDC.DiscoverHosts) != 0 {
		svrmgr, err := discover.NewServiceManager(&conf.DiscoverDC)
		if err != nil {
			log.Fatal("NewServiceManager for dc failed:", err)
		}
		discoverdDCs := svrmgr.Services()
		log.Info("loadDC - discoverd num:", len(discoverdDCs))
		for i, info := range discoverdDCs {
			var attrs DCAttrs
			if err := info.Attrs.ToStruct(&attrs); err != nil {
				log.Warn("loadDC - Attrs.ToStruct failed:", err, i+1, info.Addr)
				continue
			}
			svr := dc.DCConn{
				Keys: attrs.Keys,
				Host: canonicalHTTPAddr(info.Addr),
			}
			log.Infof("loadDC - discover %d\t%v", i+1, svr)
			conf.DCConf.Servers = append(conf.DCConf.Servers, svr)
		}
	}
	// init dc
	if len(conf.DCConf.Servers) == 0 {
		fmt.Fprintf(os.Stderr, "empty dc servers\n")
		os.Exit(1)
	}

	dcTransport := dc.NewTransport(conf.DCConf.DialTimeoutMS, conf.DCConf.RespTimeoutMS, conf.DCConf.TransportPoolSize)
	cache := dc.NewClient(conf.DCConf.Servers, conf.DCConf.TryTimes, dcTransport)

	for _, task := range tasks {
		xl := xlog.NewDummy()
		info, err := rsClient.EntryInfo(xl, task.Uid, task.Bucket, task.Key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v, EntryInfo failed, %v\n", task, err)
			continue
		}
		xl.Infof("entryInfo:%v", info)
		fh, err := base64.URLEncoding.DecodeString(info.EncodedFh)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v decode fh failed, %v\n", task, err)
			continue
		}

		cacheKey := fop.CacheKey(fh, info.Fsize, task.Cmd, "")
		_, err = cache.Delete(xl, cacheKey)
		if err != nil {
			if e, ok := err.(*rpc.ErrorInfo); ok && e.Code == 404 {
				fmt.Fprintf(os.Stdout, "%v, %s\n", task, e.Err)
			} else {
				fmt.Fprintf(os.Stderr, "%v, clean failed, %v\n", task, err)
			}
			continue
		}
		fmt.Fprintf(os.Stdout, "%v, clean success\n", task)

	}
	fmt.Println("all task done")
}
