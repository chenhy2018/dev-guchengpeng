package cli

import (
	"fmt"
)

type Disable struct{}

func (c Disable) Definition() string {
	return "disable service"
}

func (c Disable) Exec(args []string) error {
	if len(args) == 0 {
		c.Help()
		return nil
	}
	return Client.ServiceDisable(nil, args[0])
}

func (c Disable) Help() {
	fmt.Println("usage: service/disable <addr>")
}
