package cli

import (
	"fmt"
	"os"
)

type Command interface {
	Definition() string
	Exec(args []string) error
}

type App struct {
	subCommand map[string]Command
}

func New() *App {
	return &App{subCommand: make(map[string]Command)}
}

func (a *App) AddCommand(name string, command Command) {
	a.subCommand[name] = command
}

func (a *App) Run() {
	args := os.Args[1:]
	if len(args) == 0 {
		a.exitWithHelp()
	}

	command, ok := a.subCommand[args[0]]
	if !ok {
		fmt.Println("error: no such command", args[0])
		a.exitWithHelp()
	}

	err := command.Exec(args[1:])
	if err != nil {
		exitWithErr(err)
	}
	os.Exit(0)
}

func (a *App) Help() {
	fmt.Printf("\nqboxdiscoverctl: service discovery command line tool\n")
	fmt.Printf("\nhost: %s\nyou can use DISCOVERD_HOST shell variable to change host.\n\n", DiscoverHost)
	fmt.Printf("sub-commands:\n\n")
	for n, command := range a.subCommand {
		fmt.Printf("  %-30s  ", n)
		fmt.Println(command.Definition())
	}
}

func (a *App) exitWithHelp() {
	a.Help()
	os.Exit(0)
}

func exitWithErr(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}
