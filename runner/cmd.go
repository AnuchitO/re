// +build !windows

package runner

import (
	"errors"
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
		r.cmd.Wait()
		close(done)
	}()

	// try soft kill
	syscall.Kill(-pid, syscall.SIGINT)
	select {
	case <-time.After(3 * time.Second):
		// go hard because soft is not always the solution
		err := syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			return errors.New("Fail killing on going process")
		}
	case <-done:
	}

	return nil
}
