#!/bin/sh

if  [ ! -n "$ARCH" ] ;then
	echo !!!ARCH not specify
	exit 1
fi

if [ -d link-sdk ];then
	if [ $ARCH = "hi" ];then
		cp -rvf link/libtsuploader/libtsuploader.a link-sdk/lib/arm/
		cp -rvf link/libtsuploader/c-sdk/libqiniu.a link-sdk/lib/arm/
	elif [ $ARCH = "80386" ];then
		cp -rvf link/libtsuploader/libtsuploader.a link-sdk/lib/x86/
		cp -rvf link/libtsuploader/c-sdk/libqiniu.a link-sdk/lib/x86/
	fi
	cp -rvf link/libtsuploader/*.h link-sdk/include/
fi
