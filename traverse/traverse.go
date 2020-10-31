package traverse

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

var errHasModify = errors.New("rerun immediately: stop walk because has to modify")

func walkFunc(lastMod time.Time) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		if filepath.Base(path) == ".git" && fi.IsDir() {
			return filepath.SkipDir
		}

		if isHiddenFile(path) {
			return nil
		}

		if fi.ModTime().After(lastMod) {
			return errHasModify
		}

		return nil
	}
}

func isHiddenFile(path string) bool {
	return filepath.Base(path)[0] == '.'
}

// IsModify check if has file an update or not
func IsModify(dir string, lastMod time.Time) bool {
	err := filepath.Walk(dir, walkFunc(lastMod))
	return err == errHasModify
}
