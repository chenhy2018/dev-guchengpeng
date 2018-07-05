if [ $# != 1 ];then
    echo "USAGE:$0 arch"
    echo "e.g:$0 mstar"
    exit 1;
fi

prefix=
if [ "$1" = "mstar" ];then
    export CC=arm-linux-gnueabihf-gcc
    export CXX=arm-linux-gnueabihfg++
elif [ "$1" = "a12" ];then
    export CC=gcc
    export CXX=g++
elif [ "$1" = "x86" ];then
    export CC=gcc
    export CXX=g++
fi

export ARCH=$1

cp -rf ../../third_party/openssl-1.1.0h/prefix/lib ./lib
cd test
make clean
make
cd ..
cd test_demo
make clean
make
cd ..
cd test_calling
make clean
make
cd ..
cd test_receivedcall
make clean
make

cd ..

if [ "$1" = "mstar" ];then
cd sdk_demo
make clean
make
fi
