// +build !linux

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func newProcess() (cmd *exec.Cmd, err error) {

	cmd = exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
	// only worked on linux
	//		Pdeathsig: syscall.SIGKILL,
	}

	return
}
