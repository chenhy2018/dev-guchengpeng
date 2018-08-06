package ufoputil

import (
	"net/http"
	"reflect"
)

type FlushWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func NewFlushWriter(w http.ResponseWriter) *FlushWriter {
	return &FlushWriter{w, false}
}

func (fw *FlushWriter) WriteHeader(code int) {
	fw.wroteHeader = true
	fw.ResponseWriter.WriteHeader(code)
}

func (fw *FlushWriter) Write(data []byte) (n int, err error) {
	defer func() {
		flusher, ok := getFlusher(fw.ResponseWriter)
		if ok {
			flusher.Flush()
		}
	}()
	if !fw.wroteHeader {
		fw.WriteHeader(200)
	}
	return fw.ResponseWriter.Write(data)
}

func getFlusher(i interface{}) (f http.Flusher, ok bool) {

	f, ok = i.(http.Flusher)
	if ok {
		return
	}

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	w := v.FieldByName("ResponseWriter")
	if !w.IsValid() || !w.CanInterface() {
		return
	}

	i = w.Interface()
	f, ok = i.(http.Flusher)
	return
}
