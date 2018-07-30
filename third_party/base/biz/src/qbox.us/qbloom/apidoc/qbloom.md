Bloom Filter API
==================

# 部署结构

* 为了高可用，建议写是双写单读


# 协议

## set (设置)

请求包：

```
POST /set?v=<UrlsafeBase64EncodedValue>
```

返回包：

```
200 OK
```

## test (测试)

请求包：

```
POST /test?v=<UrlsafeBase64EncodedValue>
```

返回包：

```
200 OK
612 NoSuchEntry
```

## tas (测试并设置)

请求包：

```
POST /tas?v=<UrlsafeBase64EncodedValue>
```

返回包：

```
200 OK
612 NoSuchEntry
```

