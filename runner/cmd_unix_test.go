//go:build !windows
// +build !windows

package runner

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKillCommandNilProcess(t *testing.T) {
	t.Run("should return nil when no command has been started", func(t *testing.T) {
		r := &Runner{stdout: os.Stdout, stderr: os.Stderr}

		err := r.KillCommand()

		assert.NoError(t, err, "KillCommand on an unstarted runner should return nil")
	})

	t.Run("should return nil when cmd is set but process is nil", func(t *testing.T) {
		r := &Runner{
			stdout: os.Stdout,
			stderr: os.Stderr,
		}
		// simulate cmd set without a started process
		r.cmd = nil

		err := r.KillCommand()

		assert.NoError(t, err, "KillCommand with nil cmd should return nil")
	})
}

func TestKillCommandRunningProcess(t *testing.T) {
	t.Run("should kill a running process without error", func(t *testing.T) {
		r := &Runner{
			prog:   "sleep",
			args:   []string{"30"},
			stdout: os.Stdout,
			stderr: os.Stderr,
		}

		err := r.Start()
		assert.NoError(t, err, "should start sleep command successfully")

		// Give the OS time to launch the process
		time.Sleep(50 * time.Millisecond)

		err = r.KillCommand()
		assert.NoError(t, err, "should kill the running process without error")
	})
}
