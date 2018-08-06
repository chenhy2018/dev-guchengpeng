package state

import (
	"errors"
	"net/http"
	"github.com/qiniu/log.v1"
	"strconv"
)

// ----------------------------------------------------------

type ResponseWriterEx interface {
	http.ResponseWriter
	GetStatusCode() int
}

// ----------------------------------------------------------

type responseWriter struct {
	statusCode int
	http.ResponseWriter
}

func (r *responseWriter) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriter) GetStatusCode() int {
	return r.statusCode
}

func getResponseWriterEx(writer http.ResponseWriter) ResponseWriterEx {
	if itf, ok := writer.(ResponseWriterEx); ok {
		return itf
	}
	log.Info("getResponseWriterEx: make new responseWriter to get status code")
	return &responseWriter{http.StatusOK, writer}
}

// ----------------------------------------------------------

var ErrPaniced = errors.New("paniced")

func Handler(name string, f func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		err := ErrPaniced
		w1 := getResponseWriterEx(w)

		unit := state.Enter(name)
		defer unit.Leave(&err)

		f(w1, req)
		code := w1.GetStatusCode()
		if code == 200 {
			err = nil
		} else {
			err = errors.New("E" + strconv.Itoa(code))
		}
	}
}

// ----------------------------------------------------------
