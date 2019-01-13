package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

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
			log.Fatal("can't open file", err)
		}
		mod := info.ModTime().Unix()

		if lastmod < mod {
			lastmod = mod

			fmt.Println("\n\nrerun")
			cmd := exec.Command(prog, params...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

		time.Sleep(800 * time.Millisecond)
	}
}
