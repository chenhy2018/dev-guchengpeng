package main

import (
	"fmt"
	"github.com/qiniu/fsw"
	"github.com/qiniu/log.v1"
	"os"
	"qbox.us/cc/signal"
)

func main() {

	log.SetOutputLevel(0)

	if len(os.Args) < 2 {
		fmt.Println("Usage qfsmon <WatchingDir>")
		return
	}
	dir := os.Args[1]
	fmt.Println("Watching", dir, "...")

	watcher, err := fsw.Open(dir)
	if err != nil {
		log.Fatal(err)
	}

	watcher.Start(nil, func(ev fsw.Event) {
		log.Println(ev.String())
	})

	signal.WaitForInterrupt()
	watcher.Close()
}
