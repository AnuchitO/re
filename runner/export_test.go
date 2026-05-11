//go:build !windows
// +build !windows

package runner

import "syscall"

// SetSyscallKill replaces syscallKill for testing. Returns a restore function.
func SetSyscallKill(fn func(pid int, sig syscall.Signal) error) func() {
	orig := syscallKill
	syscallKill = fn
	return func() { syscallKill = orig }
}
