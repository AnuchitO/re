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

var syscallKill = func(pid int, sig syscall.Signal) error { return syscall.Kill(pid, sig) }

func (r *Runner) Start() error {
	cmd := exec.Command(r.prog, r.args...)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	r.cmd = cmd
	r.done = make(chan struct{})

	if err := cmd.Start(); err != nil {
		close(r.done)
		return err
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("process exited: %v", err)
		}
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

	// try soft kill; log but continue — the hard kill handles the timeout case
	if err := syscallKill(-pid, syscall.SIGINT); err != nil {
		log.Printf("SIGINT to process group %d: %v", pid, err)
	}
	select {
	case <-time.After(r.killTimeout):
		// go hard because soft is not always the solution
		if err := syscallKill(-pid, syscall.SIGKILL); err != nil {
			return errors.New("fail killing ongoing process")
		}
	case <-r.done:
	}

	return nil
}
