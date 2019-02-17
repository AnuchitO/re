package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/AnuchitO/re/runner"
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
	startTime := time.Now()
	err := taskRunner.Run()
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case <-stop:
			return
		default:
		}
		hasChanged := false
		filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if path == ".git" && fi.IsDir() {
				log.Println("skipping .git directory")
				return filepath.SkipDir
			}

			// ignore hidden files
			if filepath.Base(path)[0] == '.' {
				return nil
			}

			if fi.ModTime().After(startTime) {
				hasChanged = true
				startTime = time.Now()
				return errors.New("reload immediately: stop walking")
			}

			return nil
		})

		if hasChanged {
			err := taskRunner.Run()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("\n============ Rerun ============\n\n")
			}
		}

		time.Sleep(800 * time.Millisecond)
	}
}
