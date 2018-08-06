qiniupkg.com/x/errors.v8
===============

整体来说这个版本放弃了 qiniupkg.com/x/errors.v7 的实现，改用 [github.com/juju/errors](https://github.com/juju/errors)。

在 github.com/juju/errors 基础上，我们做了如下调整：

* IsBadRequest, IsAlreadyExists, IsNotFound 的判断条件兼容了 qiniupkg.com/x/errors.v7 的惯例，基于 syscall.EINVAL, syscall.EEXIST, syscall.ENOENT。
* 对 qiniupkg.com/x/errors.v7 生成的错误进行了兼容处理。
