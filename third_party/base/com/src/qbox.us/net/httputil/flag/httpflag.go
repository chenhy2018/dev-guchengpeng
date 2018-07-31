package flag

import (
	"qbox.us/errors"
	"strconv"
	"strings"
)

/*
Command line:
	METHOD <Command>/<MainParam>/<Switch1>/<Switch1Param>/.../<SwitchN>/<SwitchNParam>
*/

type SW struct {
	Name string
	Val  interface{} // must
	Has  *bool       // opt
}

func Parse(cmdline string, mainsw SW, sw ...SW) (err error) {

	return ParseEx(strings.Split(cmdline, "/"), mainsw, sw...)
}

func ParseEx(query []string, mainsw SW, sw ...SW) (err error) {

	if len(query) > 1 {
		err = parseOne(query[1], mainsw.Val, mainsw.Has)
		if err != nil {
			err = errors.Info(errors.EINVAL, "Parse main param failed -", query[1]).Detail(err)
			return
		}
		for i := 3; i < len(query); i += 2 {
			for j := range sw {
				if query[i-1] == sw[j].Name {
					err = parseOne(query[i], sw[j].Val, sw[j].Has)
					if err != nil {
						err = errors.Info(errors.EINVAL, "Parse switch failed -", query[i-1], query[i]).Detail(err)
						return
					}
				}
			}
		}
	}
	return
}

var ErrUnsupportedType = errors.New("unsupported type")

func parseOne(s string, val interface{}, has *bool) (err error) {

	switch v := val.(type) {
	case *string:
		*v = s
	case *int:
		*v, err = strconv.Atoi(s)
	case *int64:
		*v, err = strconv.ParseInt(s, 10, 64)
	case *uint32:
		var v64 uint64
		v64, err = strconv.ParseUint(s, 10, 32)
		*v = uint32(v64)
	case *uint:
		var v64 uint64
		v64, err = strconv.ParseUint(s, 10, 0)
		*v = uint(v64)
	case *int32:
		var v64 int64
		v64, err = strconv.ParseInt(s, 10, 32)
		*v = int32(v64)
	case *bool:
		*v = (s != "0" && strings.ToLower(s) != "false")
	case *uint64:
		*v, err = strconv.ParseUint(s, 10, 64)
	case *float32:
		var v64 float64
		v64, err = strconv.ParseFloat(s, 32)
		*v = float32(v64)
	case *float64:
		*v, err = strconv.ParseFloat(s, 64)
	default:
		err = ErrUnsupportedType
		return
	}
	if has != nil {
		*has = true
	}
	return
}
