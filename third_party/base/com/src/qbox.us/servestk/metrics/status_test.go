package metrics

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	s := NewStatus()
	handler := s.Handler()

	onesecf := func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Second)
	}
	for i := 0; i < 10; i++ {
		go handler(nil, nil, onesecf)
	}

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, 10, s.Counter.Count(), "Count")
	time.Sleep(500 * time.Millisecond)

	onemsf := func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Millisecond)
	}
	for i := 0; i < 90; i++ {
		handler(nil, nil, onemsf)
	}
	assert.Equal(t, 100, s.Timer.Count(), "Timer")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, nil)
	b, _ := ioutil.ReadAll(w.Body)
	fmt.Println(string(b))
}
