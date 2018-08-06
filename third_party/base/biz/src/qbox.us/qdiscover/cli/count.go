package cli

import (
	"flag"
	"fmt"

	"qbox.us/qdiscover/discover"
)

var (
	coutFlags                        *flag.FlagSet
	countNode, countName, countState string
)

func init() {
	coutFlags = flag.NewFlagSet("count", flag.ExitOnError)
	coutFlags.StringVar(&countNode, "node", "", "node name")
	coutFlags.StringVar(&countName, "name", "", "service name")
	coutFlags.StringVar(&countState, "state", "", "service state")
}

type Count struct{}

func (c Count) Definition() string {
	return "count services"
}

func (c Count) Exec(args []string) error {
	coutFlags.Parse(args)

	names, state, err := formalNameAndState(countName, countState)
	if err != nil {
		return err
	}

	queryArgs := &discover.QueryArgs{
		Node:  countNode,
		Name:  names,
		State: string(state),
	}
	count, err := Client.ServiceCountEx(nil, queryArgs)
	if err != nil {
		return fmt.Errorf("count failed: %v", err)
	}
	fmt.Println("count:", count)
	return nil
}

func (c Count) Help() {
	fmt.Printf("usage of service/count:\n\n")
	coutFlags.PrintDefaults()
}
