package retrys

import (
	"encoding/base64"
	"io"
	"strings"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/bdgetter/internal/bufpool"
	"qbox.us/errors"
	. "qbox.us/fh/proto"
)

type retrysGetter struct {
	getter ReaderGetter
	retrys int
}

func NewRetrys(g ReaderGetter, retrys int) CommonGetter {

	return &retrysGetter{getter: g, retrys: retrys}
}

// -----------------------------------------------------------------------------

func isWriterErr(err error) bool {

	return err == io.ErrShortWrite || strings.Contains(err.Error(), "write tcp")
}

func isRateLimitError(err error) bool {
	_, ok := err.(*errors.RateLimitError)
	return ok
}

func (p *retrysGetter) Get(xl1 rpc.Logger, fh []byte, w io.Writer, from, to int64) (n int64, err error) {

	xl := xlog.NewWith(xl1)
	var n1 int64
	var rc io.ReadCloser
	for trys := 0; trys < p.retrys+1; trys++ {
		rc, _, err = p.getter.Get(xl1, fh, from+n, to)
		if err != nil {
			// Failed to get a reader, no retry.
			xl.Warn("retrysGetter: getter.Get failed =>", err, base64.URLEncoding.EncodeToString(fh))
			break
		}

		n1, err = bufpool.Copy(w, rc)
		rc.Close()
		n += n1
		if err == nil {
			break
		}
		xl.Warn("retrysGetter.Get: Copy failed =>", err, n1)
		if from+n >= to || isWriterErr(err) || isRateLimitError(err) {
			break
		}
	}
	return
}
