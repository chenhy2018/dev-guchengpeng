make clean
cd ./third_party/openssl-1.1.0h
make clean
cd ../pjproject-2.7.2
make clean
cd ../ffmpeg-4.0
make clean
cd ../curl-7.61.0
make clean
cd ../../link/libsip/
make clean
cd ../librtp/
make clean
cd ../libice/
make clean
cd ../libmqtt/
make clean
cd ../stun/
make clean
cd ../libsdk/
make clean
cd ../../
if [ -f CMakeCache.txt ];then
rm -rf CMakeCache.txt
fi

if [ -d CMakeFiles ];then
rm -rf  CMakeFiles
fi

if [ -d output ]; then
rm -rf output
fi
