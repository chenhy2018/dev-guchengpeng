订单系统接口
======

## 返回对象:

**seller** 对象：

```json
{
	"id":19,  //商家id
	"email":"evm@qiniu.com", //商家对应acc的邮箱(唯一索引，不允许重复)
	"title":"弹性计算", //商家标题,用于展示性
 	"name":"evm", //商家的名称，后面商品的model会使用次属性做校验，非唯一索引，但是建议不要重复
	"callback":"http://evm2.qiniu.com/trade/callback", //回调地址，充值支付回调使用
	"update_time":"2015-10-21T13:57:19+08:00", //更新时间
	"create_time":"2015-10-21T13:57:19+08:00", //创建时间
	"status":1 //商家状态 1:enable 0:disable
}
```

**product** 对象:

```json
{
	"id":19, //商品id
	"seller_id":18, //商家id 外键 **seller**对象的id
	"name":"compute_3", //商品名称
	"model":"evm:compute:c:3", //产品型号，支持根据产品型号获取产品信息， 推荐构成方式:seller_name:product_property
	"unit":1, //单位, 默认是按月 1: 年，2: 月 3: 周 4: 天 99: 一次性购买
	"price":0, //价格， 支持价格为0， 单位：元，精度5位，长度15位
	"property":"{\"cpu\":\"16 core\", \"memory\":\"64\"}", //产品属性
	"description":"4核1G内存", // 产品描述
	"update_time":"2015-10-20T11:25:55+08:00", // 更新时间
	"create_time":"2015-10-20T11:24:33+08:00", // 创建时间
	"start_time":"2015-12-20T00:00:00+08:00", // 产品上线时间
	"end_time":"2018-12-19T23:59:59+08:00", // 产品下线时间
	"status":1 // 产品状态  1:新建 2:在线 3:已失效 4:已删除
}
```

**order**  对象:

> 一个订单可以包涵同一商家， 同一订单类型的多个商品订单

```json
{
	"id":17, //订单id
	"order_hash":"47ed733b8d10be225eceba344d533586", //订单号
	"seller_id":19, //商家id
	"buyer_id":3774353, //买家uid
	"fee":339, //订单费用
	"actually_fee":339, //用户需要支付的费用
	"memo":"computer update", //订单说明
	"update_time":"2015-10-21T16:43:25+08:00", //更新时间
	"create_time":"2015-10-21T16:43:25+08:00", //创建时间
	"pay_time":"0001-01-01T08:00:00+08:00", //支付时间
	"status":1, //订单状态 1: 未支付 2: 已支付 3:作废
    “products":[
        product1,
        product2,
        ...
    ],
	"product_orders":
	[
		product_order1, // 单个商品订单
		product_order2,
		...
	]
}
```

**product_order** 对象:

```json
{
	"id":14, // id
	"product_id":21, // 产品id
	"seller_id":19, // 商家id
	"buyer_id":3774353, // 买家uid
	"order_id":17, // 总订单id
	"order_hash":"47ed733b8d10be225eceba344d533586", // 订单号
    "order_type":1, //订单类型  1: 新建 2:续费 3: 升级 4: 补偿 5:退款
    "product_order_id": 0, // 关联订单id，默认为0
	"product_name":"compute_4", // 商品名称
	"product_property":"{\"cpu\":\"4 core\", \"memory\":\"2G\"}", // 商品属性
	"property":"", // 订单属性
	"duration":2, // unit对应的倍数，如服务器按月购买，这边可以一次购买2个月
	"quantity":2, // 数量，同样配置
	"fee":339, // 费用
	"update_time":"2015-10-21T16:43:25+08:00", //更新时间
	"create_time":"2015-10-21T16:43:25+08:00", //创建时间
	"start_time":"0001-01-01T08:00:00+08:00", //服务开始时间
	"end_time":"0001-01-01T08:00:00+08:00", //服务结束时间
	"status": 1 // 订单状态： 1: 新建 2:完成
}
```

## 错误返回:

http 状态码非200多为错误，具体可能的错误码有：

+ 400: bad request
+ 404: not found
+ 401: unauthorized
+ 409: conflict
+ 500: internal server error

```
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
	"error":"product not found: 190" //错误描述
}
```


