//go:build !windows
// +build !windows

package runner

import (
	"errors"
	"log"
	"os/exec"
	"syscall"
	"time"
)

func (r *Runner) Start() error {
	cmd := exec.Command(r.prog, r.args...)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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

	pid := r.cmd.Process.Pid
	done := make(chan struct{})
	go func() {
		if err := r.cmd.Wait(); err != nil {
			log.Printf("process exited: %v", err)
		}
		close(done)
	}()

	// try soft kill; log but continue — the hard kill handles the timeout case
	if err := syscall.Kill(-pid, syscall.SIGINT); err != nil {
		log.Printf("SIGINT to process group %d: %v", pid, err)
	}
	select {
	case <-time.After(3 * time.Second):
		// go hard because soft is not always the solution
		err := syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			return errors.New("fail killing ongoing process")
		}
	case <-done:
	}

	return nil
}
