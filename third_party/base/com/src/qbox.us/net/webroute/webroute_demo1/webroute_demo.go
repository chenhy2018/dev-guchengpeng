package main

import (
	"fmt"
	"io"
	"net/http"
	"github.com/qiniu/log.v1"
	"qbox.us/net/session"
	"qbox.us/net/webroute"
	"strconv"
	"time"
)

// ---------------------------------------------------------------------------

type Service struct {
}

func (r *Service) Do_(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	sessions["map"] = map[string]string{
		"a": "A",
		"b": "B",
	}
	io.WriteString(w, "Do_: "+req.URL.String())
}

func (r *Service) DoFoo_bar(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	sessions["map1"] = map[string]string{
		"c": "C",
		"d": "D",
	}
	io.WriteString(w, "DoFoo_bar: "+req.URL.String())
}

func (r *Service) DoFoo_bar_(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	fmt.Println(sessions)
	io.WriteString(w, "DoFoo_bar_: "+req.URL.String())
}

func (r *Service) DoPage(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	delete(sessions, "map1")
	io.WriteString(w, "DoPage: "+req.URL.String())
}

func (r *Service) DoPageAction1(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPageAction1: "+req.URL.String())
}

// ---------------------------------------------------------------------------

func (r *Service) DoLogin(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	//sessions["userId"] = req.FormValue("user")
	sessions["userId"] = time.Now().UnixNano()
	io.WriteString(w, "DoLogin: "+req.URL.String())
}

func (r *Service) DoSth(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	userId, _ := sessions["userId"].(int64)
	user := strconv.FormatInt(userId, 10)
	io.WriteString(w, "DoSth: "+user)
}

// ---------------------------------------------------------------------------

func main() {
	service := &Service{}
	//sessions := session.New("", 0, 0)
	d, _ := time.ParseDuration("24h")
	sessions := session.New("", 3600, d)
	log.SetOutputLevel(0)
	log.Fatal(webroute.ListenAndServe(":8080", service, sessions))
}

// ---------------------------------------------------------------------------
