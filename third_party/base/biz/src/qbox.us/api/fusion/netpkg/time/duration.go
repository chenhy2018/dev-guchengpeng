package time

import (
	"errors"
	"strconv"
	"time"
)

var errInvalidDur = errors.New("invalid duration")

// 增强 time.ParseDuration，支持单位天，用 d 表示。
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, errInvalidDur
	}
	unit := s[len(s)-1]
	if unit == 'd' { // day to hour
		v, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, errInvalidDur
		}
		v = v * 24
		s = strconv.Itoa(v) + "h"
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return 0, errInvalidDur
	}
	return dur, nil
}
