lbd.Get流程：

// 先在dc缓存中找。
if dc.Get(key) 存在 { return }

// 在dc缓冲中找不到需要到bd中找。
for i := 0; i < 3; i++ {
    if bds[i] == 0xffff { break }
    bdClients[bds[i]].Get(key): 
       0) 如果正常，return
       1) 如果 key 找不到，break
       2) 如果 conn 不在，continue
}

// 在bd中找不到需要到up节点找，因为数据可能还在up的bd缓存里，还没迁移到bd。
if bdcacheClient.Get(key) 存在 { return }

// 在up节点没找到需要到bd中再找一次，因为去up节点找的时候数据可能恰好被迁移到bd。
for i := 0; i < 3; i++ {
    if bds[i] == 0xffff { break }
    bdClients[bds[i]].Get(key): 
       0) 如果正常，return
       1) 如果 key 找不到，break
       2) 如果 conn 不在，continue
}
