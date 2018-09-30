make clean
if [ "$WITH_OPENSSL" = "ON" ];then
    cd ./third_party/openssl-1.1.0h
    make clean
    cd ../../
fi

if [ "$WITH_PJSIP" = "ON" ];then
    cd ./third_party/pjproject-2.7.2
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
fi

if [ "$WITH_FFMPEG" = "ON" ];then
    cd ./third_party/ffmpeg-4.0
    make clean
    cd ../../
fi

if [ "$ARCH" = "x86" ];then
    cd ../curl-7.61.0
    make clean
    cd ../../
fi

if [ -f CMakeCache.txt ];then
    rm -rvf CMakeCache.txt
fi

if [ -d CMakeFiles ];then
    rm -rvf  CMakeFiles
fi

if [ -d output ]; then
    rm -rvf output
fi

