QCHECKSTYLE=go run $(QBOXROOT)/base/qiniu/src/github.com/qiniu/checkstyle/gocheckstyle/gocheckstyle.go -config=.qcodestyle
all:
	cd src; go install ./...
	go fmt ./...
	@$(QCHECKSTYLE) src
	mv ./bin/main ./bin/linking_vod
clean:
	cd src; go clean -i ./...

test:
	cd src/qiniu.com/controllers; go test
	cd src/qiniu.com/models; go test
	cd src/qiniu.com/m3u8; go test
	cd src/qiniu.com/db; go test
	cd src/qiniu.com/auth; go test
style:
	@$(QCHECKSTYLE) src
