Memory-based Message Queue API
==================

# 协议

## put (发送消息)

请求包：

```
POST /put/<MqId>/id/<MsgId>
Content-Type: application/octet-stream

<MessageData>
```

返回包：

```
200 OK
801 ErrMQFull
570 GracefulQuit
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

<MessageData>
```

## delete (删除消息)

请求包：

```
POST /delete/<MqId>/id/<MsgId>
```

返回包：

```
200 OK
```

## stat (统计MQ状态)

请求包：

```
POST /stat/<MqId>
```

返回包：

```
200 OK {
	todo: <TodoLen>
	doing: <ProcessingLen>
}
```

