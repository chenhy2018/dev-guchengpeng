package main

import (
    "fmt"
)


func main() {
    fmt.Printf("enter main...\n")
    pPlaylist := new( MediaPlaylist )
    pPlaylist.Init( 32, 32 )
    pPlaylist.AppendSegment( "https://www.baidu.com/test1.ts", 10.00, "program" )
    pPlaylist.AppendSegment( "https://www.baidu.com/test2.ts", 9.88, "program" )
    pPlaylist.AppendSegment( "https://www.baidu.com/test3.ts", 9.88, "" )
    fmt.Printf( "%s\n", pPlaylist.String() );

}
