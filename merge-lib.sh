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

prefix=
if [ "$1" = "mstar" ];then
    prefix=arm-linux-gnueabihf-
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

# 1. copy all the .a file from pjproject to libs/ori directory
cp -rf third_party/pjproject-2.7.2/prefix/lib/* ${OUTPUT}/ori/$1
rm -rf ${OUTPUT}/ori/$1/libpjsua2-x86_64-unknown-linux-gnu.a


cp -rvf link/libsip/*.a ${OUTPUT}/ori/$1
cp -rvf link/librtp/*.a ${OUTPUT}/ori/$1
cp -rvf link/libmqtt/*.a ${OUTPUT}/ori/$1
cp -rvf link/libsdk/*.a ${OUTPUT}/ori/$1

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
