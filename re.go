package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func run(prog string, params ...string) {
	cmd := exec.Command(prog, params...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var lastmod int64
	args := os.Args
	prog := args[1]
	params := args[2:]

	for {
		info, err := os.Lstat(dir)
		if err != nil {
			log.Fatal("can't get file info.", err)
		}
		mod := info.ModTime().Unix()

		if lastmod < mod {
			lastmod = mod

			fmt.Println("\n\nrerunx")
			run(prog, params...)
		}

		time.Sleep(800 * time.Millisecond)
	}
}
