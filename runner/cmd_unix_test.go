//go:build !windows
// +build !windows

package runner

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKillCommand(t *testing.T) {
	t.Run("nil process should return nil when no command has been started", func(t *testing.T) {
		r := &Runner{stdout: os.Stdout, stderr: os.Stderr}

		err := r.KillCommand()

		assert.NoError(t, err, "KillCommand on an unstarted runner should return nil")
	})

	t.Run("nil process should return nil when cmd is set but process is nil", func(t *testing.T) {
		r := &Runner{
			stdout: os.Stdout,
			stderr: os.Stderr,
		}
		// simulate cmd set without a started process
		r.cmd = nil

		err := r.KillCommand()

		assert.NoError(t, err, "KillCommand with nil cmd should return nil")
	})

	t.Run("nil process should return nil when cmd is set but Process field is nil", func(t *testing.T) {
		r := &Runner{
			stdout: os.Stdout,
			stderr: os.Stderr,
		}
		// Set cmd but leave Process as nil (not yet started)
		cmd := exec.Command("echo", "hello")
		r.cmd = cmd
		// cmd.Process is nil because we haven't called Start

		err := r.KillCommand()

		assert.NoError(t, err, "KillCommand with cmd.Process == nil should return nil")
	})

	t.Run("running process should kill without error", func(t *testing.T) {
		r := &Runner{
			prog:        "sleep",
			args:        []string{"30"},
			stdout:      os.Stdout,
			stderr:      os.Stderr,
			killTimeout: 3 * time.Second,
		}

		err := r.Start()
		assert.NoError(t, err, "should start sleep command successfully")

		// Give the OS time to launch the process
		time.Sleep(50 * time.Millisecond)

		err = r.KillCommand()
		assert.NoError(t, err, "should kill the running process without error")
	})

	t.Run("already exited", func(t *testing.T) {
		r := New("echo", "hello")
		err := r.Start()
		assert.NoError(t, err)
		<-r.Done()
		err = r.KillCommand()
		assert.NoError(t, err)
	})

	t.Run("sigkill timeout", func(t *testing.T) {
		r := New("sh", "-c", "trap '' INT; sleep 30")
		r.killTimeout = 50 * time.Millisecond
		err := r.Start()
		assert.NoError(t, err)
		time.Sleep(20 * time.Millisecond)
		err = r.KillCommand()
		assert.NoError(t, err)
	})

	t.Run("sigkill error", func(t *testing.T) {
		origKill := syscallKill
		syscallKill = func(pid int, sig syscall.Signal) error {
			return errors.New("mock error")
		}

		r := New("sleep", "30")
		r.killTimeout = 50 * time.Millisecond
		err := r.Start()
		assert.NoError(t, err)
		time.Sleep(20 * time.Millisecond)

		err = r.KillCommand()
		assert.Error(t, err)
		assert.Equal(t, "fail killing ongoing process", err.Error())

		// Restore and clean up the process
		syscallKill = origKill
		if r.cmd != nil && r.cmd.Process != nil {
			_ = syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)
		}
		<-r.done
	})

	t.Run("sigint error", func(t *testing.T) {
		origKill := syscallKill
		// SIGINT fails but SIGKILL succeeds (so process dies)
		syscallKill = func(pid int, sig syscall.Signal) error {
			if sig == syscall.SIGINT {
				return errors.New("sigint mock error")
			}
			// Real SIGKILL
			return syscall.Kill(pid, sig)
		}
		defer func() { syscallKill = origKill }()

		r := New("sleep", "30")
		r.killTimeout = 50 * time.Millisecond
		err := r.Start()
		assert.NoError(t, err)
		time.Sleep(20 * time.Millisecond)

		err = r.KillCommand()
		// SIGKILL succeeds, so no error
		assert.NoError(t, err)
		<-r.done
	})
}

func TestDone(t *testing.T) {
	r := New("echo", "hello")
	err := r.Start()
	assert.NoError(t, err)
	select {
	case <-r.Done():
		// channel closed after process exits
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for Done channel")
	}
}

func TestSetStdout(t *testing.T) {
	var buf bytes.Buffer
	r := New("echo", "hello")
	r.SetStdout(&buf)
	err := r.Start()
	assert.NoError(t, err)
	<-r.Done()
	assert.Contains(t, buf.String(), "hello")
}

func TestSetStderr(t *testing.T) {
	var buf bytes.Buffer
	r := New("sh", "-c", "echo err >&2")
	r.SetStderr(&buf)
	err := r.Start()
	assert.NoError(t, err)
	<-r.Done()
	assert.Contains(t, buf.String(), "err")
}
