package traverse

import (
	"testing"
	"time"
)

// BenchmarkIsModify_NoChange benchmarks the common case: no files changed.
// The entire tree must be walked to confirm nothing is newer than lastMod.
func BenchmarkIsModify_NoChange(b *testing.B) {
	dir := ".."
	lastMod := time.Now()
	patterns := ReadGitignore(dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsModify(dir, lastMod, patterns)
	}
}

// BenchmarkIsModify_Change benchmarks the early-exit case: first file is newer.
// Walk should stop at the first match.
func BenchmarkIsModify_Change(b *testing.B) {
	dir := ".."
	lastMod := time.Time{} // zero — every file is "newer"
	patterns := ReadGitignore(dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsModify(dir, lastMod, patterns)
	}
}

// BenchmarkWalkFunc_NoChange benchmarks walkFunc directly, no gitignore I/O.
func BenchmarkWalkFunc_NoChange(b *testing.B) {
	dir := ".."
	lastMod := time.Now()
	patterns := []string{"re", "*.log", "coverage.out"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = walk(dir, lastMod, patterns)
	}
}
