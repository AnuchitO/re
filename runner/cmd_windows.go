//go:build windows
// +build windows

package runner

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type taskListOutput struct {
	out string
}

func (k *taskListOutput) Write(p []byte) (n int, err error) {
	k.out = string(p)
	return len(p), nil
}

func isNoTaskRunning(output string) bool {
	if strings.Contains(output, "No tasks") {
		return true
	}
	return false
}

func kill(pid int) error {
	o := &taskListOutput{}
	tasklist := exec.Command("TASKLIST", "/fi", "pid eq "+strconv.Itoa(pid))
	tasklist.Stderr = os.Stderr
	tasklist.Stdout = o
	err := tasklist.Run()
	if err != nil {
		log.Println("tasklist err", err)
		return err
	}

	if isNoTaskRunning(o.out) {
		return nil
	}

	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	return kill.Run()
}

func (r *Runner) Start() error {
	cmd := exec.Command(r.prog, r.args...)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr

	r.cmd = cmd
	r.done = make(chan struct{})

	if err := cmd.Start(); err != nil {
		close(r.done)
		return err
	}

	go func() {
		cmd.Wait() //nolint:errcheck
		close(r.done)
	}()

	return nil
}

func (r *Runner) KillCommand() error {
	if r.cmd == nil {
		return nil
	}

	if r.cmd.Process == nil {
		return nil
	}

	// Process already exited naturally — nothing to kill.
	select {
	case <-r.done:
		return nil
	default:
	}

	pid := r.cmd.Process.Pid
	cmd := r.cmd

	if err := kill(pid); err != nil {
		log.Println("kill error: ", err)
		return err
	}

	select {
	case <-time.After(3 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Println("failed to kill: ", err)
		}
	case <-r.done:
	}

	return nil
}
