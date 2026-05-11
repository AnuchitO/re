# Performance Optimization Guide — `traverse` package

This document explains the two optimizations applied to `traverse/traverse.go`,
what caused the cost, what changed, and why performance improved.
Results are recorded in `bench.log`.

---

## Baseline (Before Any Optimization)

```
BenchmarkIsModify_NoChange  ~17,000 ns/op   7280 B/op   53 allocs/op
BenchmarkIsModify_Change    ~17,200 ns/op   7280 B/op   53 allocs/op
BenchmarkWalkFunc_NoChange  ~11,800 ns/op   2688 B/op   33 allocs/op
```

The entry point was:

```go
// traverse/traverse.go (original)

func walkFunc(root string, lastMod time.Time, ignorePatterns []string) filepath.WalkFunc {
    return func(path string, fi os.FileInfo, err error) error {
        // ...
        if base == ".git" && fi.IsDir() { ... }
        if isHiddenFile(base)           { ... }
        for _, pattern := range ignorePatterns { ... }

        if fi.ModTime().After(lastMod) {
            return errHasModify
        }
        return nil
    }
}

func IsModify(dir string, lastMod time.Time, extraIgnore ...string) bool {
    patterns := append(readGitignore(dir), extraIgnore...)
    err := filepath.Walk(dir, walkFunc(dir, lastMod, patterns))
    return err == errHasModify
}
```

---

## Optimization 1 — Replace `filepath.Walk` with `filepath.WalkDir`

**Commit:** `fe95432`
**File:** `traverse/traverse.go`

### What was costing CPU and memory?

`filepath.Walk` calls `os.Lstat` on **every** entry before handing it to the
callback.  That means even files you are about to skip (hidden files, ignored
patterns, `.git/`) first pay a **syscall** and an **`os.FileInfo` allocation**.

```
filepath.Walk  →  for each entry:
    1. os.Lstat(path)        ← syscall + alloc (os.FileInfo)
    2. call WalkFunc(path, fileInfo, err)
```

With 53 allocations for a small tree, most of those were `os.FileInfo` structs
allocated before the callback even had a chance to skip the entry.

### What changed?

**Signature change** — `filepath.WalkFunc` → `fs.WalkDirFunc`:

```go
// BEFORE
func walkFunc(...) filepath.WalkFunc {
    return func(path string, fi os.FileInfo, err error) error {
        if base == ".git" && fi.IsDir() { ... }
        if isHiddenFile(base) {
            if fi.IsDir() { ... }
        }
        // patterns checked here...

        if fi.ModTime().After(lastMod) {  // fi already exists, no extra call
            return errHasModify
        }
        return nil
    }
}
```

```go
// AFTER
func walkFunc(...) fs.WalkDirFunc {
    return func(path string, d fs.DirEntry, err error) error {
        // All checks below use d.IsDir() — no syscall needed.
        if base == ".git" && d.IsDir() { ... }
        if isHiddenFile(base) {
            if d.IsDir() { ... }
        }
        // patterns checked here...

        // Only call Info (stat syscall) AFTER all cheap checks pass.
        fi, err := d.Info()
        if err != nil { return nil }
        if fi.ModTime().After(lastMod) {
            return errHasModify
        }
        return nil
    }
}
```

**Call-site change:**

```go
// BEFORE
err := filepath.Walk(dir, walkFunc(dir, lastMod, patterns))

// AFTER
err := filepath.WalkDir(dir, walkFunc(dir, lastMod, patterns))
```

### Why it improves

`filepath.WalkDir` hands the callback an `fs.DirEntry` which is populated from
the OS directory-read buffer — **no extra syscall**.  `d.IsDir()` is free.
`d.Info()` (which calls `stat`) is only invoked **after** the cheap guards
pass, so skipped entries (`.git`, hidden files, ignored patterns) never pay for
a syscall at all.

| Step | Walk | WalkDir |
|------|------|---------|
| Read dir listing | bulk OS call | bulk OS call |
| Per-entry `stat` | always | only if needed |
| `os.FileInfo` alloc | always | only if needed |

### Result after Optimization 1

```
BenchmarkIsModify_NoChange  ~5,700 ns/op   4816 B/op   23 allocs/op   (~3x faster, 56% fewer allocs)
BenchmarkIsModify_Change    ~5,770 ns/op   4816 B/op   23 allocs/op   (~3x faster, 56% fewer allocs)
BenchmarkWalkFunc_NoChange    ~528 ns/op    227 B/op    3 allocs/op   (~22x faster, 91% fewer allocs)
```

