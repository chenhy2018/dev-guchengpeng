package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kavu/go_reuseport"
)

var (
	args               []string
	runLocker          sync.Mutex
	currentCmd         *exec.Cmd
	processes          = make(map[int]*exec.Cmd)
	isSupportReusePort bool
	envDefaultVar      = map[string]int{
		"RESTARTER_LISTEN_TRY_TIME":    20,
		"RESTARTER_LISTEN_INTERVAL_MS": 200,
	}
)

func getIntFromEnv(key string) int {

	str := os.Getenv(key)
	vint, err := strconv.Atoi(str)
	if err != nil {
		k, ok := envDefaultVar[key]
		if !ok {
			panic("cannot find default value of: " + key)
		}
		return k
	}
	return vint
}

func checkReusePort() {
	ln, err := reuseport.NewReusablePortListener("tcp4", "127.0.0.1:0")
	if err != nil {
		if strings.Contains(err.Error(), "protocol not available") {
			log.Println("SO_REUSEPORT is not supported in this kernal")
		} else {
			log.Println("restarter listen failed: ", err)
		}
		return
	}
	isSupportReusePort = true
	ln.Close()
}

func checkCommand() error {

	_, err := exec.Command("lsof", "-h").Output()
	if err != nil {
		return err
	}
	return nil
}

func main() {

	if len(os.Args) <= 1 {
		log.Fatal("??????????????")
	}
	if err := checkCommand(); err != nil {
		log.Fatal("checkCommand failed:", err)
	}

	checkReusePort()
	log.Println("run command:", os.Args)
	args = make([]string, len(os.Args)-1)
	copy(args, os.Args[1:])
	log.Println("restarter pid:", os.Getpid())

	run()
	processSignal()
}

func run() error {

	runLocker.Lock()
	defer runLocker.Unlock()

	cmd, err := newProcess()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	log.Println("start new process success:", cmd.Process.Pid)

	currentCmd = cmd
	processes[cmd.Process.Pid] = cmd

	go func() {
		err = cmd.Wait()
		log.Println("finish with:", cmd.Process.Pid, err)
		runLocker.Lock()
		delete(processes, cmd.Process.Pid)
		if cmd.Process.Pid == currentCmd.Process.Pid {
			if err == nil {
				os.Exit(0)
			}
			if eerr, ok := err.(*exec.ExitError); ok {
				if status, ok := eerr.Sys().(syscall.WaitStatus); ok {
					log.Printf("exit status: %d", status.ExitStatus())
					os.Exit(status.ExitStatus())
				}
			}
			os.Exit(-1)
		}
		runLocker.Unlock()
	}()

	return nil
}

func getListenAddrs(pid int) (addrs []string) {

	pidS := strconv.Itoa(pid)
	output, err := exec.Command("lsof", "-p", pidS, "-nP").Output()
	if err != nil {
		log.Printf("get listen addrs of pid(%s) failed, err: %v", pidS, err)
		return
	}
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		if bytes.Contains(line, []byte("LISTEN")) {
			fields := bytes.Split(line, []byte(" "))
			if len(fields) <= 1 {
				log.Println("invalid lsof output:", string(line))
				continue
			}
			addrs = append(addrs, string(fields[len(fields)-2]))
		}
	}
	return
}

func waitListen(oldPid, newPid int) bool {

	oldAddrs := getListenAddrs(oldPid)
	log.Println("oldAddrs:", oldAddrs)
	tryTime := getIntFromEnv("RESTARTER_LISTEN_TRY_TIME")
	intervalMs := time.Duration(getIntFromEnv("RESTARTER_LISTEN_INTERVAL_MS"))
	for i := 0; i < tryTime; i++ {
		newAddrs := getListenAddrs(newPid)
		log.Println("newAddrs:", newAddrs)
		if subset(newAddrs, oldAddrs) {
			return true
		}
		time.Sleep(time.Millisecond * intervalMs)
	}
	return false
}

func subset(base, sub []string) bool {

	for _, s := range sub {
		if !contain(base, s) {
			return false
		}
	}
	return true
}

func contain(base []string, sub string) bool {

	for _, s := range base {
		if s == sub {
			return true
		}
	}
	return false
}

func processSignal() {

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for {
		s := <-c
		log.Println("signal received:", s)
		switch s {
		case syscall.SIGCHLD:
			// SIGCHLD: receive when subprocess exit, so no need send it to subprocess
		case syscall.SIGUSR2:
			// restart
			oldCmd := currentCmd
			if isSupportReusePort {
				// 1. start new process
				err := run()
				if err != nil {
					log.Println("run new process failed, old is remained:", err)
				} else {
					waitSuccess := waitListen(oldCmd.Process.Pid, currentCmd.Process.Pid)
					if !waitSuccess {
						log.Println("wait listen failed, but continue")
					}

					// 2. send signal to old process
					err := oldCmd.Process.Signal(syscall.SIGTERM)
					if err != nil {
						log.Println("send SIGTERM to old failed:", oldCmd.Process.Pid, err)
					}
				}
			} else {
				// 2. send signal to old process
				err := oldCmd.Process.Signal(syscall.SIGTERM)
				if err != nil {
					log.Println("send SIGTERM to old failed:", oldCmd.Process.Pid, err)
				} else {
					// 2. start new process
					err = run()
					if err != nil {
						log.Println("run new process failed", err)
						// sleep 1s and try run again
						time.Sleep(time.Second)
						err = run()
						if err != nil {
							log.Fatal("run new process failed", err)
						}
					}
				}
			}
		default:
			runLocker.Lock()
			for _, proc := range processes {
				err := proc.Process.Signal(s)
				if err != nil {
					log.Printf("send signal to %d failed, err: %v", proc.Process.Pid, err)
				}
			}
			runLocker.Unlock()
		}
	}
}
