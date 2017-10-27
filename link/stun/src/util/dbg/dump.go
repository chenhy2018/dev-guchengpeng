package dbg

import (
	"fmt"
)

func PrintMem(buf []byte, col int) {

	output := fmt.Sprintf("-------------------- %d bytes --------------------", len(buf))
	for i, v := range buf {
		if i % col == 0 {
			output += "\n"
		}
		output += fmt.Sprintf("0x%02x ", v)
	}
	output += fmt.Sprintf("\n----------------- end of %d bytes -----------------", len(buf))
	fmt.Println(output)
}
