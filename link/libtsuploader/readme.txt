c-sdk 存储提供的库，依赖于curl和openssl

ffmpeg 裁剪:
./configure --enable-small --disable-programs --disable-doc --disable-avdevice --disable-swresample --disable-swscale --disable-postproc --disable-avfilter --disable-network --disable-dct --disable-dwt --disable-error-resilience --disable-lsp --disable-lzo --disable-mdct --disable-rdft --disable-fft --disable-faan --disable-pixelutils --disable-everything --enable-muxer=mpegts

libcurl 裁剪：
./configure --enable-shared=no  --without-libidn2 --disable-libtool-lock --enable-http --disable-ftp --disable-file --disable-ldap --disable-ldaps --disable-rtsp --disable-proxy --disable-dict --disable-telnet --disable-tftp --disable-pop3 --disable-imap --disable-smb --disable-smtp --disable-gopher --disable-manual --disable-libcurl-option --enable-ipv6 --disable-largefile --disable-sspi --disable-ntlm-wb --disable-unix-sockets --disable-cookies --disable-crypto-auth --disable-tls-srp

关于openssl:
其中opensll版本不能大于1.1.0。 linking中的openssl是1.1.0h的，导致编译通不过
openssl的裁剪比较麻烦，按照齐斌的意思openssl 链接摄像头的动态库，这个上传的sdk就不需要编译openssl了


多线程改为单线程上传:
	目前本来就是串行上传，只是不等待前一个的上传结果，但是有极小可能前一个上传成功的回调慢于下一个
	解决办法：
	1. 还是不使用单线程，等待调用结果时候会卡住
	2. 设置超时时间，因为切片时间是5s左右，设置一个3s的超时，这个在curl通过lowspeed和lowspeedtime设置

2. c-sdk 改为client上传
	1. 保留了ak sk的借口
	2. 增加了传递token的接口

3. key格式
	

4. segment的起始条件

5. 音视频格式的指定

1. upload 失败立即结束上传线程，开启下个上传线程

名字可选: 回调函数,应该不用了

1. 从token中获取expire加到文件名中. 完成
2. openssl 开关. 完成
3. http获取token，解析expre. 完成
4. 时间应该是第一个ts切片的时间，即现场要在第一个push后在启动或者说在开始工作
5. 从服务器获取时间，然后对应到本地时间(机器启动时间或者程序启动时间)
6. 新片段开始, 时间间隔，之前的时间和当前的时间差
