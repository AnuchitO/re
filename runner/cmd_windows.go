// +build windows

package runner

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func kill(cmd *exec.Cmd) error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	return kill.Run()
}

func (r *Runner) Start() error {
	cmd := exec.Command(r.prog, r.args...)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr

	r.cmd = cmd

	return r.cmd.Start()
}

func (r *Runner) KillCommand() error {
	if r.cmd == nil {
		return nil
	}

	if r.cmd.Process == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		r.cmd.Wait()
		close(done)
	}()

	if err := r.cmd.Process.Kill(); err != nil {
		log.Println("kill error: ", err)
		return err
	}

	select {
	case <-time.After(3 * time.Second):
		if err := r.cmd.Process.Kill(); err != nil {
			log.Println("failed to kill: ", err)
		}
	case <-done:
	}

	return nil
}
