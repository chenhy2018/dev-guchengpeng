// +build ignore

package main

import (
	"time"

	"github.com/qiniu/xlog.v1"

	"qbox.us/qdiscover/discover"
)

func main() {

	xl := xlog.NewDummy()
	cmds := []string{"exif", "imageMogr", "imageInfo", "urlInfo", "imageView", "watermark", "imageMogr2", "imageView2", "mgjThumb"}
	attrs := map[string]interface{}{
		"cmds": cmds,
	}

	client := discover.New([]string{"http://127.0.0.1:18888"}, nil)
	//client := discover.New([]string{"http://192.168.34.29:18888", "http://192.168.34.30:18888"}, nil)
	for {
		err := client.ServiceRegister(xl, "me.cc:41000", "fopagent", attrs)
		//	err := client.ServiceRegister(xl, "b2gkk663.qiniu.qcos.qbox.me:80", "fopagent", attrs)
		if err != nil {
			xl.Error(err)
		}
		time.Sleep(time.Second * 7)
	}
}
