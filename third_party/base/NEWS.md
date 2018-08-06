* 2014-05-07 [#8293] cc/mime增加mimeType到文件后缀的映射
* 2014-04-22 [#4595] 对gif文件增加frameNumber的检测
* 2013-10-28 [#4322] 增加 admin_api/account: 新的第三方相关 API
* 2013-10-15 [#4662] 移除 account:/user/send_invitation
* 2013-10-14 [#4455] 增加函数 "qbox.us/mgo2".IsSessionClosed, CopyDatabase, CloseDatabase, CopyCollection, CloseCollection
* 2013-10-14 [#4455] labix.org/v2/mgo 更新到 r2013.09.04
## ibuck (pre)

2013-10-06 [issue #765](https://github.com/qbox/base/pull/765)

* Mime type by ext: support ipa, webp, m4a, etc.
* new package: qbox.us/http/account
* new package: github.com/qiniu/http/servestk
* github.com/qiniu/http/flag: support new attribute 'has'
* github.com/qiniu/http/{flag, formutil} bugfix: Zero value if no param found


## update mgo

2013-09-30 [issue #757](https://github.com/qbox/base/pull/757)

* 2013-09-30 [#4456] 修复 Query.Apply 不支持 mongodb 1.8.3
* 2013-09-25 [#4365] 增加 "qbox.us/mgo2"
