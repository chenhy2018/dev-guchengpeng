#!/bin/sh
# merge-lib.sh
# pjsip will create a lot of xxx.a files, but we need only one,  so we need ar -x all the .a file
# and merge all the .o file to a only one .a file. the name is libqnsip.a

cnt=0
pjsip_path=third_party/pjproject-2.7.2

if [ $# != 1 ];then
    echo "USAGE:$0 arch"
    echo "e.g:$0 mstar"
    exit 1;
fi

LIBPREFIX=x86_64-unknown-linux-gnu
prefix=
if [ "$1" = "mstar" ];then
    prefix=arm-linux-gnueabihf-
    LIBPREFIX=arm-unknown-linux-gnueabihf
elif [ "$1" = "a12" ];then
    prefix=arm-linux-gnueabi-
fi
OUTPUT=output
target=./${OUTPUT}/lib/$1/libua.a

rm -rf ./${OUTPUT}/objs
rm -rf ./${OUTPUT}/tmp
rm -rf ./${OUTPUT}/ori/$1

mkdir -p ./${OUTPUT}/ori
mkdir -p ./${OUTPUT}/objs
mkdir -p ./${OUTPUT}/ori/$1
mkdir -p ./${OUTPUT}/lib/$1/
#read WITH_P2P
# 1. copy all the .a file from pjproject to libs/ori directory
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libpjsip*-${LIBPREFIX}.a ${OUTPUT}/ori/$1
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libpjmedia-${LIBPREFIX}.a ${OUTPUT}/ori/$1
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libpj-${LIBPREFIX}.a ${OUTPUT}/ori/$1
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libpjlib-util-${LIBPREFIX}.a ${OUTPUT}/ori/$1

if [ "${WITH_P2P}" = "ON" ]; then
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libpjnath-${LIBPREFIX}.a ${OUTPUT}/ori/$1
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libsrtp-${LIBPREFIX}.a ${OUTPUT}/ori/$1
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libpjmedia-audiodev-${LIBPREFIX}.a ${OUTPUT}/ori/$1
cp -rvf third_party/pjproject-2.7.2/prefix/$1/lib/libpjmedia-codec-${LIBPREFIX}.a ${OUTPUT}/ori/$1
fi

cp -rvf link/libsip/libsip-${LIBPREFIX}.a ${OUTPUT}/ori/$1

if [ "${WITH_P2P}" = "ON" ]; then 
cp -rvf link/librtp/librtp-${LIBPREFIX}.a ${OUTPUT}/ori/$1
fi

cp -rvf link/libmqtt/libmqtt-${LIBPREFIX}.a ${OUTPUT}/ori/$1
cp -rvf link/libsdk/libsdk-${LIBPREFIX}.a ${OUTPUT}/ori/$1

cd ${OUTPUT}/ori/$1
for f in ./*
do
    if test -f $f
    then
        mkdir -p ../../tmp/$f
        cp $f ../../tmp/$f
        cd ../../tmp/$f/
        # 2. release all the .o to tmp directory
#        echo $f
        ${prefix}ar x $f
        for bin in ./*.o
        do
            # 3. copy all the .o files to libs/objs, if found a file already exist, then rename the current file
            if [ -f "../../objs/$bin" ];then
                cnt=`expr $cnt + 1`
                echo [ $cnt ] found exist file ${f%%arm*}${bin##*/}
                cp $bin ../../objs/${f%%arm*}${bin##*/}
            else
                cp $bin ../../objs/
            fi
        done
        cd ../../ori/$1
    fi
done
cd ../../objs
# 4. strip all the .o files
#${prefix}strip *.o --strip-unneeded
# 5. gen new library 
echo "gen new lib"
${prefix}ar r ../../${target} *.o
cp ../../link/libsdk/sdk_interface.h ../../${OUTPUT}/lib/$1/sdk_interface.h
