GOVERSIONGT15 := $(shell go version | grep version >/dev/null && echo 1 || expr `go version | cut -f3 -d ' '` \>= "go1.5")
ifeq "$(GOVERSIONGT15)" "1"
        LDFLAGS := "-X 'github.com/qiniu/version.version=${BUILD_NUMBER} ${BUILD_ID} ${BUILD_URL}' -X 'github.com/qiniu/version.pkgName=${PACKAGE_NAME}'"
else
        LDFLAGS := "-X github.com/qiniu/version.version '${BUILD_NUMBER} ${BUILD_ID} ${BUILD_URL}' -X github.com/qiniu/version.pkgName '${PACKAGE_NAME}'"
endif

QCHECKSTYLE=go run $(QBOXROOT)/base/qiniu/src/github.com/qiniu/checkstyle/gocheckstyle/gocheckstyle.go -config=.qcodestyle

all:
	CGO_ENABLED=0 go install -ldflags ${LDFLAGS} -v ./...
	mv ./bin/qiniugate ./bin/linking_apigate
	mv ./bin/qiniumockgate ./bin/linking_mockgate

rebuild:
	CGO_ENABLED=0 go install -ldflags ${LDFLAGS} -a -v ./...

install: all
	@echo

test:
	go test ./...

testaone:
       
	 curl https://aone.qiniu.io/api/coverage/collect?token=A59B8029-E9A1-4FF3-B72B-3657CEAC64D4 | bash

testv:
	go test -v ./...

clean:
	@echo

style:
	@$(QCHECKSTYLE) src

