package lib

import (
	"github.com/stretchr/testify.v2/assert"
	"net/http"
	"testing"
	"time"
)

func TestGoSilent(t *testing.T) {
	c := make(chan interface{})
	goSilent(func() {
		panic("panic")
	}, func() {
		close(c)
	})
	<-c
}

func TestServeGoMetricsReusePort(t *testing.T) {
	address := "127.0.0.1:8084"
	err := ServeGoMetrics(address)
	assert.Nil(t, err)

	err = ServeGoMetrics(address)
	assert.NotNil(t, err)

	req, _ := http.NewRequest("GET", "http://"+address+"/metrics", nil)

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)

	assert.Nil(t, err)

	assert.Equal(t, 200, resp.StatusCode)
}
