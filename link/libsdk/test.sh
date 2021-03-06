if [ $# != 1 ];then
    echo "USAGE:$0 arch"
    echo "e.g:$0 mstar"
    exit 1;
fi

prefix=
if [ "$1" = "mstar" ];then
    export CC=arm-linux-gnueabihf-gcc
    export CXX=arm-linux-gnueabihfg++
    export CROSS_COMPILE=arm-linux-gnueabihf-
elif [ "$1" = "a12" ];then
    export CC=gcc
    export CXX=g++
elif [ "$1" = "x86" ];then
    export CC=gcc
    export CXX=g++
fi

export ARCH=$1

echo "******build test*******"
cd test
make clean
make
echo "*****build test_calling*****"
cd ..
cd test_calling
make clean
make
echo "*****build test_receivedcall****"
cd ..
cd test_receivedcall
make clean
make

echo "*****build test_mqtt****"
cd ..
cd test_mqtt
make clean
make

if [ "${WITH_P2P}" = "ON" ]; then
echo "*****build test_demo****"
cd ..
cd test_demo
make clean
make

if [ "$1" = "mstar" ];then
cd ..
cd sdk_demo
make clean
make
fi

fi
