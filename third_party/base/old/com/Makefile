TARGETS={ metric metric/np encryptor cc/strings cc/logging cc cc/binary cc/unsafe cc/signal cc/ts cc/filepath cc/filepath2 cc/cmd cc/runner cc/sha1 cc/objid cc/dht osl webutil etag/sha1 etag dev api rpc servestk servestk/logh auditlog auditlog/file auditlog/logh oauth qauth multipart mime stack encoding/binary net/rpc net/jsonrpc net/gobrpc }

all:
	@bmake ${TARGETS}

install:
	@bmake ${TARGETS} install

test:
	@bmake ${TARGETS} test

clean:
	@bmake ${TARGETS} clean
