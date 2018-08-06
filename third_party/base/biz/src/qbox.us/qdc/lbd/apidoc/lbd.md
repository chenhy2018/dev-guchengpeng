Local Block Device
======================

* qboxlbd 本身的代码已经冻结，如需修改，请改 qboxdc 的最新版本

# 协议

## put (写数据)

请求包：

```
POST /put?key=<Sha1>&len=<Bsize>
Content-Type: application/octet-stream

<BlockData>
```

返回包：

```
200 OK
Content-Type: application/octet-stream

<Sha1>
```


## put_local (写数据)

请求包：

```
POST /put_local?key=<Sha1>&len=<Bsize>
Content-Type: application/octet-stream

<BlockData>
```

返回包：

```
200 OK
Content-Type: application/octet-stream

<Sha1>
```

## get (读数据)

请求包：

```
POST /get?key=<Sha1>&from=<From>&to=<To>&idc=<Idc>
```

返回包：

```
200 OK
Content-Type: application/octet-stream

<BlockData>
```


## get_local (读数据)

请求包：

```
POST /get_local?key=<Sha1>
```

返回包：

```
200 OK
Content-Type: application/octet-stream

<BlockData>
```

## service-stat (服务状态)

请求包：

```
POST /service-stat
```

返回包：

```
200 OK {
	missing: <Missing>
	total: <Total>
	wtotal: <Wtotal>
}
```

