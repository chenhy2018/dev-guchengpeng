ps -e | grep demoserver | sed 's|\([ 0-9]*\).*|\1|g' | xargs -n 1 --no-run-if-empty kill

