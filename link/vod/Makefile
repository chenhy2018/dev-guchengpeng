all:
	cd src; go install -v ./...
clean:
	cd src; go clean -i ./...

test:
	cd src/qiniu.com/controllers; go test
	cd src/qiniu.com/models; go test
	cd src/qiniu.com/m3u8; go test
	cd src/qiniu.com/db; go test
