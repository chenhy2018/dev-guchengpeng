export QBOXROOT="${PWD}/third_party"
export GOPATH=$QBOXROOT/vendor/:${PWD}/link/vod:${PWD}/link/manager:$GOPATH
source "$QBOXROOT/base/env.sh"
source "$QBOXROOT/base/mockacc/env.sh"
