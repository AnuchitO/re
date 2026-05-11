package traverse

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var errHasModify = errors.New("rerun immediately: stop walk because has to modify")

func walkFunc(root string, lastMod time.Time, ignorePatterns []string) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		base := filepath.Base(path)

		if base == ".git" && fi.IsDir() {
			return filepath.SkipDir
		}

		if isHiddenFile(base) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Never apply ignore patterns to the root itself — its base name may
		// coincidentally match a pattern (e.g. a project named "re" and a
		// .gitignore entry "re" for the binary), which would SkipDir the
		// entire walk and make IsModify always return false.
		if path != root {
			for _, pattern := range ignorePatterns {
				matched, _ := filepath.Match(pattern, base)
				if matched {
					if fi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		if fi.ModTime().After(lastMod) {
			return errHasModify
		}

		return nil
	}
}

func isHiddenFile(name string) bool {
	return name != "." && strings.HasPrefix(name, ".")
}

// readGitignore returns simple glob patterns from .gitignore in the given directory.
// Lines starting with # and empty lines are ignored.
// Note: only simple glob patterns (e.g. *.log, vendor) are supported.
func readGitignore(dir string) []string {
	f, err := os.Open(filepath.Join(dir, ".gitignore"))
	if err != nil {
		return nil
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("close .gitignore: %v", err)
		}
	}()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip leading slash so "filepath.Match" can match by base name
		patterns = append(patterns, strings.TrimPrefix(line, "/"))
	}
	return patterns
}

// IsModify checks if any file in dir has been modified after lastMod.
// It automatically reads .gitignore patterns and also accepts additional
// ignore patterns via extraIgnore (supports filepath.Match glob syntax).
func IsModify(dir string, lastMod time.Time, extraIgnore ...string) bool {
	patterns := append(readGitignore(dir), extraIgnore...)
	err := filepath.Walk(dir, walkFunc(dir, lastMod, patterns))
	return err == errHasModify
}
