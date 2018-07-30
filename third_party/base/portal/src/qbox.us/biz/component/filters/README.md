
#Audit log for teapot app

##背景
audit log是七牛自定义的一种日志格式， 用来记录接收前端发起的http req和返回给前端的http rsp。日志记录格式如下：

```
REQ     ONE     14367775957550359       POST    /domains        {"Accept-Encoding":"gzip","Content-Length":"28","Content-Type":"application/x-www-form-urlencoded","Host":"192.168.34.200:23100","IP":"192.168.34.30","User-Agent":"QiniuGo/6.0.6 (linux; amd64; ) go1.3.1","X-Forwarded-For":"192.168.33.41","X-Real-Ip":"192.168.33.41","X-Reqid":"kGEAANykV2icdfAT","X-Scheme":"http"}       {"owner":"1380307205","tbl":"romhome"}  200     {"Content-Length":"62","Content-Type":"application/json","X-Log":["ONE:8"],"X-Reqid":"kGEAANykV2icdfAT"}        ["romhome.qiniudn.com","img.romhome.com","static.romhome.com"]  62      81283
```

这个日志格式是固定的。七牛的日志同步系统，会定时将audit log上传到hdfs， 然后数据统计系统根据audit log来提取统计信息，最后在tsdb.qiniu.io上展现统计图表。**因此，audit log是七牛后台系统，日志管理、数据统计以及监控非常重要的组成部分。**

在github库：qbox/base的audit包中，对audit log的生成进行了封装（包地址：qbox/base/biz/src/qbox.us/http/audit）。引入该包后，程序将自动提取http req和http rsp中的内容，生成audit log。

##问题
目前只有基于golang自带http框架或七牛自研http框架的go svr，才能使用audit log包， 而基于teapot框架的web svr无法使用该包。 原因在于audit log包是和http框架紧密耦合在一起的（其实现原理是在初始化audit log包时，往http框架注册一个audit log特有的Handler函数，后续处理http请求时，处理流程会先转移到该函数，在函数内部调用实际的Hanlder函数来处理请求，并log下req和rsp包到文件）。而teapot实现了自己的一套http请求处理高阶框架，无法再复用该audit log包。
这就导致，基于teapot框架的web svr，无法方便地生成audit log，进而无法方便复用七牛现有的日志管理、数据统计和报表展现系统。

##解决方法
参考base库里的audit log包，重写了一个teapot app专用的audit log包。该方法能够达到以下效果：
1.提供一个简单的audit log生成方式， 供teapot app方便地生成audit log
2.不修改teapot框架源码
3.生成的audit log格式和现有的保持完全一致，并包含用户信息（uid，utype等）

此次修改，新增了以下代码：
```
qbox/base/portal/src/qbox.us/biz/component/filters/auditlog.go
qbox/base/biz/src/qbox.us/http/audit/teapotlog
```

##使用方法
使用该teapot app专用的audit log包， 需要进行三个步骤的操作：

###1.在配置文件：conf/app.ini中增加audit log配置：
```
[auditlog]
#audit log地址：
logdir = ./run/auditlog/qboxfusion
chunkbits = 29
#需要log的rsp包的最大长度：
bodylimit = 256
#audit log中的模块名称：
modulename = FUSION
```

###2.在global/setting.go的Setting结构中，中增加以下内嵌struct：
```
	AuditLog struct {
		LogPath    string `conf:"logdir"`
		ChunkBits  uint   `conf:"chunkbits"`
		BodyLimit  int    `conf:"bodylimit"`
		ModuleName string `conf:"modulename"`
	} `conf:"auditlog"`
```

###3.在env/filters.go中，修改filter注册代码：
将：
```
	tea.Filter(
		// 在静态文件之后加入，跳过静态文件请求
		reqlogger.ReqLoggerFilter(logOut, loggerOption),
		// 在 action 里直接返回一般请求结果
		teapot.GenericOutFilter(),		
	)
```

修改为：
```
	tea.Filter(
		// 在静态文件之后加入，跳过静态文件请求
		reqlogger.ReqLoggerFilter(logOut, loggerOption),
	)

	auditLogCfg := filters.AuditLogConfig{
		LogDir:     global.Env.AuditLog.LogPath,
		ChunkBits:  global.Env.AuditLog.ChunkBits,
		BodyLimit:  global.Env.AuditLog.BodyLimit,
		ModuleName: global.Env.AuditLog.ModuleName,
	}

	tea.Filter(
		filters.UseAuditLog(tea, auditLogCfg),
	)

	tea.Filter(
		// 在 action 里直接返回一般请求结果
		teapot.GenericOutFilter(),
	)
```

即：将UseAuditLog这个filter函数，注册到teapot框架里。后续在处理每个http req时，都将调用该函数来生成audit log。

