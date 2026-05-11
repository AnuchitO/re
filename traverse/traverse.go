package traverse

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var errHasModify = errors.New("rerun immediately: stop walk because has to modify")

var openFile = func(name string) (io.ReadCloser, error) { return os.Open(name) }

func walkFunc(root string, lastMod time.Time, ignorePatterns []string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		base := filepath.Base(path)

		// All checks below use d.IsDir() / d.Type() — no syscall needed.
		if base == ".git" && d.IsDir() {
			return filepath.SkipDir
		}

		if isHiddenFile(base) {
			if d.IsDir() {
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
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		// Only call Info (stat syscall) after all cheap checks pass.
		fi, err := d.Info()
		if err != nil {
			return nil
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

// ReadGitignore returns simple glob patterns from .gitignore in the given directory.
// Lines starting with # and empty lines are ignored.
// Note: only simple glob patterns (e.g. *.log, vendor) are supported.
func ReadGitignore(dir string) []string {
	f, err := openFile(filepath.Join(dir, ".gitignore"))
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

// walk is the inner walk used by IsModify, exposed for benchmarking.
func walk(dir string, lastMod time.Time, patterns []string) error {
	return filepath.WalkDir(dir, walkFunc(dir, lastMod, patterns))
}

// IsModify checks if any file in dir has been modified after lastMod.
// patterns should be pre-built by the caller (e.g. via ReadGitignore) and
// cached across calls to avoid repeated file I/O on every poll.
func IsModify(dir string, lastMod time.Time, patterns []string) bool {
	return walk(dir, lastMod, patterns) == errHasModify
}
