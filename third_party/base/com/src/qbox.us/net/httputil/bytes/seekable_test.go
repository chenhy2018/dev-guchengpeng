package bytes

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeekable_EOFIfReqAlreadyParsed(t *testing.T) {
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	req.ParseForm()
	_, err = Seekable(req)
	assert.Equal(t, err.Error(), "EOF")
}

func TestSeekable_WorkaroundForEOF(t *testing.T) {
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, _ = Seekable(req)
	req.ParseForm()
	assert.Equal(t, req.FormValue("a"), "1")
	_, err = Seekable(req)
	assert.NoError(t, err)
}

func TestSeekable(t *testing.T) {
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, err = Seekable(req)
	assert.NoError(t, err)
}

func TestSeekableLength(t *testing.T) {
	old := MaxSeekableLength
	defer func() {
		MaxSeekableLength = old
	}()
	MaxSeekableLength = 2
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, err = Seekable(req)
	assert.Error(t, err)
}
