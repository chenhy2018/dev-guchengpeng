//package main
package m3u8

import (
	"fmt"
	"testing"

	xlog "github.com/qiniu/xlog.v1"
)

func TestMakem3u8(t *testing.T) {
	var playlist []map[string]interface{}

	m1 := map[string]interface{}{
		"duration": 10.00,
		"url":      "http://pdwjeyj6v.bkt.clouddn.com/ts/testdeviceid6/1536117002275/1536117010257/1536116983692/7.ts?e=1536122749&token=JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ:aNokJK3jwd_99T8AtvoKlDS_n0E=",
	}

	m2 := map[string]interface{}{
		"duration": 9.88,
		"url":      "http://pdwjeyj6v.bkt.clouddn.com/ts/testdeviceid6/1536117010183/1536117018099/1536116983692/7.ts?e=1536122756&token=JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ:6BoDrHa0hs_IutuJla4hrkkeTj0=",
	}

	m3 := map[string]interface{}{
		"duration": 9.88,
		"url":      "http://pdwjeyj6v.bkt.clouddn.com/ts/testdeviceid6/1536117018116/1536117026053/1536116983692/7.ts?e=1536122763&token=JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ:4QNa5NOspUhIMZHQhkQAFsnEYI4=",
	}

	m4 := map[string]interface{}{
		"duration": 9.23,
		"url":      "http://pdwjeyj6v.bkt.clouddn.com/ts/testdeviceid6/1536117126015/1536117133972/1536116983692/7.ts?e=1536122770&token=JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ:6T7bO4z3z44V_z0uVzykJz9VwTs=",
	}

	playlist = append(playlist, m1)
	playlist = append(playlist, m2)
	playlist = append(playlist, m3)
	playlist = append(playlist, m4)

	res := Mkm3u8(playlist, xlog.NewDummy())
	fmt.Println(res)
}
