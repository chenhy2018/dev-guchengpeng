Mocking Network System
==========

## 节点进程(Node)

1、替换 "net"、"net/http" 为 "qiniupkg.com/mocking/net"、"qiniupkg.com/mocking/net/http"。

2、func main 函数的开始，调用 "qiniupkg.com/mocking" 的 Init 函数进行初始化（这一步也可能会设计为自动化完成）。

实际上，qiniupkg.com/mocking 会解析该 app 的命令行参数：

* `-node`: 本Node的名字（NodeName）
* `-top`: 集群的拓扑（Topology）

其中 topology 是 urlsafe base64 编码后的 json 内容。示意如下：

```
{
	"idcs": [
		{
			"name": <IdcName1>,
			"nodes": [
				{
					"name": <NodeName1>,
					"ips": {
						"tel": "<LocalIP11>",
						"bgp": "<LocalIP12>"
						...
					},
					"defaultIsp": "tel", # for net.Dial
					"workdir": <WorkDir1>,
					"exec": [<App1>, <Arg11>, ..., <Arg1N>]
				},
				...
			]
		},
		...
	],
	"speeds": [
		{
			"from": <IdcNameFrom:Isp>,
			"to": <IdcNameTo:Isp>,
			"speed": [<SpeedItem1>, ...] #详细见net.Speed的定义，在len(speed)=0时表示两个Idc不连通
		},
		...
	],
	"defaultSpeed": [<SpeedItem1>, ...]  #不在speeds列表里面的都会使用defaultSpeed
}
```

## 规格

```go
func Init() {} // 由节点进程(Node)的main函数调用，用来初始化mocking/net、mocking/net/http、etc。

type Cluster struct {
	...
}

func (p *Cluster) Shutdown(name string, wait bool) (err error) {}

func RunCluster(topology string) (cluster *Cluster) {}
```

