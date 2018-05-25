how to test:
1. 启动程序，输入1 或 2选择成为offer或者answer

如果是offer：
2. 程序自动在当前目录生成offer.sdp
3. 将文件copy到answer的当前目录
6. 输入ok, 读取answer的answer.sdp
7. 输入ok开始协商


如果如answer：
4. 输入ok 读取offer生成的offer.sdp。 并且在当前目录生成answer.sdp
5. 将answer.sdp copy到offer的当前目录
7. 输入ok开始协商

TODO
1. MediaConfig不在使用union，合并到一起, 已完成
2. transport是否要放到stream里面去。 不这样做，但是现在需要确认streamTracks和transportIce的下标一致
   确定一致，在addtrack的时候会初始化对应下标的transport
3. ice协商时候把每个mline拆分单独协商
4. 回调函数，目前是sleep等待超时ice的状态
