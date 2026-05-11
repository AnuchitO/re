package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AnuchitO/re/runner"
	"github.com/AnuchitO/re/traverse"
)

func printRerunHeader() {
	ts := time.Now().Format("15:04:05")
	if isFancyTerminal() {
		fmt.Print("\n\033[2m────────────────────────────────────────\033[0m\n")
		fmt.Printf("          \033[1m↺  rerun · %s\033[0m\n", ts)
		fmt.Print("\033[2m────────────────────────────────────────\033[0m\n\n")
	} else {
		fmt.Print("\n----------------------------------------\n")
		fmt.Printf("     rerun · %s\n", ts)
		fmt.Print("----------------------------------------\n\n")
	}
}

var spinnerFramesFancy = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
var spinnerFramesPlain = []string{"|", "/", "-", "\\"}
var dotFrames = []string{"", ".", "..", "..."}

// spinUntil shows a spinner with animated dots until done or stop is closed.
func spinUntil(done <-chan struct{}, stop <-chan struct{}) {
	frames := spinnerFramesFancy
	if !isFancyTerminal() {
		frames = spinnerFramesPlain
	}
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r\033[K")
			return
		case <-stop:
			fmt.Print("\r\033[K")
			return
		case <-ticker.C:
			fmt.Printf("\r %s running%-3s", frames[i%len(frames)], dotFrames[i%len(dotFrames)])
			i++
		}
	}
}

// eraseOnFirstWrite erases the spinner line before the first byte of command
// output reaches the terminal. once is shared between stdout and stderr so
// only one erase happens regardless of which stream writes first.
type eraseOnFirstWrite struct {
	w    io.Writer
	once *sync.Once
}

func (e *eraseOnFirstWrite) Write(p []byte) (n int, err error) {
	e.once.Do(func() { fmt.Print("\r\033[K") })
	return e.w.Write(p)
}

// patternCache holds .gitignore patterns and re-reads the file only when
// its ModTime changes, avoiding file I/O on every poll.
type patternCache struct {
	modTime  time.Time
	patterns []string
}

func (c *patternCache) get(dir string, extra []string) []string {
	fi, err := os.Stat(filepath.Join(dir, ".gitignore"))
	mod := time.Time{}
	if err == nil {
		mod = fi.ModTime()
	}
	if !mod.Equal(c.modTime) {
		c.modTime = mod
		c.patterns = append(traverse.ReadGitignore(dir), extra...)
	}
	return c.patterns
}

func run(dir string, task *runner.Runner, stop chan struct{}, wg *sync.WaitGroup, interval time.Duration, ignorePatterns []string, clear bool) {
	defer wg.Done()
	lastMod := time.Now()

	cache := &patternCache{}
	patterns := cache.get(dir, ignorePatterns)

	// stopSpin is closed to stop the current spinner goroutine.
	stopSpin := make(chan struct{})
	close(stopSpin) // nothing spinning yet

	startRun := func() {
		// Stop the previous spinner before starting the next run.
		select {
		case <-stopSpin:
		default:
			close(stopSpin)
		}
		stopSpin = make(chan struct{})

		// Erase the spinner line on the first byte of output from either stream.
		once := &sync.Once{}
		task.SetStdout(&eraseOnFirstWrite{w: os.Stdout, once: once})
		task.SetStderr(&eraseOnFirstWrite{w: os.Stderr, once: once})

		if err := task.Run(); err != nil {
			log.Printf("command error: %v", err)
		}
		go spinUntil(task.Done(), stopSpin)
	}

	startRun()

	for {
		select {
		case <-stop:
			select {
			case <-stopSpin:
			default:
				close(stopSpin)
			}
			if err := task.KillCommand(); err != nil {
				log.Printf("kill error: %v", err)
			}
			return
		default:
		}

		patterns = cache.get(dir, ignorePatterns)

		if traverse.IsModify(dir, lastMod, patterns) {
			lastMod = time.Now()
			if clear {
				clearScreen()
			} else {
				printRerunHeader()
			}
			startRun()
		}

		time.Sleep(interval)
	}
}
