package httputil

import (
	"net/http/httptest"
	"testing"

	"qbox.us/errors"

	"github.com/stretchr/testify/assert"

	qerrors "github.com/qiniu/errors"
	qhttputil "github.com/qiniu/http/httputil.v1"
)

func TestReplyWithCode(t *testing.T) {

	w := httptest.NewRecorder()
	ReplyWithCode(w, 200)
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Body.String(), "{}")

	w = httptest.NewRecorder()
	ReplyWithCode(w, 400)
	assert.Equal(t, w.Code, 400)
	assert.Equal(t, w.Body.String(), `{"error":"invalid argument"}`)

	w = httptest.NewRecorder()
	ReplyWithCode(w, 401)
	assert.Equal(t, w.Code, 401)
	assert.Equal(t, w.Body.String(), `{"error":"bad token"}`)
}

func doWrapper(err error) error {

	err = errors.Info(err, "abc")
	err = qerrors.Info(err, "efg").Detail(err)
	err = errors.Info(err, "hijk").Detail(err)
	return err
}

func TestDetectError(t *testing.T) {

	var err error

	err = NewError(403, "abc")
	err = doWrapper(err)
	code, msg := DetectError(err)
	if code != 403 || msg != "abc" {
		t.Fatal("DetectError failed:", code, msg)
	}

	err = qhttputil.NewError(405, "abcd")
	err = doWrapper(err)
	code, msg = DetectError(err)
	if code != 405 || msg != "abcd" {
		t.Fatal("DetectError failed:", code, msg)
	}
}
