package cli

import (
	"flag"
	"fmt"

	"qbox.us/qdiscover/discover"
)

var (
	listAllFlags *flag.FlagSet

	listAllName, listAllState string
)

func init() {
	listAllFlags = flag.NewFlagSet("listall", flag.ExitOnError)
	listAllFlags.StringVar(&listAllName, "name", "", "service name")
	listAllFlags.StringVar(&listAllState, "state", "", "service state")
}

type ListAll struct{}

func (c ListAll) Definition() string {
	return "list all services"
}

func (c ListAll) Exec(args []string) error {
	listAllFlags.Parse(args)

	names, state, err := formalNameAndState(listAllName, listAllState)
	if err != nil {
		return err
	}

	var list discover.ServiceListRet
	err = Client.ServiceListAll(nil, &list, names, state)
	if err != nil {
		return fmt.Errorf("listall failed: %v", err)
	}
	printListRet(list)
	return nil
}

func (c ListAll) Help() {
	fmt.Println("usage of service/listall:")
	listAllFlags.PrintDefaults()
}
