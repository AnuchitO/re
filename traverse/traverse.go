package traverse

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

var hasModify = errors.New("rerun immediately: stop walk because has to modify")

func walkFunc(lastMod time.Time) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		if filepath.Base(path) == ".git" && fi.IsDir() {
			return filepath.SkipDir
		}

		// ignore hidden files
		if filepath.Base(path)[0] == '.' {
			return nil
		}

		if fi.ModTime().After(lastMod) {
			return hasModify
		}

		return nil
	}
}

// IsModify check if has file an update or not
func IsModify(dir string, lastMod time.Time) bool {
	err := filepath.Walk(dir, walkFunc(lastMod))
	if err == hasModify {
		return true
	}

	return false
}
