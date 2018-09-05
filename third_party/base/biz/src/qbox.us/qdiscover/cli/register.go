package cli

import (
	"fmt"
)

type Register struct{}

func (c Register) Definition() string {
	return "register service"
}

func (c Register) Exec(args []string) error {
	if len(args) < 2 {
		c.Help()
		return nil
	}
	return Client.ServiceRegister(nil, args[0], args[1], nil)
}

func (c Register) Help() {
	fmt.Println("usage: service/register <addr> <name>")
}
