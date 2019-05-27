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

func (i info) IsDir() bool {
	return true
}

func TestWalkFunc(t *testing.T) {
	t.Run("should skip .git directory", func(t *testing.T) {
		walk := walkFunc(time.Now())

		fi := info{}

		err := walk("/user/project/.git", fi, nil)

		assert.Equal(t, filepath.SkipDir, err, "should Skip directory .git but it not.")
	})
}
