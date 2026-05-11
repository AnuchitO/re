package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFancyTerminal(t *testing.T) {
	t.Run("no color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		assert.False(t, isFancyTerminal())
	})

	t.Run("dumb term", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		t.Setenv("TERM", "dumb")
		assert.False(t, isFancyTerminal())
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		t.Setenv("TERM", "xterm-256color")
		// Test process has no TTY, so this returns false
		assert.False(t, isFancyTerminal())
	})
}
