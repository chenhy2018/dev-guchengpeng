package autoua

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/qiniu/version"

	rpc1 "github.com/qiniu/rpc.v1"
	rpc2 "github.com/qiniu/rpc.v2"
	rpc3 "github.com/qiniu/rpc.v3"
	rpc7 "qiniupkg.com/x/rpc.v7"
)

func program() string {
	return path.Base(os.Args[0])
}

func pid() string {
	return fmt.Sprint(os.Getpid())
}

func goversion() string {
	return runtime.Version()
}

func goos() string {
	return runtime.GOOS
}

func goarch() string {
	return runtime.GOARCH
}

func hostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func programVersion() string {
	sp := strings.Fields(strings.TrimSpace(version.Version()))
	if len(sp) == 0 || sp[0] == "develop" {
		data, err := ioutil.ReadFile(os.Args[0])
		if err != nil {
			return "_"
		}
		return fmt.Sprintf("%x", md5.Sum(data))[:10]
	}
	if len(sp) > 10 {
		return sp[0][:10]
	}
	return sp[0]
}

func init() {
	ua := fmt.Sprintf("%s/%s (%s/%s; %s) %s/%s",
		program(),
		programVersion(),
		goos(),
		goarch(),
		goversion(),
		hostname(),
		pid(),
	)
	setUA := func() {
		if rpc1.UserAgent != ua {
			rpc1.UserAgent = ua
		}
		if rpc2.UserAgent != ua {
			rpc2.UserAgent = ua
		}
		if rpc3.UserAgent != ua {
			rpc3.UserAgent = ua
		}
		if rpc7.UserAgent != ua {
			rpc7.UserAgent = ua
		}
	}
	setUA()
	time.AfterFunc(time.Second, setUA)
}
