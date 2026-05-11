//go:build !windows
// +build !windows

package main

import "testing"

func TestClearScreen(t *testing.T) {
	// Just verify clearScreen doesn't panic; it prints ANSI codes to stdout.
	clearScreen()
}
