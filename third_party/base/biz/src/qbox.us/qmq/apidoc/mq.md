Message Queue API
==================

# 协议

## put (发送消息)

请求包：

```
POST /put/<MqId>
Content-Type: application/octet-stream
Authorization: <Token>

<MessageData>
```

返回包：

```
200 OK
X-Id: <MsgId>
```

## get (获取消息)

请求包：

```
POST /get/<MqId>
```

返回包：

```
200 OK
X-Id: <MsgId>
Content-Type: application/octet-stream
Authorization: <Token>

<MessageData>
```

## delete (删除消息)

请求包：

```
POST /delete/<MqId>
X-Id: <MsgId>
Authorization: <Token>
```

返回包：

```
200 OK
```

# 管理协议

## admin-make (创建MQ)

请求包：

```
POST /admin-make/<MqOwnerUidBase36-MqId>/expires/<Expires>
Authorization: <AdminToken>
```

* 一个消息被get后在expires时间内如果没有被delete，则可能会被重新get到

返回包：

```
200 OK
```

## admin-filter (过滤消息)

请求包：

```
POST /admin-filter/<MqOwnerUidBase36-MqId>/by/<Uid>/to/<MqOwnerUidBase36-MqId2>
Authorization: <AdminToken>
```

* 将 mq 中所有某个 `<Uid>` 的消息都过滤出来，放到目标的 mq 中

返回包：

```
200 OK
```

## admin-stat (消息统计)

请求包：

```
POST /admin-stat/<MqOwnerUidBase36-MqId>
Authorization: <AdminToken>
```

返回包：

```
200 OK {
	<Uid>: <MqItemCount>
}
```

