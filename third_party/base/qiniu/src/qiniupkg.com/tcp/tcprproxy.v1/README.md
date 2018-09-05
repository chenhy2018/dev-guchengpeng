tcprproxy - tcp reverse proxy
==============================

这个tcp反向代理服务的特点是：

* 能够冷却后端的业务服务器，方便进行服务的热升级（非停服升级）；
* 能够传递客户端地址(RemoteAddr)给业务服务器；

支持业务服务器的热升级，这个详细可以参考 qtcpgate 实用程序。

支持传递客户端地址(RemoteAddr)给业务服务器，从实现原理来说是这样的：

* reverse proxy 收到流后，在流的开头添加了RemoteAddr信息，然后转给业务服务器；
* 业务服务器识别到这是 reverse proxy 服务器转发的流后，提取流开头的RemoteAddr信息；

对业务服务器的侵入性：

* 业务服务器不需要修改业务代码，只需要listen一个额外的端口（比如正常业务端口为port，收proxy转发的端口可以是port+1。之所以需要监听额外的端口是为了区分这是proxy转发过来的数据）。这个额外的端口通过一个统一的代码实现(见 tcprproxy.ListenAndServe)即可。
* reverse proxy 服务代理转发到 port+1 端口，而不是正常的业务服务 port 端口。

举一个例子。

假设我们现在有业务服务器：

```go
type FooService struct {
}

func (p *FooService) Serve(l net.Listener) (err error) {
	...
}
```

那么我们监听业务端口和代理转发端口：

```go
import "qiniupkg.com/tcp/tcprproxy.v1"
import "qiniupkg.com/tcp/tcputil.v1"

service := &FooService{...}
go tcputil.ListenAndServe(":1000", service, nil)
go tcprproxy.ListenAndServe(":1001", service, nil)
```

然后我们架设 qtcprproxyd 服务器：

```
$ cat qtcprproxyd.conf

{
    "bind_host": "0.0.0.0:1000",
    "backends": ["192.168.0.10:1001", "192.168.0.11:1001", "192.168.0.12:1001"],
    "max_procs": 0,    #0表示使用系统默认
    "debug_level": 1
}

$ qtcprproxyd -f qtcprproxyd.conf
```

