package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func run(prog string, params ...string) {
	cmd := exec.Command(prog, params...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func lstat(path string, files map[string]int64) int64 {
	info, err := os.Lstat(path)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			delete(files, path)
			return 0
		}
		log.Fatal("can't get file info.", err)
	}
	mod := info.ModTime().Unix()
	return mod
}

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
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args

	// TODO: handle index out of range
	prog := args[1]
	params := args[2:]

	files := map[string]int64{}
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if !strings.Contains(path, "/.git") {
			files[path] = 0
		}
		return nil
	})
	if err != nil {
		log.Fatal("walk through file error", err)
	}

	for {
		hasChanged := false
		for path, lastmod := range files {
			mod := lstat(path, files)
			if lastmod < mod {
				files[path] = mod
				hasChanged = true
			}
		}

		if hasChanged {
			fmt.Println("\nrerun")
			run(prog, params...)
		}
		time.Sleep(800 * time.Millisecond)
	}
}
