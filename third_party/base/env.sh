if [ "$QBOXROOT" = "" ]; then
	QBOXROOT=$(cd ../; pwd)
	export QBOXROOT
fi

if [ "$TRAVIS_BUILD_DIR" != "" ]; then
	export GOPATH=
fi

export GOPATH=$GOPATH:$QBOXROOT/base/qiniu:$QBOXROOT/base/account-api:$QBOXROOT/base/docs:$QBOXROOT/base/com:$QBOXROOT/base/biz:$QBOXROOT/base/portal:$QBOXROOT/base/cgo
export PATH=$PATH:$QBOXROOT/base/qiniu/bin:$QBOXROOT/base/biz/bin:$QBOXROOT/base/cgo/bin:$QBOXROOT/base/com/bin:$QBOXROOT/base/portal/bin
export INCLUDE_QBOX_BASE=1

if [[ $GOPATH == :* ]]; then export GOPATH=`echo $GOPATH | sed "s/://"`; fi
