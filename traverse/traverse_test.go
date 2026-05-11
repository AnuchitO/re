package traverse

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunnerWalk(t *testing.T) {
	t.Run("No files change should return last modify time", func(t *testing.T) {
		now := time.Now()
		dir := "."

		mod := IsModify(dir, now)

		assert.False(t, mod, "should return last modify time.")
	})

	t.Run("should return trure when file has change", func(t *testing.T) {
		form := "Mon Jan _2 15:04:05 2006"
		lastMod, _ := time.Parse(form, "Sat Feb 08 07:00:00 1992")
		dir := "."

		mod := IsModify(dir, lastMod)

		assert.True(t, mod, "should return lastest modify time.")
	})
}

type info struct {
	os.FileInfo
}

func (i info) IsDir() bool  { return true }
func (i info) ModTime() time.Time { return time.Time{} } // zero — never "after" a recent lastMod

func TestWalkFunc(t *testing.T) {
	t.Run("should skip .git directory", func(t *testing.T) {
		root := "/user/project"
		walk := walkFunc(root, time.Now(), nil)

		fi := info{}

		err := walk("/user/project/.git", fi, nil) //nolint:errcheck

		assert.Equal(t, filepath.SkipDir, err, "should Skip directory .git but it not.")
	})

	t.Run("should not skip root even when its name matches an ignore pattern", func(t *testing.T) {
		// Simulates a project directory named "re" with a .gitignore entry "re".
		// The root must never be SkipDir'd or the entire walk is skipped.
		root := "/user/project/re"
		// Use time.Now() so the root's zero ModTime is not "after" lastMod,
		// meaning walkFunc returns nil (no match, no skip) — not SkipDir.
		walk := walkFunc(root, time.Now(), []string{"re"})

		fi := info{} // IsDir() = true

		err := walk(root, fi, nil)

		assert.NoError(t, err, "root directory must not be skipped by an ignore pattern")
	})
}
