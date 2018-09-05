qtcpgate - tcp server gate
==================

## 应用场景: 热升级

### 初始环境

```bash
$ cat instance1/qfooserver.conf

{
	"bind_host": "127.0.0.1:8888",
	"proxy_host": "127.0.0.1:8889",
	...
}

$ cd instance1; ./qfooserver -f ./qfooserver.conf

$ cat qtcpgate.conf

{
	"bind_host": ":9999",
	"admin_host": "127.0.0.1:1111",
	"debug_level": 1
}

$ qtcpgate -f qtcpgate.conf start
$ qtcpgate -f qtcpgate.conf add 127.0.0.1:8889
```

这样用户就可以通过 9999 端口访问 qfooserver 的服务。

### 升级过程

```bash
$ cat instance2/qfooserver.conf

{
	"bind_host": "127.0.0.1:9000",
	"proxy_host": "127.0.0.1:9001",
	...
}

$ cd instance2; ./qfooserver -f ./qfooserver.conf
$ qtcpgate -f qtcpgate.conf add 127.0.0.1:9001
$ qtcpgate -f qtcpgate.conf rm 127.0.0.1:8889
```

确认是否升级已经完成：

```
$ qtcpgate -f qtcpgate.conf list
```

## 命令行手册

### 启动服务

```bash
qtcpgate -f qtcpgate.conf start
```

* 在 qtcpgate 的 backend 服务器列表发生变化时，qtcpgate 都会进行持久化。
* start 的含义是恢复此前持久化的 backend 服务器列表。如果之前没有进行过 backend 服务器的添加操作，则 start 会打印一个没有 backend 服务器的告警，并建议你调用 `qtcpgate add <backend-host>` 来添加 backend 服务器。

### 添加backend服务器

```bash
qtcpgate -f qtcpgate.conf add <backend-host> [<backend-pid>]
```

* `<backend-host>`: backend服务器地址。
* `<backend-pid>`: backend服务器的进程id。可选。这个参数仅对于那些和qtcpgate在同一机器上的backend服务器有意义。

### 移除backend服务器

```bash
qtcpgate -f qtcpgate.conf rm <backend-host>
```

注意：

* 移除backend时，服务器并不是一下子就直接被拿掉，而是进入冷却状态，直到没有连接才真正被删除。
* 考虑到 tcp 连接可能很长时间不会结束，这意味着服务器冷却时间可能非常长（几个小时甚至一天，所以需要有机会能够查询状态）。
* 如果backend添加的时候指定了pid，则qtcpgate会向该backend服务器发kill信号。

### 单例化backend服务器

```bash
qtcpgate -f qtcpgate.conf singleton <backend-host> [<backend-pid>]
```

这个操作等价于：

* 用 list 获得所有已经存在的 backend hosts；
* 用 add 注册自己；
* 用 rm 删除所有前面用 list 获得的 backend hosts；

### 查询backend服务器状态

```bash
qtcpgate -f qtcpgate.conf list
```

显示的结果：

```
[
	{
		"host": <backend-host1>,
		"pid": <backend-pid1>,
		"date": <datetime-add-backend>,
		"state": <normal|frozen>
	},
	...
]
```
