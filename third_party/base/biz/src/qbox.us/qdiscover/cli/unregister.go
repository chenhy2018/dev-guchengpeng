package cli

import (
	"fmt"
)

type Unregister struct{}

func (c Unregister) Definition() string {
	return "unregister service"
}

func (c Unregister) Exec(args []string) error {
	if len(args) == 0 {
		c.Help()
		return nil
	}
	return Client.ServiceUnregister(nil, args[0])
}

func (c Unregister) Help() {
	fmt.Println("usage: service/unregister <addr>")
}
