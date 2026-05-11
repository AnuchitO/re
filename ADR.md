# Architecture Decision Records

---

## ADR-001: Polling over OS file system events

**Status:** Accepted

**Context:**
File watchers typically use OS-native event APIs — inotify on Linux, FSEvents/kqueue on macOS, ReadDirectoryChangesW on Windows. These push events immediately when a file changes rather than requiring periodic checks.

**Decision:**
Use `filepath.WalkDir` on a polling interval instead of OS events.

**Reasons:**
- OS event APIs do not watch recursively — you must manually add every new subdirectory as it appears, which requires handling `Create` events on directories.
- Editors write files in bursts (temp file → rename, or multiple saves in rapid succession), requiring event debouncing logic.
- Platform differences require significant per-OS code paths or a cross-platform library with its own edge cases.
- For a typical Go project (hundreds of files), polling at 800ms is fast enough that the latency difference is not noticeable in practice.

**Consequences:**
- Maximum detection latency equals the poll interval (default 800ms, configurable).
- CPU overhead scales with directory size, but is negligible for typical projects.
- Zero platform-specific code in the watch path.

---

## ADR-002: WalkDir over Walk

**Status:** Accepted

**Context:**
The original implementation used `filepath.Walk`, which calls `os.Lstat` on every directory entry regardless of whether the entry will be used.

**Decision:**
Switch to `filepath.WalkDir` and defer `d.Info()` (the stat syscall) until after all cheap checks pass.

**Reasons:**
- `filepath.WalkDir` reads directory entries in bulk from the OS via `Getdents`-style calls. `IsDir()` and `Type()` are available without a syscall.
- Hidden files, `.git`, and `.gitignore`-matched entries are filtered using only `d.IsDir()` — no stat needed.
- `d.Info()` is called only for entries that pass all filters, i.e. the files we actually care about.

**Consequences:**
- ~22x speedup on the raw walk, ~3x on the full `IsModify` check vs the original `Walk` implementation.
- 94% fewer allocations per poll.
- Requires Go 1.16+ (`fs.WalkDirFunc`).

---

## ADR-003: Gitignore pattern caching

**Status:** Accepted

**Context:**
The original `IsModify` read `.gitignore` on every poll call — opening, scanning, and closing the file every 800ms. Benchmarks showed this accounted for 91% of total poll time (~5200 ns out of ~5700 ns per call).

**Decision:**
Move `.gitignore` reading out of `IsModify` into the caller. The watcher caches patterns and invalidates the cache only when `.gitignore`'s `ModTime` changes.

**Reasons:**
- `.gitignore` is rarely edited during a work session.
- A single `os.Stat` call per poll (~200 ns) is far cheaper than opening and scanning the file (~5000 ns).
- Keeping `IsModify` free of I/O makes it easier to test and benchmark in isolation.

**Consequences:**
- `IsModify` now takes `[]string` instead of reading patterns internally — callers are responsible for providing them.
- Poll path reduces to ~509 ns/op, identical to the raw directory walk cost.
- If `.gitignore` is edited, changes are picked up on the next poll automatically.

---

## ADR-004: Process group killing

**Status:** Accepted

**Context:**
When a file changes, the previous run must be stopped before starting a new one. Sending a signal only to the top-level process leaves child processes (e.g. test binaries spawned by `go test ./...`) running as orphans.

**Decision:**
Send `SIGINT` to the entire process group using a negative PID (`syscall.Kill(-pid, SIGINT)`). If the process group does not exit within 3 seconds, send `SIGKILL`. Track process lifetime with a `done` channel closed in a `Wait` goroutine started by `Start`.

**Reasons:**
- Negative PID targets the process group, ensuring all child processes receive the signal.
- `SIGINT` gives processes a chance to clean up (flush buffers, print summaries) before `SIGKILL` forces termination.
- The `done` channel lets `KillCommand` detect that the process already exited naturally — skipping the signal entirely and avoiding "operation not permitted" errors when the previous run finished on its own.

**Consequences:**
- On Unix, processes must be started with `Setpgid: true` so they lead their own process group.
- Windows uses `TASKKILL /T /F` instead, which achieves the same tree-kill behaviour.
- The 3-second grace period adds latency only in the rare case where a process ignores `SIGINT`.

---

## ADR-005: Terminal-aware output

**Status:** Accepted

**Context:**
ANSI escape codes and Unicode characters improve readability in modern terminals but produce garbage in CI logs, piped output, and dumb terminals.

**Decision:**
Detect terminal capabilities at runtime and fall back to plain ASCII when any of the following are true:
1. `$NO_COLOR` is set (https://no-color.org)
2. `$TERM` equals `dumb`
3. stdout is not a TTY (`golang.org/x/term.IsTerminal`)

**Reasons:**
- `$NO_COLOR` is a widely adopted convention for user opt-out.
- `$TERM=dumb` is set by many CI systems and editors that embed terminals.
- TTY check catches pipes and redirects (`re ... | tee log`).
- All three checks together cover the vast majority of non-interactive environments.

**Consequences:**
- Adds `golang.org/x/term` as a dependency (thin wrapper around OS TTY detection).
- The spinner erases its line before the first byte of command output via a shared `sync.Once` across stdout and stderr wrappers, preventing interleaving regardless of which stream writes first.
- Plain ASCII fallback (`|/-\`, `-`, `rerun`) carries the same information as the fancy version.
