七牛开发指南
====

# 架构委员会 & 开发规范

七牛成立了一个[架构委员会](https://github.com/qbox/base/blob/develop/docs/arch-committee.md)，并赋予该架构委员会以技术上至高无上的权力，包括：

* 审查一个业务的架构是否合理，技术方案是否存在风险。
* 审查一个业务的开发过程是否规范，文档化是否达到要求。
* 就任何潜在技术风险，可要求产品部门限时整改。比如：测试覆盖率过低、核心模块没有进行压力测试的评估。

每个产品部门，文档相关要求如下：

* 对外的可公开给客户的文档，请统一放到 https://github.com/qbox/product 相应产品的目录下。
* 内部架构设计，或者内部服务的API协议文档，请放到模块所在的repo。可以统一放到 repo 根目录 docs 目录下，也可以放在模块本身的 README.md 或 docs 中。但是有一个硬性的要求，就是 repo 的 README.md 需要有这些模块相关文档的链接。

在任何时刻，均需要保证文档（特别是API协议文档），与实现代码的一致性。这一点应该作为团队 CodeReview 的核心检查点之一。

更详细的开发规范，请参阅[七牛开发规范](https://github.com/qbox/base/blob/develop/docs/devspec.md)。

# 怎么做复杂系统的架构？

第一步：根据业务分析，就 API 协议在团队内获得一致认知，确认定义的 API 已经满足业务需求。成果体现为`API 协议文档`。具体可参阅[API 协议文档范例](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/foo.v1/API.md)。

第二步：模块划分，以及每个模块的概要设计。成果体现为`概要设计文档`，内容包括：数据库的描述文档，以及重要业务场景的流程描述。具体可参阅[概要设计文档范例](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/foo.v1/DESIGN.md)。

第三步：每个模块的负责人，完成模块的 `详细设计文档`，内容包括：每个组件（类）的规格（重要公开方法的原型定义）。如果模块并不复杂，这一步多数情况下会跳过。但是非常建议模块负责人习惯性进行自我审察，看看模块设计是否有进一步的优化空间。另外，可以考虑基于文档自动生成工具（doxygen之类）来抽取文档。

# 七牛新人最需要了解的基础库有哪些？

1、七牛的所有开源库（[qiniupkg.com/*](https://github.com/qbox/base/tree/develop/qiniu/src/qiniupkg.com)），具体如下：

* [qiniupkg.com/x](https://godoc.org/qiniupkg.com/x): 经常使用的一些常规基础组件，包括 log/xlog, rpc, mockhttp 等等。
* [qiniupkg.com/http](https://godoc.org/qiniupkg.com/http): 一些用于实现HTTP服务器的组件。这块大部分还没有开源。
* [qiniupkg.com/qiniutest](https://github.com/qiniu/httptest): 七牛的HTTP服务测试工具。
* [qiniupkg.com/api.v7](https://godoc.org/qiniupkg.com/api.v7): 七牛云服务的 Go SDK，通过这个包可了解七牛的业务。

2、一些未来有可能开源的库（[github.com/qiniu/*](https://github.com/qbox/base/tree/develop/qiniu/src/github.com/qiniu)），重点如下：

* [github.com/qiniu/http/*](https://github.com/qbox/base/tree/develop/qiniu/src/github.com/qiniu/http): 一些用于实现HTTP服务器的组件。

# 怎么写一个服务器？

我们通常会将服务器分拆为两个package，一个是library（编译成为.a文件），一个是application（编译出可执行文件）。

### restrpc 服务器框架

* [restrpc 服务器框架的使用样例](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/foo.v1/)
* [单元测试](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/foo.v1/foo_test.go)
* [可执行程序](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/foo.v1/qfoo/)

### 基于 MongoDB 做持久化

* [基于 MongoDB 做持久化的服务器](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/mongo.v1/)
* [单元测试](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/mongo.v1/mongo_test.go)
* [可执行程序](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/mongo.v1/qmongo/)

### 基于 MySQL 做持久化

* TODO

# 怎么测试你的服务器？

* 单元测试：推荐用 qiniu httptest 框架，详细参见 [单元测试样例](https://github.com/qbox/base/blob/develop/docs/src/qiniu.com/kodo/foo.v1/foo_test.go)。
* 集成测试：TODO

# 怎么接入七牛帐号系统？

TODO

# TODO

...

