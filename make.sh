if [ $# != 1 ];then
    echo "USAGE:$0 arch"
    echo "e.g:$0 mstar"
    exit 1;
fi

source env.sh
export ARCH=$1
export WITH_P2P="ON"
export WITH_OPENSSL="OFF"
export WITH_FFMPEG="OFF"
export WITH_PJSIP="OFF"
export WITH_MOSQUITTO="OFF"

prefix=
if [ "$1" = "mstar" ];then
   export CC=arm-linux-gnueabihf-gcc
   export CXX=arm-linux-gnueabihf-g++
   export HOST=arm-linux-gnueabihf
   export LIBPREFIX=arm-unknown-linux-gnueabihf
elif [ "$1" == "hi" ];then
    export CC=arm-hisiv300-linux-uclibcgnueabi-gcc
    export CXX=arm-hisiv300-linux-uclibcgnueabi-g++
    export HOST=arm-hisiv300-linux-uclibcgnueabi
elif [ "$1" == "80386" ];then
     export CC=gcc
     export CXX=g++
     export CFLAGS=-m32
elif [ "$1" = "a12" ];then
    export CC=gcc
    export CXX=g++
elif [ "$1" = "x86" ];then
    export CC=gcc
    export CXX=g++
    export LIBPREFIX=x86_64-unknown-linux-gnu
elif [ "$1" = "clean" ];then
    ./make_clean.sh
    exit 1;
fi

if [ -f CMakeCache.txt ];then
    rm CMakeCache.txt
fi

cmake .
if ! make VERBOSE=1; then echo "build failed"; exit 1; fi

if [ "$WITH_PJSIP" = "ON" ];then
    if ! ./merge-lib.sh $1; then echo "merge lib failed"; exit 1; fi

    cd link/libsdk/
    ./test.sh $1
fi

if [ "$ARCH" = "mstar" ];then
	cp -rvf link/libtsuploader/ipc-testupload /home/share/ipc-testupload-dbg
	arm-linux-gnueabihf-strip link/libtsuploader/ipc-testupload
	cp -rvf link/libtsuploader/ipc-testupload /home/share/
fi

if [ "$ARCH" = "hi" ] || [ "$ARCH" = "80386" ];then
    ./make-link-sdk.sh
fi

