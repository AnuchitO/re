package runner

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Runner is the task runnde
type Runner struct {
	prog string
	args []string
	cmd  *exec.Cmd
}

var taskRunner *Runner

// NewRunner creates new task runner if not exists
func NewRunner(prog string, args ...string) *Runner {
	if taskRunner == nil {
		return &Runner{
			prog: prog,
			args: args,
		}
	}

	return taskRunner
}

func newCommandRunner(prog string, args ...string) *exec.Cmd {
	cmd := exec.Command(prog, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

// Run starts the runner
func (rn *Runner) Run() error {
	err := rn.KillCommand()
	if err != nil {
		return err
	}

	// start the new one
	rn.cmd = newCommandRunner(rn.prog, rn.args...)
	err = rn.cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func (rn *Runner) KillCommand() error {
	done := make(chan struct{})
	go func() {
		if rn.cmd != nil {
			rn.cmd.Wait()
		}
		close(done)
	}()

	// try soft kill
	if rn.cmd != nil && rn.cmd.Process != nil {
		syscall.Kill(-rn.cmd.Process.Pid, syscall.SIGINT)
		select {
		case <-time.After(3 * time.Second):
			// go hard because soft is not always the solution
			err := syscall.Kill(-rn.cmd.Process.Pid, syscall.SIGKILL)
			if err != nil {
				return errors.New("Fail killing on going process")
			}
		case <-done:
		}
	}

	return nil
}