1. **系统中所有时间参数均为 go time 包中的 RFC3339**
2. 接口中所有和 **page** 以及 **page_size** 有关的参数， **page** 默认为0， **page_size** 最大为 200
 
## 1. 商家接口


### 1.1 新建商家

```
POST /seller/new HTTP/1.1
Authorization:{token}
Content-Type: application/x-www-form-urlencoded

name=gaea&callback=https://gaea.qiniu.io/callback&title=aaa

email: 用户email(required)
title: 标题 (required)
name: 商家名称(ex:cdn,evm,ufop...) (required)
callback: 回调地址(订单支付结果回调地址) (required)
```

#### 成功返回:

**seller**

回调接口说明：

1. 回调接口通过POST的方式请求
2. 回调通过 admin token 授权
3. 请求参数为订单以及商品信息，json 方式

example:

```
POST http://aaa/api/callback
Authorization: Bearer uaFRcD1RtcOLNe-uQivs2MPlYUrwEHbwAoVMN-o8Iigzt9Egm0GY1NteDSkQ_jDUW7U1vLTUYQfOuUj8ZXm6X7JGYNAUhMb8rs9yraTfueJLXQjEyMAD2wAbvxAnvUMl3spp_EVz3PhiZ3kBELVOjKVKbJe0akFgFz3G5WSe3wMvll4_8_LA9rQaTcg3ZSMnpB-v4D0AA3yQ2RhwsI8FjQ==
Content-Type: applciation/json

{"id":11,"order_hash":"13c8ffd977013703a701cf8e11deac65","seller_id":1,"buyer_id":1380416976,"fee":1,"actually_fee":1,"memo":"入门级购买","order_type":1,"update_time":"2016-09-21T11:04:42+08:00","create_time":"2016-09-21T11:04:28+08:00","pay_time":"2016-09-21T11:04:42+08:00","status":2,"product_orders":[{"id":10,"product_id":1,"seller_id":1,"buyer_id":1380416976,"order_id":11,"order_hash":"13c8ffd977013703a701cf8e11deac65","product_name":"入门型点播云","product_property":"{\"space\":20, \"tranfer\":60}","property":"","duration":1,"quantity":1,"fee":1,"update_time":"2016-09-21T11:04:28+08:00","create_time":"2016-09-21T11:04:28+08:00","start_time":"0001-01-01T00:00:00+08:00","end_time":"0001-01-01T00:00:00+08:00","status":1}],"products":[{"id":1,"seller_id":1,"name":"入门型点播云","model":"dora:vod:start","unit":1,"price":1,"property":"{\"space\":20, \"tranfer\":60}","description":"入门型","update_time":"2016-09-13T17:47:11+08:00","create_time":"2016-09-13T17:27:11+08:00","start_time":"2016-09-13T17:47:11+08:00","end_time":"2019-09-13T17:27:11+08:00","status":2}]}
```

正确返回:

```json
{
    "code": 200,
    "msg": "ok"
}
```

错误返回可通过 http statusCode 或者 body 中的 code 为非 200 表示

**错误会重试三次，每次重试后会 sleep 1 分钟**

### 1.2 根据商家id或者email获取商家信息

```
GET /seller/get?id=3 HTTP/1.1
Authorization:{token}

id和uid二选一
```

#### 成功返回：


**seller**


### 1.3 更新商家

```
POST /seller/update HTTP/1.1
Authorization:{token}
Content-Type: application/x-www-form-urlencoded

id=3&title =gaea123&callback=https://gaea.qiniu.io/callback&status=1

id: 商家id (requird)
title: 标题 (optional)
name: 商家名称(optional)
email: 商家邮箱(optional)
callback: 回调地址(订单支付结果回调地址) (optional)
status: 商家状态 (optional)
```

#### 成功返回:


**seller**


### 1.4 商家列表

```
GET /seller/list?page=0&page_size=10 HTTP/1.1
Authorization: {token}

page: 页数，从0开始 (optional)
page_size: 默认200 (optional)
```

#### 正确返回:

```json
[
	seller1,
	seller2,
	...
]
```

## 2. 商品接口

### 2.1 新建商品

