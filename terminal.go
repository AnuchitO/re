package main

import (
	"os"

	"golang.org/x/term"
)

// isFancyTerminal reports whether stdout supports ANSI escape codes and Unicode.
// Returns false when output is piped/redirected, $TERM=dumb, or $NO_COLOR is set.
func isFancyTerminal() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}
