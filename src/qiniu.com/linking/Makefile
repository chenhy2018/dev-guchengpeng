QCHECKSTYLE=go run $(GOPATH)/src/vendor/github.com/qiniu/checkstyle/gocheckstyle/gocheckstyle.go -config=.qcodestyle
all:
	go install ./...
	gofmt -w -s .
	@$(QCHECKSTYLE) .
test:
	go test ./...
style:
	@$(QCHECKSTYLE)

gofmt-check:
	@test `gofmt -s -l . | wc -l` -eq 0
govet-check:
	find . -name '*.go' | xargs -L 1 go tool vet
