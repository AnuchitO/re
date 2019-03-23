package runner

import (
	"io"
	"os"
	"os/exec"
)

// Runner is the task runnde
type Runner struct {
	prog   string
	args   []string
	cmd    *exec.Cmd
	stdout io.Writer
	stderr io.Writer
}

var task *Runner

// New creates new task runner if not exists
func New(prog string, args ...string) *Runner {
	if task == nil {
		return &Runner{
			prog:   prog,
			args:   args,
			stderr: os.Stderr,
			stdout: os.Stdout,
		}
	}

	return task
}

type iRunner interface {
	Start() error
	KillCommand() error
}

// Run starts the runner
func (r *Runner) Run() error {
	return run(r)
}

func run(r iRunner) error {
	err := r.KillCommand()
	if err != nil {
		return err
	}

	err = r.Start()
	if err != nil {
		return err
	}

	return nil
}
