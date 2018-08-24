apigate example
===================

# 测试环境

第一步：启动 apigate_example

```bash
apigate_example -f ./apigate_example.conf
```

第二步：启动 mockapigate

```bash
qiniumockgate -f ./qiniugate.conf #正式环境下应该用 qiniugate
```

第三步：测试 apigate 的可访问性

```
./test.qtf
```

这个测试程序为mock环境而建，如果要在正式环境，需要将

```bash
auth qinutest `authstub -uid 1 -utype 4`
```

这一句改为：

```bash
auth qinutest `qbox <AK> <SK>` #其中 AK/SK 需要调整为测试帐号的 AK/SK
```

