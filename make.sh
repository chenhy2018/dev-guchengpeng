if [ $# != 1 ];then
    echo "USAGE:$0 arch"
    echo "e.g:$0 mstar"
    exit 1;
fi
export ARCH=$1

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
    export CROSS_COMPILE=
    export LIBPREFIX=x86_64-unknown-linux-gnu
fi

cmake .
make

./merge-lib.sh $1

cd link/libsdk/
./test.sh $1
