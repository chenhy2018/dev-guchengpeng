package agmapi

import (
	"github.com/qiniu/rpc.v1"
	"strconv"
)

// ------------------------------------------------------------------------

func Reload(agmHost string, l rpc.Logger, host string) error {

	return rpc.DefaultClient.CallWithForm(l, nil, agmHost+"/reload", map[string][]string{
		"host": {host},
	})
}

func Enable(agmHost string, l rpc.Logger, host string, server string, state int) error {

	return rpc.DefaultClient.CallWithForm(l, nil, agmHost+"/enable", map[string][]string{
		"host":   {host},
		"server": {server},
		"state":  {strconv.Itoa(state)},
	})
}

// ------------------------------------------------------------------------

type Info struct {
	Active int `json:"conn"`
}

func Query(agmHost string, l rpc.Logger, host string, server string) (ret Info, err error) {

	err = rpc.DefaultClient.CallWithForm(l, &ret, agmHost+"/query", map[string][]string{
		"host":   {host},
		"server": {server},
	})
	return
}

// ------------------------------------------------------------------------
