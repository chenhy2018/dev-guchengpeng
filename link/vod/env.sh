export QBOXROOT="$(dirname "$(dirname "$(pwd)")")/third_party"
export GOPATH=$GOPATH:$QBOXROOT/vendor/:`pwd`
source "$QBOXROOT/base/env.sh"
source "$QBOXROOT/base/mockacc/env.sh"
