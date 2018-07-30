package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"qbox.us/qdiscover/discover"
)

var (
	listFlags *flag.FlagSet

	listNode, listName, listState, listMarker string
	listLimit                                 int
)

func init() {
	listFlags = flag.NewFlagSet("list", flag.ExitOnError)
	listFlags.StringVar(&listName, "name", "", "service name")
	listFlags.StringVar(&listNode, "node", "", "node name")
	listFlags.StringVar(&listState, "state", "", "service state")
	listFlags.StringVar(&listMarker, "marker", "", "list marker")
	listFlags.IntVar(&listLimit, "limit", 0, "list limit")
}

func formalNameAndState(nameStr, stateStr string) (names []string, state discover.State, err error) {
	if nameStr != "" {
		names = strings.Split(nameStr, ",")
	}
	if stateStr != "" {
		var ok bool
		if state, ok = discover.ValidState(stateStr); !ok {
			err = errors.New("invalid state")
			return
		}
	}
	return
}

type List struct{}

func (c List) Definition() string {
	return "list services"
}

func (c List) Exec(args []string) error {
	listFlags.Parse(args)

	names, state, err := formalNameAndState(listName, listState)
	if err != nil {
		return err
	}

	queryArgs := &discover.QueryArgs{
		Node:  listNode,
		Name:  names,
		State: string(state),
	}
	var list discover.ServiceListRet
	err = Client.ServiceListEx(nil, &list, queryArgs, listMarker, listLimit)
	if err != nil {
		return fmt.Errorf("list failed: %v", err)
	}
	printListRet(list)
	return nil
}

func printListRet(ret discover.ServiceListRet) {
	if ret.Marker != "" {
		fmt.Println("marker:", ret.Marker)

	}
	type Attrs struct {
		Processing int64 `bson:"processing"`
	}

	type Cfg struct {
		LastChange string `bson:"last_change"`
	}

	var attrs Attrs
	var cfg Cfg
	if len(ret.Items) > 0 {
		fmt.Printf("\n%-10s%-25s%-15s%-15s%-35s%-35s%-10s\n", "number", "addr", "name", "state", "lastUpdate", "lastChange", "processing")
		for i, info := range ret.Items {
			info.Attrs.ToStruct(&attrs)
			info.Cfg.ToStruct(&cfg)
			fmt.Printf("%-10d%-25s%-15s%-15s%-35s%-35s%-10d\n", i+1, info.Addr, info.Name, info.State, info.LastUpdate, cfg.LastChange, attrs.Processing)
		}
	}
}

func (c List) Help() {
	fmt.Println("usage of service/list:")
	listFlags.PrintDefaults()
}
