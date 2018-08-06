Qmset Master API
===============

# 算法

维持两个集合：set1, set2

* get(k): 向 set1 取数据。
* add(k, v): 同时 set1, set2 进行 add(k, v)。
* 在每过 `<ExpiresInSeconds>` 时间后执行 flip 操作：swap(set1, set2); set2 清空。同时通知 slave 也执行 flip。


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

## get (读取一个集合)

请求包：

```
POST /get?c=<GrpName>&k=<SetId>
```

返回包：

```
200 OK [
	<Value1>,
	<Value2>,
	...
]
```

# Qmbloom协议

## badd (添加元素)

请求包：

```
POST /badd?c=<GrpName>&v=<Value1>&v=<Value2>&v=...
```

返回包：

```
200 OK
```

## bchk (检查元素是否存在)

请求包：

```
POST /bchk?c=<GrpName>&v=<Value1>&v=<Value2>&v=...
```

返回包：

```
200 OK [
	<Index1>,
	<Index2>,
	...
]
```

* 返回的 `<Index>` 是那些存在的元素的下标

