package main

import (
	"fmt"
	"os"
	"qbox.us/qmset/bloom/bloomutil"
	"strconv"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("qmbloomfp <N> <Fp>")
		return
	}

	n, err := strconv.ParseUint(os.Args[1], 10, 0)
	if err != nil {
		fmt.Println("Invalid <N>:", err)
		return
	}

	fp, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Println("Invalid <Fp>:", err)
		return
	}

	fmt.Println("n =", n, "fp =", fp)

	f := bloomutil.NewWithEstimates(uint(n), fp)
	fmt.Println("m, k, bits:", f.Cap(), f.K(), float64(f.Cap())/float64(n))
	fmt.Println("memory requirement (in M):", f.Cap()/(1024*1024*8))
}

// ------------------------------------------------------------------------
