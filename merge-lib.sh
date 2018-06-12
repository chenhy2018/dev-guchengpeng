#!/bin/sh
# merge-lib.sh
# pjsip will create a lot of xxx.a files, but we need only one,  so we need ar -x all the .a file
# and merge all the .o file to a only one .a file. the name is libqnsip.a

cnt=0
pjsip_path=./third_party/pjproject-2.7.2

rm -rvf ./libs/ori
rm -rvf ./libs/objs
rm -rvf ./libs/tmp

mkdir -p ./libs/ori
mkdir -p ./libs/objs

# 1. copy all the .a file from pjproject to libs/ori directory
find $pjsip_path -name "*.a" -exec cp -v {} ./libs/ori \;

cd libs/ori
for f in ./*
do
    if test -f $f
    then
        mkdir -p ../tmp/$f
        cp $f ../tmp/$f
        cd ../tmp/$f/
        # 2. release all the .o to tmp directory
        ar x $f
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
        cd ../../ori
    fi
done

cd ../objs
# 4. strip all the .o files
arm-linux-gnueabihf-strip *.o --strip-unneeded
# 5. gen new library
arm-linux-gnueabihf-ar r libqnsip.a *.o
cp libqnsip.a ../
