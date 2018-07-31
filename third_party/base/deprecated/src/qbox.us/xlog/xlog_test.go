package xlog

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type httpHeader http.Header

func (p httpHeader) ReqId() string {

	return p[reqidKey][0]
}

func (p httpHeader) Header() http.Header {

	return http.Header(p)
}

func TestNewWithHeader(t *testing.T) {

	reqid := "testnewwithheader"

	h := httpHeader(make(http.Header))
	h[logKey] = []string{"origin"}
	h[reqidKey] = []string{reqid}

	xlog := NewWith(h)
	xlog.Xput([]string{"append"})

	assert.Equal(t, h.ReqId(), reqid)
	assert.Equal(t, xlog.ReqId(), reqid)

	log := []string{"origin", "append"}
	assert.Equal(t, h[logKey], log)
	assert.Equal(t, xlog.Xget(), log)
}
