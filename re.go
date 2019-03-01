package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AnuchitO/re/runner"
)

var (
	SkipFile = errors.New("re: skip file")
	StopWalk = errors.New("re: stop walk")
)

func splitCommand(args []string) (prog string, params []string, err error) {
	if len(args) < 2 {
		err = errors.New("you should add command after re [command], e.g. 're go test -v .'")
		return
	}

	prog = args[1]
	params = args[2:]
	return
}

func main() {
	prog, params, err := splitCommand(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// doneChannel := make(chan struct{})
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	taskRunner := runner.NewRunner(prog, params...)
	go run(dir, taskRunner, stop, &wg)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	<-sig
	close(stop)
	err = taskRunner.KillCommand()
	if err != nil {
		fmt.Println(err)
	}
	wg.Wait()
	fmt.Println("process terminated")
}

func run(dir string, taskRunner *runner.Runner, stop chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := taskRunner.Run(); err != nil {
		fmt.Println(err)
	}

	var (
		// changes notify when file modification time got changes.
		changes = make(chan struct{})
		// quit use for notify walk goroutine to stop walking.
		quit = make(chan struct{})
	)

	wg.Add(1)
	go func(dir string, changes chan<- struct{}, quit <-chan struct{}, wg *sync.WaitGroup) {
		defer wg.Done()

		now := time.Now()
		for {
			select {
			case <-quit:
				return
			default:
				filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if err := detectSkipFile(info); err != nil {
						if err == filepath.SkipDir {
							return err
						}
						return nil
					}

					if info.ModTime().After(now) {
						changes <- struct{}{}
						return StopWalk
					}
					return nil
				})
				now = time.Now()
				time.Sleep(800 * time.Millisecond)
			}
		}
	}(dir, changes, quit, wg)

	for {
		select {
		case <-stop:
			// tell file walk goroutine to stop execution.
			quit <- struct{}{}
			return
		case <-changes:
			if err := taskRunner.Run(); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("\n============ Rerun ============\n")
			}
		}
	}
}

// detectSkipFile detect file and directory that want to ignore.
func detectSkipFile(info os.FileInfo) error {
	if strings.HasPrefix(info.Name(), ".") {
		if info.IsDir() {
			return filepath.SkipDir
		}
		return SkipFile
	}
	return nil
}
