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

func (r *Service) DoFooBar(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /foo/bar")
}

func (r *Service) DoFooBar_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /foo/bar/")
}

func (r *Service) DoSave_as(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /save_as")
}

func (r *Service) DoSave_as_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "handle /save_as/")
}

func main() {
	service := &Service{}
	router := webroute.Router{Style: '_', Mux: http.DefaultServeMux}
	router.Register(service)
	http.ListenAndServe(":8082", nil)
}
