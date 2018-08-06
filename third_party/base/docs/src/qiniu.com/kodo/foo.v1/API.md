FOO API 协议
=====

# 协议

## 创建foo对象

请求包：

```
POST /v1/foos
Content-Type: application/json
Authorization: Qiniu <MacToken>

{
	"a": <A>,
	"bar": <Bar>
}
```

返回包：

```
200 OK
Content-Type: application/json

{
	"id": <FooId>
}
```

## 列出foo对象

请求包：

```
GET /v1/foos
Authorization: Qiniu <MacToken>
```

返回包：

```
200 OK
Content-Type: application/json

[
	{
		"a": <A>,
		"bar": <Bar>
	},
	...
]
```

## 删除foo对象

请求包：

```
DELETE /v1/foos/<FooId>
Authorization: Qiniu <MacToken>
```

返回包：

```
200 OK
```

## 获取foo对象

请求包：

```
GET /v1/foos/<FooId>
Authorization: Qiniu <MacToken>
```

返回包：

```
200 OK
Content-Type: application/json

{
	"a": <A>,
	"bar": <Bar>
}
```

## 修改foo的bar属性

请求包：

```
POST /v1/foos/<FooId>/bar
Content-Type: application/json
Authorization: Qiniu <MacToken>

{
	"val": <Bar>
}
```

返回包：

```
200 OK
```

## 获取foo的bar属性（可匿名获取）

请求包：

```
GET /v1/foos/<FooId>/bar
```

返回包：

```
{
	"val": <Bar>
}
```

