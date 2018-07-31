基于 qtcpgate 的热升级机制
=================

## 关于qtcpgate

* 参考：[TCPGATE.md](TCPGATE.md)

## 热升级流程

假设业务服务器叫做 qfooserver，我们看下热升级应该怎么做。

### 初始过程

首先，我们启动 qtcpgate 进程：

```
$ cat qtcpgate.conf

{
	"bind_host": ":9999",
	"admin_host": "127.0.0.1:1111",
	"debug_level": 1
}

$ qtcpgate -f qtcpgate.conf start
```

然后，我们启动业务服务器，并申请成为 qtcpgate 的 backend 服务器：

```
$ cat qfooserver.conf

{
	// "bind_host": "127.0.0.1:8888",   #不应该存在，应该是随机端口
	// "proxy_host": "127.0.0.1:8889",  #不应该存在，应该是随机端口
	"tcpgate_admin": "127.0.0.1:1111",
	...
}

$ qfooserver -f qfooserver.conf
```

qfooserver 启动时会通过 `qtcpgate -f qtcpgate.conf singleton <backend-host> <backend-pid>` 向 qtcpgate 注册自己（实际上这事是通过 qtcpgate 服务的网络 api 实现的，这里的写法只是表意）。

### 升级流程

替换掉 qfooserver 的二进制文件，并且对配置文件作出必要的修改。然后执行以下命令即可：

```
$ qfooserver -f qfooserver.conf
```

## 兼容性

### supervisor 兼容性

* 当 qfooserver 挂掉被 supervisor 重启时，它会向 qtcpgate 发起注册自己的行为，并试图让原先的 qfooserver 下线（但因为原先的进程已经不存在，故无动作）。
* 当 qtcpgate 挂掉被 supervisor 重启时，因为 qtcpgate 在 backend 发生变化的时候都会持久化，故此启动后仍然可以正常工作。

### 回滚机制的兼容性

* 蛮多业务基于软链接来实现版本管理和升级，本机制并兼容该做法。虽然我们前面举例时用的是替换可执行文件和conf的方式，但是多版本情况下仍然可以正常工作。

### 业务服务器的编程逻辑

简单说，所有涉及到互斥性的系统资源的申请，都需要被改写，例如：

* 监听端口上，以前业务服务端口通常在配置文件中，现在需要改为随机端口（对业务有侵入，有的应用可能写死了端口，需要相应作出调整）。
* 日志文件上，同理不能再在配置文件中写死日志文件名，最多指定一个日志的basedir，然后在这个目录中以某种方式生成唯一的日志文件名。可选的方法有：基于进程id、基于当前时间、基于自增值。
* 其他持久化数据同样不能写死，处理方式可以和日志文件相同。
* 向 qtcpgate 注册的 backend 代理端口同样需要是随机的，当然这一点不影响正常的业务逻辑。
