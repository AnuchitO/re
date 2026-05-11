//go:build !windows
// +build !windows

package main

import "fmt"

// clearScreen clears the terminal using ANSI escape codes.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
