qiniumockgate
===========

仿真环境下的 [apigate](https://github.com/qbox/apigate) 程序（mock主要是指不连接真实的账号授权服务，而是直接用 authstub 授权方式）。

命令行格式为：

```bash
qiniumockgate -f <qiniumockgate.conf>
```

详细的 conf 文件格式参考：

* [qiniu.com/apigate.v1/qiniumockgate/qiniugate.conf](https://github.com/qbox/apigate/blob/develop/src/qiniu.com/apigate.v1/qiniumockgate/qiniumockgate.conf)

这个 conf 还会引用一个 apigate.conf，其格式参考：

* [qiniu.com/apigate.v1/qiniumockgate/apigate.conf](https://github.com/qbox/apigate/blob/develop/src/qiniu.com/apigate.v1/qiniumockgate/apigate.conf)

### 使用案例

* [qiniu.com/examples/apigate.v1/apigate_example](https://github.com/qbox/apigate/tree/develop/src/qiniu.com/examples/apigate.v1/apigate_example)
