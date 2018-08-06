# GAEA API

本目录包含了 GAEA (account.qiniu.com) 提供的 SDK.

GAEA 提供用户信息相关的一系列操作，区别于底层的 acc 服务，GAEA 主要提供上层的用户信息、敏感操作两步认证、OpLog 读写等功能.

## 接口说明

### 获取用户信息 (DeveloperService.Info)

获取用户的详细信息，即原 developers 表相关字段.

[接口](developer.go#L74)

### [Admin] 通过 UID 获取用户信息 (AdminDeveloperService.InfoByUid)

根据 UID 获取用户的详细信息，即原 developers 表相关字段.

[接口](admin_developer.go#L28)

### [Admin] 通过 Email 获取用户信息 (AdminDeveloperService.InfoByEmail)

根据 Email 获取用户的详细信息，即原 developers 表相关字段.

[接口](admin_developer.go#L48)

### 获取两步认证状态 (VerificationService.Check)

检查用户在 GAEA 上进行两步认证后的认证状态.

如返回 true, 则可允许进行敏感操作.

[接口](verification.go#L25)

### 销毁两步认证状态 (VerificationService.Consume)

销毁用户在 GAEA 上进行两步认证后的认证状态.

销毁后，将重置为未认证状态.

[接口](verification.go#L40)

### 创建 OpLog (OpLogService.Create)

创建一条 OpLog.

[接口](oplog.go#L40)

### 查询 OpLog (OpLogService.Query)

根据用户的 UID 查询 OpLog.

[接口](oplog.go#L86)
