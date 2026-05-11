package main

import (
	"errors"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/AnuchitO/re/runner"
)

// mockRunner implements taskRunner for testing.
type mockRunner struct {
	runErr     error
	killErr    error
	doneCh     chan struct{}
	runCalled  int
	killCalled int
}

func newMockRunner(runErr, killErr error) *mockRunner {
	ch := make(chan struct{})
	close(ch) // done immediately
	return &mockRunner{
		runErr:  runErr,
		killErr: killErr,
		doneCh:  ch,
	}
}

func (m *mockRunner) SetStdout(_ io.Writer) {}
func (m *mockRunner) SetStderr(_ io.Writer) {}
func (m *mockRunner) Run() error {
	m.runCalled++
	return m.runErr
}
func (m *mockRunner) Done() <-chan struct{} { return m.doneCh }
func (m *mockRunner) KillCommand() error {
	m.killCalled++
	return m.killErr
}

func TestPrintRerunHeader(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		// isFancyTerminal returns false in test env (no TTY), plain path is exercised
		printRerunHeader()
	})

	t.Run("fancy", func(t *testing.T) {
		orig := isFancyTerminal
		isFancyTerminal = func() bool { return true }
		defer func() { isFancyTerminal = orig }()

		// Redirect stdout to avoid polluting test output
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		origStdout := os.Stdout
		os.Stdout = w

		printRerunHeader()

		w.Close()
		os.Stdout = origStdout
		r.Close()
	})
}

func TestSpinUntil(t *testing.T) {
	t.Run("done", func(t *testing.T) {
		done := make(chan struct{})
		stop := make(chan struct{})
		close(done)
		spinUntil(done, stop)
	})

	t.Run("stop", func(t *testing.T) {
		done := make(chan struct{})
		stop := make(chan struct{})
		close(stop)
		spinUntil(done, stop)
	})

	t.Run("ticks fancy", func(t *testing.T) {
		orig := isFancyTerminal
		isFancyTerminal = func() bool { return true }
		defer func() { isFancyTerminal = orig }()

		done := make(chan struct{})
		stop := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			spinUntil(done, stop)
		}()

		time.Sleep(200 * time.Millisecond)
		close(stop)
		wg.Wait()
	})

	t.Run("ticks plain", func(t *testing.T) {
		orig := isFancyTerminal
		isFancyTerminal = func() bool { return false }
		defer func() { isFancyTerminal = orig }()

		done := make(chan struct{})
		stop := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			spinUntil(done, stop)
		}()

		time.Sleep(200 * time.Millisecond)
		close(stop)
		wg.Wait()
	})
}

func TestRun(t *testing.T) {
	t.Run("file change", func(t *testing.T) {
		dir := t.TempDir()
		goFile := dir + "/main.go"
		if err := os.WriteFile(goFile, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}

		task := runner.New("go", "version")
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go run(dir, task, stop, &wg, 100*time.Millisecond, nil, false)

		// Wait for initial run to set lastMod, then modify the file
		time.Sleep(50 * time.Millisecond)
		if err := os.WriteFile(goFile, []byte("package main\n// modified\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Wait for rerun to fire
		time.Sleep(300 * time.Millisecond)
		close(stop)
		wg.Wait()
	})

	t.Run("file change clear screen", func(t *testing.T) {
		dir := t.TempDir()
		goFile := dir + "/main.go"
		if err := os.WriteFile(goFile, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}

		task := runner.New("go", "version")
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go run(dir, task, stop, &wg, 100*time.Millisecond, nil, true)

		time.Sleep(50 * time.Millisecond)
		if err := os.WriteFile(goFile, []byte("package main\n// modified\n"), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(300 * time.Millisecond)
		close(stop)
		wg.Wait()
	})

	t.Run("command error", func(t *testing.T) {
		dir := t.TempDir()
		task := runner.New("nonexistent-command-xyz-abc")
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go run(dir, task, stop, &wg, 100*time.Millisecond, nil, false)

		time.Sleep(100 * time.Millisecond)
		close(stop)
		wg.Wait()
	})

	t.Run("spinner active on rerun", func(t *testing.T) {
		dir := t.TempDir()
		goFile := dir + "/main.go"
		if err := os.WriteFile(goFile, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}

		task := runner.New("sleep", "10")
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go run(dir, task, stop, &wg, 100*time.Millisecond, nil, false)

		// Wait for the command to start and spinner to begin
		time.Sleep(50 * time.Millisecond)
		// Modify file to trigger rerun while spinner is active
		if err := os.WriteFile(goFile, []byte("package main\n// modified\n"), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(300 * time.Millisecond)
		close(stop)
		wg.Wait()
	})

	t.Run("spinner active on stop", func(t *testing.T) {
		dir := t.TempDir()
		task := runner.New("sleep", "10")
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go run(dir, task, stop, &wg, 100*time.Millisecond, nil, false)

		// Close stop quickly before the sleep command exits
		time.Sleep(50 * time.Millisecond)
		close(stop)
		wg.Wait()
	})

	t.Run("kill error", func(t *testing.T) {
		dir := t.TempDir()
		task := newMockRunner(nil, errors.New("mock kill error"))
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go run(dir, task, stop, &wg, 100*time.Millisecond, nil, false)

		time.Sleep(50 * time.Millisecond)
		close(stop)
		wg.Wait()
	})

	t.Run("stop with spinner already stopped", func(t *testing.T) {
		dir := t.TempDir()
		goFile := dir + "/main.go"
		if err := os.WriteFile(goFile, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// mockRunner with done already closed so spinner exits immediately
		task := newMockRunner(nil, nil)
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go run(dir, task, stop, &wg, 50*time.Millisecond, nil, false)

		// Modify file to trigger rerun (startRun closes old stopSpin, creates new)
		time.Sleep(30 * time.Millisecond)
		if err := os.WriteFile(goFile, []byte("package main\n// mod\n"), 0644); err != nil {
			t.Fatal(err)
		}
		// Give time for rerun to happen, then close stop
		time.Sleep(100 * time.Millisecond)
		close(stop)
		wg.Wait()
	})
}
