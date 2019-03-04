package runner

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// Runner is the task runnde
type Runner struct {
	prog   string
	args   []string
	cmd    *exec.Cmd
	dir    string
	stdout io.Writer
	stderr io.Writer
}

var task *Runner

// New creates new task runner if not exists
func New(dir, prog string, args ...string) *Runner {
	if task == nil {
		return &Runner{
			prog:   prog,
			args:   args,
			dir:    dir,
			stderr: os.Stderr,
			stdout: os.Stdout,
		}
	}

	return task
}

func (r *Runner) Walk(lastMod time.Time) time.Time {
	filepath.Walk(r.dir, func(path string, fi os.FileInfo, err error) error {
		if path == ".git" && fi.IsDir() {
			log.Println("skipping .git directory")
			return filepath.SkipDir
		}

		// ignore hidden files
		if filepath.Base(path)[0] == '.' {
			return nil
		}

		if fi.ModTime().After(lastMod) {
			lastMod = time.Now()
			return errors.New("reload immediately: stop walking")
		}

		return nil
	})

	return lastMod
}

func (r *Runner) Start() error {
	cmd := exec.Command(r.prog, r.args...)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	r.cmd = cmd

	return r.cmd.Start()
}

// Run starts the runner
func (r *Runner) Run() error {
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

func (r *Runner) KillCommand() error {
	done := make(chan struct{})
	go func() {
		if r.cmd != nil {
			r.cmd.Wait()
		}
		close(done)
	}()

	if r.cmd != nil && r.cmd.Process != nil {
		// try soft kill
		syscall.Kill(-r.cmd.Process.Pid, syscall.SIGINT)
		select {
		case <-time.After(3 * time.Second):
			// go hard because soft is not always the solution
			err := syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)
			if err != nil {
				return errors.New("Fail killing on going process")
			}
		case <-done:
		}
	}

	return nil
}
