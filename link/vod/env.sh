export QBOXROOT="$(dirname "$(dirname "$(pwd)")")/third_party"
export GOPATH=$QBOXROOT/vendor/:`pwd`:$GOPATH
source "$QBOXROOT/base/env.sh"
source "$QBOXROOT/base/mockacc/env.sh"