```
POST /product/new HTTP/1.1
Authorization:Bearer {token}
Content-Type: application/x-www-form-urlencoded

"email=evm@qiniu.com&name=c3&model=evm:c:1&unit=1&price=511&description=memo&property={cpu:16 core,memory:16G,harddriver:1T,bandwidth:20M}"

email: 商家email (required)
name: 商品名称 (required)
model: 商品型号 (required)
property:属性(required)
price: 商品价格 (required) default: 0.0000元
unit: 商品单位 (optional)  default:yearly
description: 描述(optional)
start_time: 上线时间 (optional) 默认当前时间 格式RFC3339 "2006-01-02T15:04:05+07:00"
end_time: 下线时间 (optional) 默认start_time + 3年 格式同上
```

#### 正确返回：

**product**


### 2.2 更新商品信息

```
POST /product/update HTTP/1.1
Authorization: {token}
Content-Type: application/x-www-form-urlencoded

name=compute_8&model=evm%3Acompute%3Ac%3A4&property=%7B%22cpu%22%3A%224+core%22%2C+%22memory%22%3A%222G%22%7D&description=4%E6%A0%B81G%E5%86%85%E5%AD%98&id=21&price=56.50&unit=1

失效或者删除的产品不需要更新；更新产品的状态为下线或者失效时其他参数不更新;

id: (requird)
price: (optional)
status: (optional)
name: (optional)
model: 只有未上线的商品才能更新 (optional)
property: 同model (optional)
unit: 商品周期单位 (optional)
description: (optional)
start_time (optional)
end_time (optional)
```

#### 正确返回:

**product**

### 2.3 根据id或者型号获取商品信息

```
GET /product/get?id=3 HTTP/1.1
Authorization:Bearer {token}

id: (optional) 和model二选一
model: (optional)
```

#### 正确返回:

**product**

### 2.4 获取商家的商品列表

```
GET /seller/product?seller_id=19 HTTP/1.1
Authorization: {token}

seller_id:商家id (optional) 和email二选一
email: 商家email (optional) 
page: (optional)
page_size: (optional)
```

#### 正确返回:

```json
[
	product1,
	product2,
	...
]
```

### 2.5 根据id列表获取商品列表

```
GET /product/ids?ids=1,2,3 HTTP/1.1
Authorization: {token}

ids: (requird) 多个id逗号分隔
page: (optional)
page_size: (optional)
```

#### 正确返回:

```json
[
	product1,
	product2,
	...
]
```

### 2.6 商品列表

```
GET /product/list HTTP/1.1
Authorization: {token}

page: (optional)
page_size: (optional)
```

#### 正确返回:

```json
[
	product1,
	product2,
	...
]
```


### 2.7 上线商品

```
POST /product/release?id=1 HTTP/1.1
Authorization:{token}
```

#### 正确返回:

**product**

## 3. 订单接口

### 3.1 新建订单

> 订单支持同一类型同一商家多产品新建

```
POST /order/new HTTP/1.1
Authorization: {token}
Content-Type: application/x-www-form-urlencoded

data={"memo":"computer update", "order_type":2, "uid":3774353, "orders":[{"product_id":21, "duration":2,"quantity":3},{"product_id":22, "duration":2,"quantity":2}]}

uid: 用户id(required)
order_type: 订单类型(optional) default:1
memo: 订单备注(optional)
orders: []order订单数组

order包涵三个属性:
product_id: 产品id
duration: 时长（int）产品unit属性的倍数
quantity: 数量（int）
```

#### 正确返回:

```
{"order_hash":"58c89562f58fd276f592420068db8c09"}  // 订单号
```

### 3.2 获取订单信息

> 此接口支持根据订单id或者根据订单号获取订单信息，同时支持获取订单详情(购物清单).


```
GET /order/get?order=58c89562f58fd276f592420068db8c09 HTTP/1.1
Authorization: Bearer {token}

order_hash: 订单号(optional) 和id二选一
id: 订单id(optional)
with_detail: 是否包涵订单详细 (optional) default:false
```

#### 正确返回:

**order**

### 3.3 订单支付接口

```
POST /order/pay HTTP/1.1
Authorization: Bearer {token}
Content-Type: application/x-www-form-urlencoded

order_hash=58c89562f58fd276f592420068db8c09
```

#### 正确返回:

```
HTTP/1.1 200 OK
```


