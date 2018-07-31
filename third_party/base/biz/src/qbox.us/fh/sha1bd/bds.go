package sha1bd

import (
	"encoding/binary"
)

func DecodeFh(rawFh []byte) (fh []byte, bds [4]uint16) {

	fh = rawFh
	bds = [4]uint16{0, 0xffff, 0xffff, 0xffff}
	switch len(fh) % 20 {
	case 10:
		for i := 0; i < 3; i++ {
			bds[i] = binary.LittleEndian.Uint16(fh[i*2 : (i+1)*2])
			if bds[i] == 0xffff {
				break
			}
		}
		fh = fh[7:]
		fallthrough
	case 3:
		bds[3] = binary.LittleEndian.Uint16(fh[:2])
		fh = fh[2:]
	}
	return
}
