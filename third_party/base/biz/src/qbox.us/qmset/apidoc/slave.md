Qmset Slave API
===============

# Qmset协议

## adds (批量添加集合元素)

请求包：

```
POST /adds?c=<GrpName>&kv=<KeyValue1>&kv=<KeyValue2>&kv=...
```

* 其中 `<KeyValue>` 是 `<Key>:<Value>` 这样一个字符串（这就假定了`<Key>`里面不能有:号）

返回包：

```
200 OK
```

# Qmbloom协议

同 master


# 通用协议

## flips (翻转)

* 来自 master，用以通知所有 slave 进行 flip（翻转）。

请求包：

```
POST /flips?c=<GrpName1>&c=<GrpName2>&clear=<DoClear>
```

* `clear=<DoClear>`: 在 master 重启的时候（前提是 master 没有做持久化），也会通知 slave 进行 flip，而且是放弃所有数据。

返回包：

```
200 OK
```

