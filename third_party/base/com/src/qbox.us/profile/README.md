# Go profile 工具

### 如何打开profile

在导入库里面增加一行： `_ "qbox.us/profile"`

```go
import (
	_ "qbox.us/profile"
)
```

如果需要展现更多程序内部变量可以publish其他内容，例如publish Goroutine个数(默认已经开启)

```go
import (
	"runtime"

	"qbox.us/profile/expvar"
)

func numGoroutine() interface{} {
	return runtime.NumGoroutine()
}

func init() {
	expvar.Publish("NumGoroutine", expvar.Func(numGoroutine))
}
```

### 如何使用profile

1. 找到profile dump脚本

方法1：profile启动后会将导出命令写入到本地文件，规则是`可执行文件地址+_profile_dump.sh`，例如

```
qboxserver@cs3:~$ ps aux | grep qboxrsf | grep -v grep
qboxser+ 29615  0.0  0.0 511512 16916 ?        Sl   10:24   0:00 /home/qboxserver/rsf/_package/qboxrsf -f qboxrsf.conf
qboxserver@cs3:~$
qboxserver@cs3:~$ cat /home/qboxserver/rsf/_package/qboxrsf_profile_dump.sh
#!/bin/sh
echo `date` dumping runtime status
set -x
curl -sS 'http://127.0.0.1:40843/debug/vars?' -o qboxrsf_15618_20171102_131309_vars
curl -sS 'http://127.0.0.1:40843/debug/pprof/block?seconds=5' -o qboxrsf_15618_20171102_131309_block
curl -sS 'http://127.0.0.1:40843/debug/pprof/goroutine?seconds=5' -o qboxrsf_15618_20171102_131309_goroutine
curl -sS 'http://127.0.0.1:40843/debug/pprof/goroutine?debug=1' -o qboxrsf_15618_20171102_131309_goroutine_debug_1
curl -sS 'http://127.0.0.1:40843/debug/pprof/goroutine?debug=2' -o qboxrsf_15618_20171102_131309_goroutine_debug_2
curl -sS 'http://127.0.0.1:40843/debug/pprof/heap?seconds=5' -o qboxrsf_15618_20171102_131309_heap
curl -sS 'http://127.0.0.1:40843/debug/pprof/mutex?seconds=5' -o qboxrsf_15618_20171102_131309_mutex
curl -sS 'http://127.0.0.1:40843/debug/pprof/threadcreate?seconds=5' -o qboxrsf_15618_20171102_131309_threadcreate
curl -sS 'http://127.0.0.1:40843/debug/pprof/profile?seconds=5' -o qboxrsf_15618_20171102_131309_profile
curl -sS 'http://127.0.0.1:40843/debug/pprof/trace?seconds=5' -o qboxrsf_15618_20171102_131309_trace
qboxserver@cs3:~$
```

2. 查看暴露的变量

```
qboxserver@cs3:~$ curl http://127.0.0.1:40258/debug/vars
{
"NumGoroutine": 33,
"cmdline": ["/home/qboxserver/rsf/_package/qboxrsf","-f","qboxrsf.conf"],
"memstats": {....}
}
qboxserver@cs3:~$ curl http://127.0.0.1:40258/debug/var/cmdline
["/home/qboxserver/rsf/_package/qboxrsf","-f","qboxrsf.conf"]
```

3. 执行profile操作

执行dump脚本即可

profile方法参考：

* [Package pprof](https://golang.org/pkg/net/http/pprof/)
* [Profiling Go Programs](https://blog.golang.org/profiling-go-programs)


### 和Go官方的 `net/http/pprof` 有什么区别？

1. Go官方的pprof使用默认的`http.ServeMux`，如果服务也是用默认的ServeMux，可能会导致服务内部细节暴露，存在安全风险
2. 程序随机监听一个本地30000以上的随机端口，确保不占用普通服务端口，局域网及外网无法访问
3. 整合Go官方的`expvar`包，可以展示服务内部的一些数据
4. 提供一键dump所有信息功能，方便问题排查
5. 支持将程序内部信息导出到prometheus,
	1. 可以通过`service_metrics_push_gateway_addr`环境变量来设置pushgateway地址，默认`http://127.0.0.1:1056`
	2. 可以通过`service_metrics_push_duration_seconds`环境变量来设置发送间隔，默认5秒发送一次，设置小于0则关闭自动push功能
