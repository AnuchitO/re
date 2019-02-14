package runner

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

// runner is the task runnde
type runner struct {
	prog string
	args []string
	cmd  *exec.Cmd
}

var taskRunner *runner

// NewRunner creates new task runner if not exists
func NewRunner(prog string, args ...string) *runner {
	if taskRunner == nil {
		return &runner{
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
	return cmd
}

// Run starts the runner
func (rn *runner) Run() error {
	if rn.IsCommandRunning() {
		err := rn.KillCommand()
		if err != nil {
			return err
		}
		rn.cmd = nil
	}

	// start the new one
	rn.cmd = newCommandRunner(rn.prog, rn.args...)
	err := rn.cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		rn.cmd.Wait()
		rn.cmd = nil
	}()

	return nil
}

func (rn *runner) IsCommandRunning() bool {
	return rn.cmd != nil && rn.cmd.Process != nil
}

func (rn *runner) KillCommand() error {
	done := make(chan struct{})
	go func() {
		rn.cmd.Wait()
		close(done)
	}()

	// try soft kill
	rn.cmd.Process.Signal(os.Interrupt)
	select {
	case <-time.After(3 * time.Second):
		// go hard because soft is not always the solution
		err := rn.cmd.Process.Kill()
		if err != nil {
			return errors.New("Fail killing on going process")
		}
	case <-done:
	}

	return nil
}
