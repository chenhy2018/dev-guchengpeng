if [ "$QBOXROOT" = "" ]; then
  QBOXROOT=$(cd ../; pwd)
  export QBOXROOT
fi

GOPATH=`pwd`/com:`pwd`/biz:`pwd`/qiniu:$QBOXROOT/account/account
export GOPATH
