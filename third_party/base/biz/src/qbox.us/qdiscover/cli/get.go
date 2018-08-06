package cli

import (
	"fmt"

	"qbox.us/qdiscover/discover"
)

type Get struct{}

func (c Get) Definition() string {
	return "get service"
}

func (c Get) Exec(args []string) error {
	if len(args) == 0 {
		c.Help()
		return nil
	}

	var info discover.ServiceInfo
	err := Client.ServiceGet(nil, &info, args[0])
	if err != nil {
		return err
	}
	fmt.Printf("%-20s%-10s%-10s%-50s\n", "addr", "name", "state", "lastUpdate")
	fmt.Printf("%-20s%-10s%-10s%-50s\n\n", info.Addr, info.Name, info.State, info.LastUpdate)
	for k, v := range info.Attrs {
		fmt.Printf("attrs.%s: %#v\n", k, v)
	}
	fmt.Println()
	for k, v := range info.Cfg {
		fmt.Printf("cfg.%s: %#v\n", k, v)
	}
	return nil
}

func (c Get) Help() {
	fmt.Println("usage: service/get <addr>")
}
