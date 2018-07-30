AuthGate API
===============

# 管理协议

## reload (重新读配置项)

请求包：

```
POST /reload?host=<Host>
```

返回包：

```
200 OK
```

## query (查询服务器状态)

请求包：

```
POST /query?host=<Host>&server=<Ip:Port>
```

返回包：

```
200 OK {
	conn: <ActiveConnectionCount>
}
```

## enable (允许/禁止服务器服务)

请求包：

```
POST /enable?host=<Host>&server=<Ip:Port>&state=<Enabled>
```

* 注：这个只是修改内存中的状态，并不修改db中的数据

返回包：

```
200 OK
```

使用场景（升级服务器）：

1. 首先将要升级的服务器(`<Ip:Port>`) 禁止（`/enable?state=0`）；
2. 定期查询该服务器的连接数（`/query`），直到该服务器没有活跃的连接（`conn == 0`），或者超时；
3. 杀死服务器进程，升级；
4. 重新允许服务器(`<Ip:Port>`) 提供服务（`/enable?state=1`）

