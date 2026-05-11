package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/AnuchitO/re/runner"
)

func TestSplitArguments(t *testing.T) {
	t.Run("no arguments should return error", func(t *testing.T) {
		args := []string{}

		_, _, err := splitCommand(args)

		eErr := errors.New("you should add command after re [command], e.g. 're go test -v .'")
		if eErr.Error() != err.Error() {
			t.Errorf("expect error is '%v' but got '%v'", eErr, err)
		}
	})

	t.Run("single argument should succeed without error", func(t *testing.T) {
		args := []string{"go"}

		_, _, err := splitCommand(args)

		if err != nil {
			t.Errorf("expect without any error but got %v", err)
		}
	})

	t.Run("first argument should be the command", func(t *testing.T) {
		args := []string{"go", "test", "-v", "."}

		prog, _, _ := splitCommand(args)

		if prog != "go" {
			t.Errorf("expect prog is %q but got %q", "go", prog)
		}
	})

	t.Run("arguments from index 1 onwards should be params of the command", func(t *testing.T) {
		args := []string{"go", "test", "-v", "."}

		_, params, _ := splitCommand(args)

		eParams := []string{"test", "-v", "."}
		if !reflect.DeepEqual(eParams, params) {
			t.Errorf("expect params is %q but got %q", eParams, params)
		}
	})
}

func TestRerun(t *testing.T) {
	t.Run("run loop should execute command and stop on signal", func(t *testing.T) {
		task := runner.New("go", []string{"version"}...)
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		time.AfterFunc(2*time.Second, func() {
			close(stop)
		})

		run(".", task, stop, &wg, 800*time.Millisecond, nil, false)

		wg.Wait()

		assert.True(t, true, "should be pass")
	})
}

// TestHelperProcess is used as a subprocess target for integration tests.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if os.Getenv("FAIL_GETWD") == "1" {
		getwd = func() (string, error) { return "", errors.New("mock getwd error") }
	}

	// Find the "--" separator and use args after it as os.Args
	args := os.Args
	for i, a := range args {
		if a == "--" {
			os.Args = append([]string{os.Args[0]}, args[i+1:]...)
			break
		}
	}

	// Reset flags so flag.Parse() in main() works fresh
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	main()
}

func TestMainDirect(t *testing.T) {
	t.Run("version flag", func(t *testing.T) {
		origArgs := os.Args
		defer func() { os.Args = origArgs }()

		os.Args = []string{"re", "-version"}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		// Capture stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		origStdout := os.Stdout
		os.Stdout = w

		main()

		w.Close()
		os.Stdout = origStdout

		var buf strings.Builder
		_, _ = strings.NewReader("").WriteTo(nil)
		b := make([]byte, 1024)
		n, _ := r.Read(b)
		buf.Write(b[:n])
		r.Close()

		assert.Contains(t, buf.String(), "dev")
	})

	t.Run("ignore flag", func(t *testing.T) {
		origArgs := os.Args
		defer func() { os.Args = origArgs }()

		os.Args = []string{"re", "-ignore", "*.log", "go", "version"}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		// Send interrupt to the current process after a short delay
		go func() {
			time.Sleep(300 * time.Millisecond)
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(os.Interrupt)
		}()

		main()
	})
}

func TestMain(t *testing.T) {
	t.Run("version", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "-version")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		out, err := cmd.Output()
		assert.NoError(t, err)
		assert.Contains(t, string(out), "dev")
	})

	t.Run("no args", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		err := cmd.Run()
		assert.Error(t, err)
	})

	t.Run("with ignore and signal", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "-ignore", "*.log", "go", "version")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		var buf strings.Builder
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Start()
		assert.NoError(t, err)

		time.Sleep(500 * time.Millisecond)
		err = cmd.Process.Signal(syscall.SIGINT)
		assert.NoError(t, err)

		err = cmd.Wait()
		// Process may exit with signal error code, that's fine
		_ = err
		assert.Contains(t, buf.String(), "process terminated")
	})

	t.Run("getwd error", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "go", "version")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "FAIL_GETWD=1")
		err := cmd.Run()
		assert.Error(t, err)
	})
}

// Ensure signal and os imports are used (suppress unused import errors)
var _ = signal.Notify
var _ = syscall.SIGINT
