package ios

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"

	"qbox.us/etag"
)

// ----------------------------------------------------------

type EtagReader struct {
	Source       io.ReadSeeker
	Base, Offset int64
	Etag         *etag.Digest
}

func NewEtagReader(f io.ReadSeeker, chunkSize int) *EtagReader {

	return &EtagReader{Source: f, Etag: etag.New(chunkSize)}
}

func (r *EtagReader) Read(buf []byte) (n int, err error) {

	n, err = r.Source.Read(buf)
	if n == 0 {
		return
	}

	offset := r.Offset + int64(n)
	if r.Base == r.Offset {
		r.Etag.Write(buf[:n])
		r.Base, r.Offset = offset, offset
		return
	}

	if r.Base >= offset || r.Offset > r.Base {
		r.Offset = offset
		return
	}

	// r.Offset < r.Base < offset

	r.Etag.Write(buf[int(r.Base-r.Offset):])
	r.Base = offset

	return
}

func (r *EtagReader) Seek(offset int64, whence int) (ret int64, err error) {

	ret, err = r.Source.Seek(offset, whence)
	if err == nil {
		r.Offset = ret
	}
	return
}

// etag = http://docs.qbox.us/file-hash
func Validate(r *EtagReader, fsize int64, etag string) bool {

	if fsize != r.Base {
		log.Println("EtagReader.Validate failed: io not completed")
		return false
	}

	fh, err := base64.URLEncoding.DecodeString(etag)
	if err != nil || len(fh) != 21 {
		log.Println("EtagReader.Validate failed: invalid etag -", etag)
		return false
	}

	if (1 << uint(fh[0])) != r.Etag.ChunkSize() {
		log.Println("EtagReader.Validate failed: invalid chunkSize")
		return false
	}

	if !bytes.Equal(r.Etag.Sum(), fh[1:]) {
		log.Println("EtagReader.Validate failed: invalid hash")
		return false
	}
	return true
}

// ----------------------------------------------------------
