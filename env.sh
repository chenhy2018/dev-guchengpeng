export QBOXROOT="${PWD}/third_party"
source "$QBOXROOT/base/env.sh"
source "$QBOXROOT/base/mockacc/env.sh"
source "$QBOXROOT/apigate/env.sh"
export GOPATH=$QBOXROOT/vendor/:${PWD}/link/vod:$GOPATH
export GOPATH=$HOME/gopath:$GOPATH
export PATH=$HOME/gopath/bin:$PATH
export GIN_MODE=release
