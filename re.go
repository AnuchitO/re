package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/AnuchitO/re/runner"
)

func splitCommand(args []string) (prog string, params []string, err error) {
	if len(args) < 1 {
		err = errors.New("you should add command after re [command], e.g. 're go test -v .'")
		return
	}

	prog = args[0]
	params = args[1:]
	return
}

var version = "dev"

var getwd = os.Getwd

func main() {
	interval := flag.Duration("interval", 800*time.Millisecond, "polling interval for file changes")
	ignore := flag.String("ignore", "", "comma-separated file patterns to ignore (e.g. '*.log,vendor')")
	clear := flag.Bool("clear", false, "clear the screen before each rerun")
	ver := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		return
	}

	prog, params, err := splitCommand(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	dir, err := getwd()
	if err != nil {
		log.Fatal(err)
	}

	var ignorePatterns []string
	if *ignore != "" {
		ignorePatterns = strings.Split(*ignore, ",")
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	task := runner.New(prog, params...)

	go run(dir, task, stop, &wg, *interval, ignorePatterns, *clear)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig
	close(stop)

	wg.Wait()
	fmt.Println("process terminated")
}
