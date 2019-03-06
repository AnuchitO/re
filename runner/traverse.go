package runner

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Traverse(dir string, lastMod time.Time) time.Time {
	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if path == ".git" && fi.IsDir() {
			log.Println("skipping .git directory")
			return filepath.SkipDir
		}

		// ignore hidden files
		if filepath.Base(path)[0] == '.' {
			return nil
		}

		if fi.ModTime().After(lastMod) {
			lastMod = time.Now()
			return errors.New("reload immediately: stop walking")
		}

		return nil
	})

	return lastMod
}
