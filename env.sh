export QBOXROOT="${PWD}/third_party"
export GOPATH=$GOPATH:$QBOXROOT/vendor/:${PWD}/link/vod
source "$QBOXROOT/base/env.sh"
source "$QBOXROOT/base/mockacc/env.sh"
source "$QBOXROOT/apigate/env.sh"
export GIN_MODE=release
