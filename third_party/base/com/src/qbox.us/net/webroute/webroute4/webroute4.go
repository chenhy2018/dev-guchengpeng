package main

import (
	"io"
	"net/http"
	"qbox.us/net/seshcookie"
	"qbox.us/net/webroute"
)

type Service struct {
}

func (r *Service) DoLogin(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	user := req.FormValue("user")
	sessions["userId"] = user
	io.WriteString(w, "Login as: "+user)
}

func (r *Service) DoSth(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	user, _ := sessions["userId"].(string)
	io.WriteString(w, "Do sth as: "+user)
}

func main() {
	service := &Service{}
	sessions := seshcookie.NewSessionManager("_test_session", "", "fweaf5e9aef9a3c6da5dwcb1e2b601f3251a536d")
	webroute.ListenAndServe(":8084", service, sessions)
}
