# Probe 监控系统 #

https://github.com/qbox/proposal/issues/1

## 概念 ##

### 数据点（point) ###

* measurement:string 唯一名称
* tag:map[string]string 类别
* field:map[string]interface{} 值：string|bool|float64
* time:time 时间戳

## SDK ##

```Go
	import "time"
	import probe "qiniu.com/probe/collector"

	probe.Mark("test", time.Now(),
		[]string{"key1", "key2"}, []string{"tag1", "tag2"},
		[]string{"v1", "v2", "v3", "v4"},
		[]interface{}{
			"abcdefghijklmnopqrstuvwxyz",
			100,
			123.321,
			true,
		},
	)
```
