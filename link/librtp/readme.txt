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

4. 多路track会有多次回调，这个要屏蔽为一次. 已完成。
	PeerConnection加了mutex和nNegFail nNegSuccess记录次数来控制回调次数
5. sdp mline支持多种格式，需要知道那种格式协商成功了: 已完成，待了解的更清楚
    pjmedia_sdp_neg 可以获取这个消息。目前能获取了
    但是没有通知调用者, 如何通知: 回调函数，PeerConnection加了一个IceNegInfo的成员，这个成员的附值在checkAndNeg里面
    可能的问题：作为answerer，active_remote 得到的sdp似乎不正确, active_local 和预期一致
6. 发送数据的接受到数据和rtcp的回调函数, 已完成
    pcmu能正常接收数据了

7. 发送h264的时候rtp marker字段设置，packetizer并不会设置这个字段 :已解决
     ffmpeg推流，wireshark抓包分析，发现marker位不能单独判断帧的起始， 并且奇怪的时候stap有时候设置marker有时候不设置
	 如果是stap-a类型，就是完整的一帧， 如果是fu-a类型，就要结合fu header和marker位来判断了
     所以在发送的时候也这样设置.
     最后方法是没有使用marker bit，通过fu-a类型判断的

8. 接收数据的回调函数分别处理音频和视频
     1) 音视频分别处理
     2) 根据samplerate或者clockrate还原时间戳
     目前pcmu传输和接收没有问题, 时间戳未还原
  h264_packetizer是可以同时pack和unpack的，但是h264来看，packetizer不会维护完整一帧的数据, 封装了下packitizer，来维护h264的数据

9. rtp丢包了怎么办。可能是通过rtp的序列号来判断是否丢包
   pjmedia h264 packetizer接口说明，貌似丢包了以null去调用，会更新内部状态， 每丢一个包调用一次吗？,目前做法，丢包会null调用一次
   自己写了jitterbuffer. pjsip的pjmedia_jbuf，测试了下，无论adpative参数怎么调整，push的比第一个frame seq大的frame都会被discard，所以没有

10. rtp 序列号restart
	接收使用int32_t类型序列号, 使用了一个非常简单的办法
	if (_nFrameSeq < _pJbuf->nMaxBufferCount+1 && _pJbuf->nLastRecvRtpSeq > 65000) {
                _nFrameSeq += 65536;
        }
	上一次的nLastRecvRtpSeq大于65000了，然后又出现了现在的_nFrameSeq小于nMaxBufferCount+1，则认为出现了翻转，
	现在的nFrameSeq则加上65536

	pop时候发现nLastRecvRtpSeq 变小了，重新排序一下队列？
        heap_rebuild 条件，如果65535这个包丢了怎么办？ 大于65535的序列号都&0x0000FFFF，这样就可以第一个翻转的包就能探测到
TODO 未解决
	如果之前判定为丢弃的包又接收到了，怎么判断?
 else if (_pJbuf->nLastRecvRtpSeq != 0 &&
                   (_nFrameSeq < _pJbuf->nLastRecvRtpSeq || (_nFrameSeq - _pJbuf->nLastRecvRtpSeq) > 65000)) {
                *_pDiscarded = 1;
                return;
        }


//这liang个问题待更严格的测试
11.  时间戳维护, rtp时间戳溢出问题？
	uint32_t类型的时间错，在90000hz时钟频率情况下，大概13h15m就会溢出
12. 音频的rtp marker位是否需要设置?

未解决：
13 sdp协商一边成功一边失败的问题
    加日志，失败时候的sdp，codec=0的那里

14 sdk测试程序发送视频coredump问题

15 sdp pj_trans.c  assert失败问题

16 setRemoteDescripton 返回失败-2

stun会挂，sip 状态更新



已明确，待选择做法:
1. 回调函数，目前是sleep等待超时ice的状态. 对于用户没有回调，即同步的. 确定是否这样做？
   做成异步
	1. 在createoffer/answer的时候在initTransportIce， 而不是在addVideo/AudioTrack的时候
        2. onIceComplete2 PJ_ICE_STRANS_OP_INIT根据次数(track数)来决定是否往sdp里面添加candidate
        
        
2. StartNegotiation移动到checkAndNeg最后面去，即offer和answer都获取到就自动协商了. 可能最后做，这样做了不好通过文件sdp手动测试了


