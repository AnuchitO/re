package runner

import (
	"testing"
	"time"
)

func TestRunnerWalk(t *testing.T) {
	t.Run("No files change should return last modify time", func(t *testing.T) {
		task := &Runner{dir: "."}
		now := time.Now()

		mod := task.Walk(now)

		if !mod.Equal(now) {
			t.Errorf("should return last modify time '%s' but got %s", now, mod)
		}
	})

	t.Run("File chagne", func(t *testing.T) {
		task := &Runner{dir: "."}
		form := "Mon Jan _2 15:04:05 2006"
		lastMod, _ := time.Parse(form, "Sat Feb 08 07:00:00 1992")

		mod := task.Walk(lastMod)

		if !mod.After(lastMod) {
			t.Errorf("should return lastest modify time '%s' but got %s", lastMod, mod)
		}
	})
}
