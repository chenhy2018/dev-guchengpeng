package main

import (
	"qbox.us/qdiscover/cli"
)

func main() {
	app := cli.New()
	app.AddCommand("service/enable", cli.Enable{})
	app.AddCommand("service/disable", cli.Disable{})
	app.AddCommand("service/register", cli.Register{})
	app.AddCommand("service/unregister", cli.Unregister{})
	app.AddCommand("service/get", cli.Get{})
	app.AddCommand("service/setcfg", cli.SetCfg{})
	app.AddCommand("service/count", cli.Count{})
	app.AddCommand("service/listall", cli.ListAll{})
	app.Run()
}
