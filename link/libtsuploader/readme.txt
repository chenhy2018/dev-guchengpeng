c-sdk 存储提供的库，依赖于curl和openssl

ffmpeg 裁剪:
./configure --enable-small --disable-programs --disable-doc --disable-avdevice --disable-swresample --disable-swscale --disable-postproc --disable-avfilter --disable-network --disable-dct --disable-dwt --disable-error-resilience --disable-lsp --disable-lzo --disable-mdct --disable-rdft --disable-fft --disable-faan --disable-pixelutils --disable-everything --enable-muxer=mpegts

libcurl 裁剪：
./configure --enable-shared=no  --without-libidn2 --disable-libtool-lock --enable-http --disable-ftp --disable-file --disable-ldap --disable-ldaps --disable-rtsp --disable-proxy --disable-dict --disable-telnet --disable-tftp --disable-pop3 --disable-imap --disable-smb --disable-smtp --disable-gopher --disable-manual --disable-libcurl-option --enable-ipv6 --disable-largefile --disable-sspi --disable-ntlm-wb --disable-unix-sockets --disable-cookies --disable-crypto-auth --disable-tls-srp

关于openssl:
其中opensll版本不能大于1.1.0。 linking中的openssl是1.1.0h的，导致编译通不过
openssl的裁剪比较麻烦，按照齐斌的意思openssl 链接摄像头的动态库，这个上传的sdk就不需要编译openssl了



1. 上传设置超时时间, 主要作用是不能让后一个切片先上传上去
2. 从关键帧开始上传，这里还没有判断
3. c-sdk 改为client上传, 以及setaccesskey这些全部不用，callbackurl这些都是写死的应该
4. key格式
