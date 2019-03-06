package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunnerWalk(t *testing.T) {
	t.Run("No files change should return last modify time", func(t *testing.T) {
		now := time.Now()
		dir := "."

		mod := Traverse(dir, now)

		assert.True(t, mod.Equal(now), "should return last modify time.")
	})

	t.Run("File chagne", func(t *testing.T) {
		form := "Mon Jan _2 15:04:05 2006"
		lastMod, _ := time.Parse(form, "Sat Feb 08 07:00:00 1992")
		dir := "."

		mod := Traverse(dir, lastMod)

		assert.True(t, mod.After(lastMod), "should return lastest modify time.")
	})
}
