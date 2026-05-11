package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AnuchitO/re/runner"
	"github.com/AnuchitO/re/traverse"
)

func run(dir string, task *runner.Runner, stop chan struct{}, wg *sync.WaitGroup, interval time.Duration, ignorePatterns []string, clear bool) {
	defer wg.Done()
	lastMod := time.Now()

	if err := task.Run(); err != nil {
		log.Printf("command error: %v", err)
	}

	for {
		select {
		case <-stop:
			if err := task.KillCommand(); err != nil {
				log.Printf("kill error: %v", err)
			}
			return
		default:
		}

		if traverse.IsModify(dir, lastMod, ignorePatterns...) {
			lastMod = time.Now()
			if clear {
				clearScreen()
			} else {
				fmt.Printf("\n----------------- Rerun ------------------\n\n")
			}
			if err := task.Run(); err != nil {
				log.Printf("command error: %v", err)
			}
		}

		time.Sleep(interval)
	}
}
