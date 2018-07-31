// +build go1.7

package lb

import (
	"net/http"
	"testing"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func TestIsCancelled(t *testing.T) {
	ast := assert.New(t)
	httpreq, err := http.NewRequest("GET", "http://127.0.0.1/test", nil)
	ast.NoError(err)
	xl := xlog.NewDummy()

	// Cancel inner http request
	{
		ctx, cancel := context.WithCancel(context.Background())
		httpreq2 := httpreq.WithContext(ctx)
		req := &Request{*httpreq2, nil, context.Background()}
		ast.False(isCancelled(xl, req))

		cancel()
		ast.True(isCancelled(xl, req))
	}

	// Cancel wrapper request
	{
		ctx, cancel := context.WithCancel(context.Background())
		req := &Request{*httpreq, nil, ctx}
		ast.False(isCancelled(xl, req))

		cancel()
		ast.True(isCancelled(xl, req))
	}

	// Cancel context of xLog
	{
		req := &Request{*httpreq, nil, context.Background()}
		ctx, cancel := context.WithCancel(context.Background())
		xl = xlog.NewDummyWithCtx(ctx)
		ast.False(isCancelled(xl, req))

		cancel()
		ast.True(isCancelled(xl, req))
	}

}
