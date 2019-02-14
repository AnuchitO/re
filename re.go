package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	taskRunner := runner.NewRunner(prog, params...)
	taskRunner.Run()
	startTime := time.Now()
	for {
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
			fmt.Printf("\n============ Rerun ============\n\n")
			taskRunner.Run()
		}

		time.Sleep(800 * time.Millisecond)
	}
}
