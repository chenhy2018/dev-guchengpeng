qiniu apigate
===================
![Coverage](https://aone.qiniu.io/api/coverage/badge.svg?token=A59B8029-E9A1-4FF3-B72B-3657CEAC64D4&repo=qbox/apigate)

# 概述

要解决的需求：

* api 性能测量。
* api 监控的配合。统一的 api 请求异常监控（请求量下降监控、请求错误码个数监控）。
* api 的开放权限。分为：授权访问（又细分为哪些等级用户可以访问，包括管理员也是等级之一）、公开（public/匿名可访问)。
* 授权解析。解析结果会将 Qiniu => QiniuStub, UpToken => UpTokenStub，等等。
* 审计日志、ReqId 生成。
* ...

相应地，服务实现方与以前相比轻盈了很多。因为：

* 不用再考虑 auth/account 问题。
* 不用再测量 api 性能。
* 不用再考虑审计日志。

# 机制

* 每台 api service 机器都部署一个 apigate。每个 apigate 只转发给本机的 api service。
* 每台 api service 只 listen 一个局域网 ip 或者只 listen localhost（如果没有局域网请求需求的话）。
* 所有 apigate 的配置一致，但是后面的 api service 可能不同，这会导致如果请求发往错误的 apigate 会得到 method not found 这样的错误。
* 如果希望局域网内非授权方式访问 api service，则可以绕过 apigate 直接访问 api service 本身的服务端口。

# 使用

## qiniugate

真实环境下的 apigate 程序。详细参考：[qiniu.com/apigate.v1/qiniugate](https://github.com/qbox/apigate/tree/develop/src/qiniu.com/apigate.v1/qiniugate)

## qiniumockgate

仿真环境下的 apigate 程序（mock主要是指不连接真实的账号授权服务，而是直接用 [authstub](https://github.com/qbox/base/blob/develop/qiniu/src/qiniu.com/http/httptest.v1/qiniutest/README.md#authstub) 授权方式）。详细参考：[qiniu.com/apigate.v1/qiniumockgate](https://github.com/qbox/apigate/tree/develop/src/qiniu.com/apigate.v1/qiniumockgate)

## 配置文件 (json)

```json
{
  "services": [
    {
      "module": "RS",
      "routes": ["rs.qiniuapi.com", "api.qiniu.com:8888/rs", ":8888/rs"],
      "auths": ["qbox/macbearer"],
      "forward": "localhost:8888",
      "apis": [
        {
          "pattern": "POST /stat/**",
          "allow": "user"
        },
        {
          "pattern": "POST /mkbucket/**",
          "allow": "admin"
        },
        {
          "pattern": "GET /bucketinfo/**",
          "allow": "public"
        },
        ...
      ]
    },
    {
      "module": "UP",
      "routes": ["*"], # 默认路由，在其他路由都不匹配情况下会进入到这里
      "auths": ["qbox/macbearer"],
      "forward": "localhost:9999",
      "apis": [
        {
          "auths": ["qbox/mac", "qbox/uptoken"],
          "pattern": "POST /mkblk/**",
          "allow": "user",
          "notallow": "expuser" #这里假设体验用户不能调用mkblk这个api
        },
        {
          "auths": ["qbox/multipart-uptoken"],
          "pattern": "POST /**",
          "forward": "/upload",
          "allow": "user"
        },
        {
          "auths": [],
          "pattern": "POST /",
          "proxy": "qbox/multipart-token" # 解析multipart上传中的token字段
        },
        ...
      ]
    },
    {
      "module": "OS",
      "routes": ["os.qiniuapi.com"],
      "auths": ["qiniu/mac"], #新服务不再支持qbox/bearer授权，另外用qiniu/mac而不是qbox/mac授权
      "forward": "localhost:8899",
      "apis": [
        {
          "pattern": "POST /v1/auths/*/rules",
          "allow": "user"
        },
        {
          "pattern": "GET /v1/auths/*",
          "allow": "user"
        },
        ...
      ]
    },
    {
      "module": "OSFWD",
      "routes": ["osfwd.qiniuapi.com/trim"], # 所有转发的路径的 /trim 字段会被去掉
      "auths": ["qbox/mac"],
      "forward": "localhost:7777/fwd", # 所有转发的路径都会加上 /fwd 的前缀
      "apis": [
        {
          "pattern": "POST /v1/auths/*/rules",
          "allow": "user"
        },
        {
          "pattern": "GET /v1/auths/*",
          "allow": "user"
        }
      ]
    }
  ]
}
```
