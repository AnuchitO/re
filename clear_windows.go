//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
)

// clearScreen clears the terminal by running the Windows cls command.
func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run() //nolint:errcheck
}
