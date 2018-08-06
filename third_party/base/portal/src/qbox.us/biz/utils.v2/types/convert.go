package types

import (
	"strconv"
	"time"
)

func Timestamp2str(ts int64) string {
	return strconv.FormatInt(ts, 10)
}

func Str2timestamp(s string) int64 {
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Now().Unix()
	}

	return ts
}

func Int2str(i int) string {
	return strconv.Itoa(i)
}

func Str2int(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return i
}

func DigitString2Bytes(s string) []byte {
	digits := make([]byte, 0, len(s))
	for _, c := range s {
		if c >= '0' && c <= '9' {
			digits = append(digits, byte(c-'0'))
		}
	}
	return digits
}
