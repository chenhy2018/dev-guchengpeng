package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"qiniupkg.com/mockable.v1"
	mockhttp "qiniupkg.com/mockable.v1/net/http"
)

func main() {

	exec.Command("go", "install", "-v", "./...").Run()

	path := os.Getenv("QBOXROOT") + "/pili-zeus/bin"
	cfgStr := fmt.Sprintf(cfgFmtStr, path, path)
	cluster := mockable.RunCluster(cfgStr)
	defer cluster.Close()

	time.Sleep(7 * time.Second)

	client := rpc.Client{&http.Client{Transport: mockhttp.DefaultTransport}}

	type Ret struct {
		TimeMs int64 `json:"timeMs"`
		Size   int64 `json:"size"`
	}

	var ret Ret
	param := map[string][]string{
		"url": []string{"http://0.0.0.202:10/size?size=5000"},
	}
	err := client.CallWithForm(xlog.NewDummy(), &ret, "http://0.0.0.200:10/get", param)
	if err != nil {
		log.Error("err:", err)
		return
	}
	fmt.Printf("ret: %+v\n", ret)

	param = map[string][]string{
		"url": []string{"http://0.0.0.203:10/size?size=5000"},
	}
	err = client.CallWithForm(xlog.NewDummy(), &ret, "http://0.0.0.201:10/get", param)
	if err != nil {
		log.Error("err:", err)
		return
	}
	fmt.Printf("ret: %+v\n", ret)
}

func errNil(err error) {
	if err != nil {
		log.Fatal("err should be nil:", err)
	}
}

var cfgFmtStr string = `
{
	"idcs": [
		{
			"name": "wan",
			"nodes": [
				{
					"name": "",
					"ips": {
						"tel": "0.0.0.100",
						"uni": "0.0.0.101"
					},
					"defaultIsp": "tel"
				}
			]
		},
		{
			"name": "idc1",
			"nodes": [
				{
					"name": "node1",
					"ips": {
						"tel": "0.0.0.200",
						"bgp": "0.0.0.201"
					},
					"defaultIsp": "tel",
					"procs": [
						{
							"name": "demo1",
							"workdir": "./",
							"exec": ["%s/demoserver", "-f", "demoserver1.conf"]
						}
					]
				}
			]
		},
		{
			"name": "idc2",
			"nodes": [
				{
					"name": "node2",
					"ips": {
						"tel": "0.0.0.202",
						"bgp": "0.0.0.203"
					},
					"defaultIsp": "tel",
					"procs": [
						{
							"name": "demo2",
							"workdir": "./",
							"exec": ["%s/demoserver", "-f", "demoserver2.conf"]
						}
					]
				}
			]
		}
	],
	"speeds": [
		{
			"from": "idc1:tel",
			"to": "idc2:tel",
			"speed": [{"Bps": 2000}]
		},
		{
			"from": "idc1:bgp",
			"to": "idc2:bgp",
			"speed": [{"Bps": 4000, "duration": 2000}, {"Bps": 3000}]
		}
	],
	"defaultSpeed": [{"Bps": 10000}]
}
`
