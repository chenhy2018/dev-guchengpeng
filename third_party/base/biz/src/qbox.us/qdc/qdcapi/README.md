Qiniu Disk Cache API (SDK)
========================

# 版本

## lbdc

* 后端基于 lbd，但仅仅保留 GetLocal/PutLocal 这两个 cache 协议
* qboxlbd 本身的代码已经冻结，如需修改，请改 qboxdc 的最新版本

## v1/lbdc

* 老版本的 lbd 的 client

## v2/lbdc

* 后端基于 lbd，但是假设 lbd 仅支持 GetLocal/PutLocal 这两个 cache 协议，在此基础上实现 v1/lbdc 兼容的所有方法