### 3.4 更新订单

```
POST /order/update HTTP/1.1
Authorization: Bearer {token}
Content-Type: application/x-www-form-urlencoded

id=234324&order=58c89562f58fd276f592420068db8c09&actually_fee=120.00&status=2

id: 订单id， 订单id和order_hash二选一
order_hash: 订单号，同上
actually_fee: 新的价格
status: 订单状态
```

#### 正常返回:

**order**

### 3.5 订单列表

```
GET /order/list?page=1&page_size=10 HTTP/1.1
Authorization: Bearer {token}

page: 页数 (optional) default:0
page_size: 每页条数(optional) default:200
```

#### 正常返回:


```json
[
	order1,
	order2,
	...
]
```

### 3.6 商家订单列表

```
GET /seller/order/list?seller_id=20&uid=13233442&page=1&page_size=10 HTTP/1.1
Authorization: Bearer {token}

seller_id: 商家id， 和 email 二选一
email: 商家email 同上
page: 页数 (optional) default:0
page_size: 每页条数(optional) default:200
```

#### 正常返回:


```json
[
	order1,
	order2,
	...
]
```

### 3.7 用户订单列表

```
GET /user/order/list?page=1&page_size=10 HTTP/1.1
Authorization: Bearer {token}

uid: 买家uid 
page: 页数 (optional) default:0
page_size: 每页条数(optional) default:200
```

#### 正常返回:


```json
[
	order1,
	order2,
	...
]
```

### 4.1 更新单个商品订单的信息和状态

```
POST /product/order/accomplish HTTP/1.1
Authorization: Bearer {token}
Content-Type: application/x-www-form-urlencoded

id=234324&property={"instanceId":"ferwUFRDdsdf3rDSfsfrwrRcdsdf"}

id: 商品订单id
property: 属性 (optional)
start_time: 生效时间 (optional), 只对新购买的商品有效
```

#### 正常返回:

**product_order**


### 4.2 订单升级

升级规则：
1. 同一商家的商品
2. 商品 unit 属性一致
3. 如果订单 quantity 大于1, 则升级保持 quantity一致， 也就是说一次购买多件商品(同一商品，数量大于1)，升级时需要一起升级
4. 升级时间必须在当前商品的有效期内 end_time>=upgrade_time >= start_time
5. 升级后订单时间为原有订单的剩余时间， 如当前订单购买了1年，已经使用了3个月，则升级订单的使用时间为9个月
6. 不支持降级，即升级的商品单价大于原有订单商品单价
```
POST /product/order/upgrade 
Authorization: Bearer ${token}
Content-Type: application/x-www-form-urlencoded

buyer_id=1380416976&current_id=5&product_id=2&start_time=2016-10-01T00:00:00%2B08:00&memo=套餐升级2

buyer_id: 买家 uid (required)
current_id: 需要升级的订单id (required)
product_id: 需要升级到的产品id (required)
start_time: 生效时间 (optional, 当前时间)
memo: 说明 (optional)
```

#### 正确返回:

```
{"order":"58c89562f58fd276f592420068db8c09"}  // 订单号
```


### 4.2 商品订单列表

```
GET /product/order/list?product_id=1&buyer_id=1380416976&page_size=1
Authorization: Bearer ${token}

```

参数： 
+ product_id: 商品 id
+ seller_id: 商家 id
+ buyer_id: 买家 id
+ order_hash: 订单号
+ page: 页数
+ page_size: 分页大小

正常返回:

```json
[
  {
    "id": 1,
    "product_id": 1,
    "seller_id": 2,
    "buyer_id": 1380416976,
    "order_id": 2,
    "order_hash": "9e688c58a5487b8eaf69c9e1005ad0bf",
    "order_type": 1,
    "product_order_id": 0,
    "product_name": "入门型点播云",
    "product_property": "{\"space\":20, \"tranfer\":60}",
    "property": "",
    "duration": 12,
    "quantity": 2,
    "fee": 24,
    "update_time": "2016-09-13T17:50:17+08:00",
    "create_time": "2016-09-13T17:50:17+08:00",
    "start_time": "0001-01-01T00:00:00+08:00",
    "end_time": "0001-01-01T00:00:00+08:00",
    "status": 1
  }
]
```

