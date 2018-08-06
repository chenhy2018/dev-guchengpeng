# 自动设置 User-Agent

### 用途

自动设置rpc请求的User-Agent，便于统计访问来源。

例如：

```
"User-Agent":"qboxio/12345 (linux/amd64; go1.7.1) nb777/8148"
```

对应信息：

```
"User-Agent":"程序名/程序版本号 (OS/OSARCH; Go版本) 主机名/Pid"
```

* 程序版本号默认使用 `github.com/qiniu/version` 的版本号，如果没设置则默认使用二进制的md5前10位

### 如何使用

在导入库里面增加一行： `_ "qbox.us/autoua"`

```go
import (
	_ "qbox.us/autoua"
)
```

* 注意：使用 autoua 的话建议不要再更改 `UserAgent`，否则可能会导致 data race
