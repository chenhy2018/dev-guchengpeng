//package main
package m3u8

import (
	"fmt"
	xlog "github.com/qiniu/xlog.v1"
)

func TestMakem3u8_old() {
	fmt.Printf("enter main...\n")
	pPlaylist := new(MediaPlaylist)
	pPlaylist.Init(32, 32)
	pPlaylist.AppendSegment("https://www.baidu.com/test1.ts", 10.00, "program")
	pPlaylist.AppendSegment("https://www.baidu.com/test2.ts", 9.88, "program")
	pPlaylist.AppendSegment("https://www.baidu.com/test3.ts", 9.88, "")
	fmt.Printf("%s\n", pPlaylist.String())

}

func TestMakem3u8() {
    var playlist []map[string]interface{}

    m1 := map[string]interface{}{
        "duration": 10.00,
        "url":      "http://pcgtsa42m.bkt.clouddn.com/7/testuid5/testdeviceid5/1533795689087/1533795686553.ts",
    }

    m2 := map[string]interface{}{
        "duration": 9.88,
        "url":      "http://pcgtsa42m.bkt.clouddn.com/7/testuid5/testdeviceid5/1533795689087/1533795686552.ts",
    }

    m3 := map[string]interface{}{
        "duration": 9.88,
        "url":      "http://pcgtsa42m.bkt.clouddn.com/7/testuid5/testdeviceid5/1533795689087/1533795686552.ts",
    }

    playlist = append(playlist, m1)
    playlist = append(playlist, m2)
    playlist = append(playlist, m3)

    res := Mkm3u8( playlist )
    fmt.Println( res )
}
