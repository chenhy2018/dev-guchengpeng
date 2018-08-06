package urlbd

import (
	. "qbox.us/fh/fhver"
)

// -----------------------------------------------------------------------------

func UrlOf(fh []byte) string {
	return string(fh[2:])
}

func Encode(url string) []byte {
	fh := make([]byte, len(url)+2)
	fh[0] = FhUrlbd
	fh[1] = 0x96
	copy(fh[2:], url)
	return fh
}

// -----------------------------------------------------------------------------
