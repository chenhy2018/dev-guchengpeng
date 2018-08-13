if [ $# != 1 ];then
    echo "USAGE:$0 arch"
    echo "e.g:$0 mstar"
    exit 1;
fi

source env.sh

export ARCH=$1

export WITH_P2P="OFF"
export WITH_OPENSSL="OFF"

prefix=
if [ "$1" = "mstar" ];then
    export CC=arm-linux-gnueabihf-gcc
    export CXX=arm-linux-gnueabihf-g++
    export HOST=arm-linux-gnueabihf
    export LIBPREFIX=arm-unknown-linux-gnueabihf
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

if ! ./merge-lib.sh $1; then echo "merge lib failed"; exit 1; fi

cd link/libsdk/
./test.sh $1