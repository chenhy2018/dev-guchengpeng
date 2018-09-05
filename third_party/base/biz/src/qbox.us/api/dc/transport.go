package dc

import (
	"net/http"
	"sync/atomic"
)

// http.Transport.CloseIdleConnections 只是关闭 idle 的 conenctions.
// 当 CloseIdleConnections 被调用后且有新的请求到达时，http.Transport 仍然会被激活.
// 通常用到 http.Transport 的地方都是异步的，没有比较好的手段保证调用  CloseIdleConnections 之后没有请求到来.
// 也就是没有比较好的方式在并发环境中完整的闭一个 http.Tranpsport, 使用 CloseIdleConnections 避不开句柄和 goroutine 泄漏.
// 这里提供一个 Transport 封装了 http.Transport，当 Close 被调用后会在 RoundTrip 之后执行 CloseidleConnections.
// 这样可以在并发环境中异步完整地关闭 http.Transport 并回收其资源.
type Transport struct {
	*http.Transport
	closed int32
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {

	resp, err := t.Transport.RoundTrip(req)
	if atomic.LoadInt32(&t.closed) != 0 {
		t.CloseIdleConnections()
	}
	return resp, err
}

func (t *Transport) Close() {

	atomic.StoreInt32(&t.closed, 1)
	t.CloseIdleConnections()
}
