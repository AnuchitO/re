package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/AnuchitO/re/runner"
	"github.com/AnuchitO/re/traverse"
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

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	task := runner.New(prog, params...)

	go run(dir, task, stop, &wg)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig
	close(stop)

	wg.Wait()
	fmt.Println("process terminated")
}

func run(dir string, task *runner.Runner, stop chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	lastMod := time.Now()

	err := task.Run()
	if err != nil {
		fmt.Println(err)
	}

	for {

		select {
		case <-stop:
			err = task.KillCommand()
			if err != nil {
				fmt.Println(err)
			}
			return
		default:
		}

		mod := traverse.IsModify(dir, lastMod)

		if mod {
			lastMod = time.Now()
			err := task.Run()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("\n----------------- Rerun ------------------\n\n")
			}
		}

		time.Sleep(800 * time.Millisecond)
	}
}
