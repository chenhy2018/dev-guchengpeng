package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: mkfinitmap pkgName funcName keyType valType invalidVal")
		os.Exit(1)
	}

	var k, v uint64
	var dict = make(map[uint64]uint64)
	for {
		n, err := fmt.Scanln(&k, &v)
		if err != nil {
			if n == 0 {
				break
			}
			fmt.Fprintln(os.Stderr, "Scan data error:", err)
			os.Exit(2)
		}
		dict[k] = v
	}

	start := len(dict)

retry:
	{
		data := make([]uint64, start)
		hass := make([]bool, start)
		for k, v := range dict {
			i := k % uint64(start)
			if hass[i] {
				start++
				goto retry
			}
			hass[i] = true
			data[i] = v
		}

		pkgName := os.Args[1]
		funcName := os.Args[2]
		keyType := os.Args[3]
		valType := os.Args[4]
		invalidVal := os.Args[5]

		fmt.Printf(
			`package %s

func %s(k %s) %s {
	i := int(k %% %s(len(tbl_%s)))
	return tbl_%s[i]
}

var tbl_%s = [%d]%s{
`,
			pkgName, funcName, keyType, valType, keyType, funcName, funcName, funcName, start, valType)
		for i, has := range hass {
			if has {
				fmt.Printf("\t%v,\n", data[i])
			} else {
				fmt.Printf("\t%v,\n", invalidVal)
			}
		}
		fmt.Println("}")
	}
}
