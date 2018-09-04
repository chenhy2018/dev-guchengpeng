export QBOXROOT="${PWD}/third_party"
source "$QBOXROOT/base/env.sh"
source "$QBOXROOT/base/mockacc/env.sh"
source "$QBOXROOT/apigate/env.sh"
export GOPATH=$GOPATH:$QBOXROOT/vendor/:${PWD}/link/vod
export GIN_MODE=release
