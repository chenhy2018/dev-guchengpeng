# base (qiniu common packages)

[![Build Status](https://magnum.travis-ci.com/qbox/base.png?token=8Frioiaq36bhxZs3jYgn&branch=develop)](https://magnum.travis-ci.com/qbox/base)

[![Qiniu Logo](http://open.qiniudn.com/logo.png)](http://www.qiniu.com/)

# 七牛开发指南

重要提示：

* 这里可以了解最新的 [七牛开发指南](https://github.com/qbox/base/blob/develop/docs/README.md)。

# 编译方式

假设代码的根是 ~/qbox（也就是本库在 ~/qbox/base 目录下），则编译流程为：

1. 安装 go，我们要求从这里下载代码安装：https://github.com/qiniu/go/tags
2. 打开 .profile，加入：

    export QBOXROOT=~/qbox

    source $QBOXROOT/base/env.sh
    
    source $QBOXROOT/base/mockacc/env.sh
    

3. 保存 .profile ，并 source 之
4. cd $QBOXROOT/base; make


# 目录说明

* qiniu: 公共基础代码，可能会开源，所以名字空间按开源来处理。
* com: 公共基础代码。如果在 qiniu 里面有类似组件，则应优选 qiniu 里面的。这里没有删除是因为有的代码还在引用它，会逐步进行更新。
* biz: 公共基础代码（业务相关）。biz 可能和 com 或 qiniu 有类似代码么？不可能。
* mockacc: 仿真(Mock)的账号认证系统，方便开发人员调试而提供。

# 模块废弃

* qiniu: 可能会结合 github.com/golang/glog 和 github.com/qiniu/{log.v1,xlog.v1}。
* biz: API Server 的推荐方式做了大调整，推荐基于 [apigate](https://github.com/qbox/base/tree/develop/qiniu/src/github.com/qiniu/apigate.v1)
* com: 模块 qbox.us/dyn 将迁移到 github.com/qiniu/dyn
* com: 模块 qbox.us/{mgo, mgo2, mgo3} 将逐步统一到 github.com/qiniu/db/mgoutil.v3
* com: 模块 qbox.us/{encoding, encoding1.2}/, github.com/qiniu/encoding.v2/ 等将逐步统一到 github.com/qiniu/encoding/ 下
* com: 模块 code.google.com/p/* 已经 golang team 挪到 golang.org/x/* 下面。比如 code.google.com/p/go.image 已经挪到 [golang.org/x/image](https://github.com/golang/image)，我们需要 follow 该调整。

