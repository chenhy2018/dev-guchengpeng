package cli

import (
	"fmt"
)

type Enable struct{}

func (c Enable) Definition() string {
	return "enable service"
}

func (c Enable) Exec(args []string) error {
	if len(args) == 0 {
		c.Help()
		return nil
	}
	return Client.ServiceEnable(nil, args[0])
}

func (c Enable) Help() {
	fmt.Println("usage: service/enable <addr>")
}
