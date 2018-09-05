package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	video, err := os.Open("./test2.ts")
	if err != nil {
		log.Fatal(err)

	}
	defer video.Close()

	http.ServeContent(w, r, "test2.ts", time.Now(), video)

}

func main() {
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8089", nil)

}
