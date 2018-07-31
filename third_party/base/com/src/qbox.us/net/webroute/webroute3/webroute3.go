package main

import (
	"io"
	"net/http"
	"qbox.us/net/webroute"
)

type Service struct {
}

func (r *Service) Do_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /")
}

func (r *Service) Do_foo_bar(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /foo/bar")
}

func (r *Service) Do_foo_bar_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /foo/bar/")
}

func (r *Service) Do_saveAs(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /saveAs")
}

func (r *Service) Do_saveAs_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /saveAs/")
}

func main() {
	service := &Service{}
	router := webroute.Router{Style: '/', Mux: http.DefaultServeMux}
	router.Register(service)
	http.ListenAndServe(":8083", nil)
}
