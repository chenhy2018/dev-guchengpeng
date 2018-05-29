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

TODO 已解决
1. MediaConfig不在使用union，合并到一起, 已完成
2. transport是否要放到stream里面去。 不这样做，但是现在需要确认streamTracks和transportIce的下标一致
   确定一致，在addtrack的时候会初始化对应下标的transport
3. ice协商时候把每个mline拆分单独协商. 这个是代码问题，现在成功了
   pjmedia_transport_media_start media_index 一直传的0
   addtrack的时候下标搞错了，导致先添加的下标位1，后添加的下标为0

TODO 未解决
4. sdp mline支持多种格式，需要知道那种格式协商成功了 
    pjmedia_sdp_neg 可以获取这个消息。目前能获取了，
    但是没有通知调用者, 如何通知？？回调函数，PeerConnection加了一个IceNegInfo的成员，这个成员的附值在checkAndNeg里面
5. 发送数据的接受到数据和rtcp的回调函数，参考siprtp. 正在做
   添加上了，还没有测试，并且还需要缓冲并拼接成一个完整的帧
6. 回调函数，目前是sleep等待超时ice的状态. 对于用户没有回调，即同步的. 确定是否这样做？
7. 时间戳维护, rtp时间戳溢出问题？
8. StartNegotiation移动到checkAndNeg最后面去，即offer和answer都获取到就自动协商了. 可能最后做，这样做了不好通过文件sdp手动测试了
