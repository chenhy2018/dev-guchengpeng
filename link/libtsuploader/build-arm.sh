
if [ $# == 1 ];then
	if [ $1 == 'distclean' ];then
		rm -rvf CMakeFiles CMakeCache.txt Makefile cmake_install.cmake
		exit
	fi
fi

cmake . -DCMAKE_TOOLCHAIN_FILE=./toolchain.cmake
make
cp -rvf testupload /home/share