`BenchmarkWalkFunc_NoChange` improves the most (~22x) because it passes
patterns directly and skips gitignore I/O entirely — only the walk itself is
measured, making the stat-elimination effect visible in isolation.

---

## Optimization 2 — Cache gitignore patterns, remove I/O from hot path

**Commit:** `d982612`
**File:** `traverse/traverse.go`

### What was costing CPU and memory?

After Optimization 1, `IsModify` still opened and read `.gitignore` on
**every poll interval**:

```go
// AFTER opt-1, BEFORE opt-2
func IsModify(dir string, lastMod time.Time, extraIgnore ...string) bool {
    patterns := append(readGitignore(dir), extraIgnore...)  // ← file I/O every call
    err := walk(dir, lastMod, patterns)
    return err == errHasModify
}
```

`readGitignore` opens the file, allocates a `bufio.Scanner`, scans all lines,
and builds a `[]string` — every single poll, even though `.gitignore` almost
never changes.  This was the dominant cost in `BenchmarkIsModify_*` after the
WalkDir improvement.

### What changed?

**`readGitignore` exported → `ReadGitignore`; `IsModify` signature changed to
accept a pre-built slice.**

```go
// BEFORE (opt-1 state)
func readGitignore(dir string) []string { ... }   // unexported, called internally

func IsModify(dir string, lastMod time.Time, extraIgnore ...string) bool {
    patterns := append(readGitignore(dir), extraIgnore...)  // file I/O here
    err := walk(dir, lastMod, patterns)
    return err == errHasModify
}
```

```go
// AFTER (opt-2)
func ReadGitignore(dir string) []string { ... }   // exported, called by caller once

func IsModify(dir string, lastMod time.Time, patterns []string) bool {
    return walk(dir, lastMod, patterns) == errHasModify
    // no file I/O — patterns are passed in, already built by caller
}
```

The caller (watcher loop) now owns the caching responsibility:

```go
// caller pseudocode
patterns := traverse.ReadGitignore(dir)   // once at startup
for range ticker.C {
    if traverse.IsModify(dir, lastMod, patterns) {
        // re-read patterns only if .gitignore itself changed
    }
}
```

### Why it improves

Eliminating one `os.Open` + `bufio.Scanner` + line allocations per poll drops
the remaining ~5,700 ns/op on `IsModify` to ~510 ns/op.

| Operation removed from hot path | Cost |
|----------------------------------|------|
| `os.Open(".gitignore")` | syscall + fd alloc |
| `bufio.NewScanner` | heap alloc |
| line-by-line scan + `append` | N string allocs |

All of that now happens once at startup and is reused across every poll.

### Result after Optimization 2

```
BenchmarkIsModify_NoChange  ~513 ns/op   227 B/op   3 allocs/op
BenchmarkIsModify_Change    ~515 ns/op   227 B/op   3 allocs/op
BenchmarkWalkFunc_NoChange  ~523 ns/op   227 B/op   3 allocs/op
```

`IsModify` is now as fast as the raw `walk` benchmark — the gitignore overhead
is gone.

---

## Summary

| Benchmark | Baseline | After opt-1 (WalkDir) | After opt-2 (cache) | Total speedup |
|-----------|----------|-----------------------|---------------------|---------------|
| IsModify_NoChange | 17,000 ns | 5,700 ns | 513 ns | **~33x** |
| IsModify_Change   | 17,200 ns | 5,770 ns | 515 ns | **~33x** |
| WalkFunc_NoChange | 11,800 ns |   528 ns | 523 ns | **~23x** |
| Allocations (IsModify) | 53 | 23 | 3 | **94% fewer** |

### Key lessons

1. **Defer expensive operations.** `filepath.Walk` called `stat` on every entry
   before your code could decide to skip it. `filepath.WalkDir` lets cheap
   guards run first.

2. **Move I/O out of the hot loop.** Reading `.gitignore` on every poll was
   wasted work. If the input rarely changes, read it once and pass the result
   in — let the caller decide when to refresh.

3. **Measure with benchmarks before and after.** The ~22x improvement on the
   raw walk and the ~33x on `IsModify` would not be obvious without isolating
   each benchmark case.
