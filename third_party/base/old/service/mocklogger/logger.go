package main

import (
	"fmt"
	"http"
	"io"
	"strconv"
)

func put(w http.ResponseWriter, req *http.Request) {
	m, err := http.ParseQuery(req.URL.RawQuery)
	s := m["len"][0]
	l, err := strconv.Atoi(s)
	if err != nil {
		w.WriteHeader(500)
	}

	buff := make([]byte, l)
	_, err = io.ReadFull(req.Body, buff)
	if err != nil {
		w.WriteHeader(500)
	}
	fmt.Println(string(buff))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/put", put)
	http.ListenAndServe("0.0.0.0:10006", mux)
}
