package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/qconf/qconfapi"
)

var id = flag.String("i", "", "qconf id")
var cmd = flag.String("o", "set", "command to perform, add|get|set|del")
var cf = flag.String("f", "qboxconfcmd.conf", "configuration of qconf client")

func main() {

	flag.Parse()

	xl := xlog.NewDummy()
	if *id == "" {
		xl.Error("empty qconf id")
		return
	}

	b, err := ioutil.ReadFile(*cf)
	if err != nil {
		xl.Error("read config failed =>", err)
		return
	}

	var cfg qconfapi.Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		xl.Error("unmarshal config failed =>", err)
		return
	}

	qconf := qconfapi.New(&cfg)
	switch *cmd {
	case "add":
		err = qconf.Insert(xl, map[string]string{"_id": *id})
	case "get":
		var data interface{}
		err = qconf.Get(xl, &data, *id, qconfapi.Cache_Normal)
		if err != nil {
			break
		}
		b, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			xl.Error("get unexpected data =>", data)
			return
		}
		fmt.Println(string(b))
	case "set":
		var _data interface{}
		err = qconf.Get(xl, &_data, *id, qconfapi.Cache_Normal)
		if err != nil {
			break
		}
		b, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			xl.Error("read input failed =>", err)
			return
		}
		var data interface{}
		err = json.Unmarshal(b, &data)
		if err != nil {
			xl.Error("unmarshal input failed =>", err)
			return
		}
		err = qconf.Modify(xl, *id, data)
	case "del":
		err = qconf.Delete(xl, *id)
	default:
		xl.Errorf("unknown command =>", *cmd)
		return
	}
	if err != nil {
		if err1, ok := err.(*rpc.ErrorInfo); ok {
			xl.Error("op failed =>", err1.ErrorDetail())
			return
		}
		xl.Error("op failed =>", err)
		return
	}
	return
}
