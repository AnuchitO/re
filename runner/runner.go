package runner

import (
	"io"
	"os"
	"os/exec"
	"time"
)

// Runner is the task runnde
type Runner struct {
	prog        string
	args        []string
	cmd         *exec.Cmd
	done        chan struct{} // closed when cmd exits naturally or is killed
	stdout      io.Writer
	stderr      io.Writer
	killTimeout time.Duration
}

// New creates new task runner if not exists
func New(prog string, args ...string) *Runner {
	return &Runner{
		prog:        prog,
		args:        args,
		stderr:      os.Stderr,
		stdout:      os.Stdout,
		killTimeout: 3 * time.Second,
	}
}

type iRunner interface {
	Start() error
	KillCommand() error
}

// Done returns a channel that is closed when the command exits.
func (r *Runner) Done() <-chan struct{} {
	return r.done
}

// SetStdout replaces the writer used for the command's stdout.
// Must be called before Run/Start.
func (r *Runner) SetStdout(w io.Writer) { r.stdout = w }

// SetStderr replaces the writer used for the command's stderr.
// Must be called before Run/Start.
func (r *Runner) SetStderr(w io.Writer) { r.stderr = w }


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
