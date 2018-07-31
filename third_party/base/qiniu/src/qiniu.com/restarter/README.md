## restarter

### 目的

令服务在升级或者重启的时候可以外部无感知

### 条件

1. 机器上要有restarter
 - `deploy/floy/restarter/restarter.csv`
1. 让服务由restarter来启动，作为restarter的子进程
 - 参考 `deploy/floy/fopagent/supervisor.conf`
1. 服务本身需要支持同时两个或以上一模一样的程序在机器上运行（配置一样，工作命令一样等）
 - bind相同端口会冲突，所以服务需要使用`SO_REUSEPORT`参数来监听tcp, `base/qiniu/src/github.com/qiniu/http/httputil.v1/reuseport_example.go`
 - 如果是非http的也参照上面那一条
 - 注意程序其他设定是否和同时启动会有冲突
 - 服务自己要处理退出信号，不打扰正在的处理的请求，安全退出

### 重启服务

向对应的restarter进程发送`SIGUSR2`信号，这时候restarter会根据原始命令来启动一个新的进程，新的进程启动成功后会关闭旧的进程。
因为有一段时间是两个服务同时在线的，根据`SO_REUSEPORT`在linux下面的行为，会外部的请求会分发到两个服务上面。
然后把旧的服务关闭这时候请求就全都到新的服务上面，结果就是做到热重启

TODO. 当前重启服务没有成功或者失败的反馈，需要添加

### 关闭服务

正常使用`supervisorctl stop xxx`来关闭服务就可以

### 注意事项

#### 句柄监控

使用了restarter之后，因为监控看到的pid是restarter，他的句柄无法体现服务的真正句柄，需要找运维把这个服务的句柄监控改为监控这个服务以及它所有子进程的句柄数

#### 服务首次使用`SO_REUSEPORT`

因为之前的tcp参数和旧的tcp参数不一致，所以当旧的服务关闭之后，会有一些连接处于`TIME_WAIT`状态，这时候新服务会有一段时间无法启动。
首次之后就没有这个问题，因为参数一致了

### .

欢迎给这里提交代码
