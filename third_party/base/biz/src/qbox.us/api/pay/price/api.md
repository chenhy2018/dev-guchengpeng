
## 接口

### 1. &nbsp; 定义价格 &nbsp;&nbsp;&nbsp;&nbsp; **/set**
* 权限
  * admin
* 参数

***-- Request --***

	POST /set HTTP/1.1
	Authorization: <Authorization>
	Content-Type: application/json
	Content-Length: xxx
	
	{
	 "uid": <uint32>,       // 用户ID
	 "infos": [
	           {
                 "type": <string>, // base/discount/package
	             "effecttime": <int64>, // 起始有效时间， 100ns。不填或者小于0表示立即生效。
	             "deadtime": <int64>,   // 结束有效时间， 100ns。不填或者小于0表示永久有效。
                 "info": <string>  // 具体类型的json字符串，详见下
               },
	           {
                 "type": <string>, // base/discount/package
	             "effecttime": <int64>, // 起始有效时间， 100ns。不填或者小于0表示立即生效。
	             "deadtime": <int64>,   // 结束有效时间， 100ns。不填或者小于0表示永久有效。
                 "info": <string>  // 具体类型的json字符串，详见下
               },
               ...
              ]
    }

    其中
    info类型分baseprice、discountinfo、packageinfo三种

    baseprice 基本阶梯
    {
	  "id": <string>,   // ID，可以指定内部已经存在的价格，且以下内容省略。不填或者为空表示为以下内容为新价格。
	  "type": <string>, // 类型，待扩展。
	  "desc": <string>, // 相关描述	
	  "space": [        // 空间相关价格
	            {
			      "range": <NumberLong(0)>,   // 阶梯, GB
				  "price": <NumberLong(0)>    // 单价, 0.01分
			    },
			    ...
			   ],
	  "transfer_out": [      // 流量相关价格
	                   {
				         "range": <NumberLong(0)>, // GB
				         "price": <NumberLong(0)>  // 0.01分
				       },
				       ...
				      ],
	  "bandwidth": [        // 带宽相关价格
				    {
				      "range": <NumberLong(0)>,
				      "price": <NumberLong(0)>
				    },
				    ...
				   ],
      "api_get": [          // get请求数价格
			      {
			        "range": <NumberLong(0)>, // 1000次
			        "price": <NumberLong(0)>
			      },
			      ...
			     ],
	  "api_put": [          // put请求数价格
	              {
	                "range": <NumberLong(0)>, // 1000次
	                "price": <NumberLong(0)>
	              },
	              ...
	             ]	
    }

    discountinfo 折扣
	{
	  "id": <string>,        // ID，可以指定内部已经存在的折扣，且以下内容省略。不填或者为空表示为以下内容为新折扣。
	  "type": <string>,      // 类型，待扩展。
	  "desc": <string>,      // 相关描述
	  "money": <int64>,      // 折扣绝对值, 0.01分，正值表示扣除相应费用
	  "percent": <int>       // 折扣百分比, e.g. 50 表示50%
	}

    packageinfo 套餐
	{
	  "id": <string>,            // ID，可以指定内部已经存在的套餐，且以下内容省略。不填或者为空表示为以下内容为新套餐。
	  "type": <string>,          // 类型，可为normal（表示用户套餐类型），reward（表示奖励）
	  "desc": <string>,          // 相关描述
	  "money": <int64>,          // 套餐金额, 0.01分
	  "space": <int64>,          // 空间额度, GB
	  "transfer_out": <int64>,   // 流量额度, GB
	  "bandwidth": <int64>,      // 带宽额度
	  "api_get": <int64>,        // get请求数额度, 1000次
	  "api_put": <int64>         // put请求数额度, 1000次
	}

### 2. &nbsp; 查询 &nbsp;&nbsp;&nbsp;&nbsp; **/get**
* 权限
  * admin
  * 用户本身
* 参数
  * uid --- 用户id
  * time --- 查询时间,100ns。给出此时间点的价格信息。不填取全部价格内容。
 
***-- Request --***

	GET /get_bills?uid=<uid>&time=<int64> HTTP/1.1
	Authorization: <Authorization>
		
***-- Response --***

正常结果

	HTTP/1.1 200 OK
	Content-Type: application/json
	Content-Length: xxx
	
	同上request

