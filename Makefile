QCHECKSTYLE=go run $(QBOXROOT)/base/qiniu/src/github.com/qiniu/checkstyle/gocheckstyle/gocheckstyle.go -config=.qcodestyle
all:
	cd link/vod/src; go install -v ./...
clean:
	cd link/vod/src; go clean -i ./...

test:
	cd link/vod/src; go test ./...
style:
	@$(QCHECKSTYLE) link/vod/src
