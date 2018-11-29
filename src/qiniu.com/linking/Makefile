QCHECKSTYLE=go run $(GOPATH)/src/vendor/github.com/qiniu/checkstyle/gocheckstyle/gocheckstyle.go -config=.qcodestyle
all:
	go install ./...
	gofmt -w -s .
	@$(QCHECKSTYLE) .
test:
	go test ./...
style:
	@$(QCHECKSTYLE)
