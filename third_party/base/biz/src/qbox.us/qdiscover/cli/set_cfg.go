package cli

import (
	"fmt"
	"strconv"

	"qbox.us/qdiscover/discover"
)

type SetCfg struct{}

func (c SetCfg) Definition() string {
	return "set cfg"
}

func (c SetCfg) Exec(args []string) error {
	if len(args) < 3 {
		c.Help()
		return nil
	}

	var value interface{}
	v, err := strconv.Atoi(args[2])
	if err != nil {
		value = args[2]
	} else {
		value = v
	}
	if args[1] != "fop_weight" && args[1] != "fop_mode" {
		c.Help()
		return nil
	}
	cfg := discover.CfgArgs{Key: args[1], Value: value}
	return Client.ServiceSetCfg(nil, args[0], &cfg)
}

func (c SetCfg) Help() {
	fmt.Println("usage: service/setcfg <addr> <key(fop_weight/fop_mode)> <value>")
}
