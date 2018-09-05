package main

import (
	"io"
	"net/http"
	"github.com/qiniu/log.v1"
	"qbox.us/net/seshcookie"
	"qbox.us/net/webroute"
)

// ---------------------------------------------------------------------------

type Service struct {
}

func (r *Service) Do_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Do_: "+req.URL.String())
}

func (r *Service) DoFoo_bar(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoFoo_bar: "+req.URL.String())
}

func (r *Service) DoFoo_bar_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoFoo_bar_: "+req.URL.String())
}

func (r *Service) DoPage(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPage: "+req.URL.String())
}

func (r *Service) DoPageAction1(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPageAction1: "+req.URL.String())
}

// ---------------------------------------------------------------------------

func (r *Service) DoLogin(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	sessions["userId"] = req.FormValue("user")
	io.WriteString(w, "DoLogin: "+req.URL.String())
}

func (r *Service) DoSth(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	userId, _ := sessions["userId"].(string)
	io.WriteString(w, "DoSth: "+userId)
}

// ---------------------------------------------------------------------------

func main() {
	service := &Service{}
	sessions := seshcookie.NewSessionManager("_test_session", "", "fweaf5e9aef9a3c6da5dwcb1e2b601f3251a536d")
	log.SetOutputLevel(0)
	log.Fatal(webroute.ListenAndServe(":8080", service, sessions))
}

// ---------------------------------------------------------------------------
