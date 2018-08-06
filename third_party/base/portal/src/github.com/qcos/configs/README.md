
# Config server (configs)

程序使用 qcos 发布的时候，我们需要将配置文件打包进镜像

对跑在 qcos 上的服务来说，有以下问题

- 在 ci 服务器上保存了配置文件供打包镜像使用（开发人员无法更新配置）
- 配置文件打包进镜像 push 到公用的 registry 服务器（带来安全问题）
- 程序回滚时带着配置回滚，多数关键配置回滚后无法工作（比如已经更改过的链接地址，旧的无法工作）

所以我们需要有个服务专门托管这些配置文件，在程序运行时加载配置文件

## 配置文件保存结构

通过程序发布流程将配置文件发布到 config server，配置文件按照目录保存

```
.config.json
<ServiceName>/
	<VersionNumber>/
		<ConfigFiles>
		...
portal-qcos/
	1/
		app.conf
	199/
		app.conf
portal-kodo/
	1/
		app.conf
	88/
		app.conf
	99/
		extra/
			extra.conf
```

- `ServiceName` 服务名称
- `.config` 保存了针对这个服务配置文件本身的设置
- `VersionNumber` 下面保存增量更新的配置文件内容

## 工作方式

### important

configs 实际上只是一个带验证的文件服务器

- 客户端使用 ak/sk 签名 api 请求返回配置
- 如果指定版本号的配置文件不存在的话，会获取最近（存在的）版本号的配置文件

线上实践：

- 实际使用中，与之前程序兼容的配置，只需要在最新版本里修改即可
- 程序不变，只做配置的修改，只要重启程序

### 1. 配置

.config.json

```
{
	// auth 里每一组 ak/sk 对应一个唯一的服务
	"auth": {
		<AccessKey:string>: {
			"service": "portal-qcos",
			"secret": <SecretKey:string>
		},
		<AccessKey:string>: {
			"service": "portal-kodo",
			"secret": <SecretKey:string>
		}
	}
}
```

### 2. Api

#### a. 批量获取配置文件内容

例子为请求 portal-kodo 的配置文件

请求包:

```
POST /v1/config/batch
Content-Type: application/json
Authorization: Qiniu <MacToken>

{
	files: [
		[
			"name": "app.conf",
			"version": 88
		],
		[
			"name": "extra/extra.conf",
			"version": 100
		]
	]
}
```

返回包:

```
200 OK
{
	files: [
		[
			"code": 200,
			"message": "",
			"data": {
				"name": "app.conf",
				"version": 1,
				"content": "xxxxxx"
			}
		],
		[
			"code": 200,
			"message": "",
			"data": {
				"name": "extra/extra.conf",
				"version": 99,
				"content": "xxxxxx"
			}
		]
	]
}
```

### 3. 验证

- client 端使用 Qiniu Mac 签名
- server 端实现了 WbrpcGetb 所以可以使用 apigate 验证

## Future

- 模板功能？
