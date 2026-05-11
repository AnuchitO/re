# re

Reruns a command automatically whenever files change. Built for tight feedback loops like TDD.

```
re go test -v ./...
```

## Installation

```sh
go install github.com/AnuchitO/re@latest
```

> Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`.

## Usage

```
re [flags] <command> [args...]
```

### Examples

Run tests on every file change:
```sh
re go test -v ./...
```

Restart an API server on every change:
```sh
re go run main.go
```

Clear the screen, poll every 300ms, ignore log files:
```sh
re -clear -interval 300ms -ignore "*.log,vendor" go test ./...
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `-interval` | `800ms` | How often to poll for file changes |
| `-ignore` | — | Comma-separated glob patterns to skip (e.g. `*.log,vendor`) |
| `-clear` | `false` | Clear the terminal before each rerun |
| `-version` | — | Print version and exit |

> `.gitignore` patterns are respected automatically — no extra configuration needed.

## Design

Key decisions — polling vs OS events, WalkDir, gitignore caching, process group killing, terminal detection — are documented with full context and reasoning in [ADR.md](ADR.md).

## Features

- [x] Rerun any command on file change
- [x] Kill entire process group on rerun (not just the top-level process)
- [x] Respect `.gitignore` with cached pattern matching
- [x] Watch nested directories
- [x] Clear screen before rerun (`-clear`)
- [x] Configurable poll interval and ignore patterns
- [x] Terminal-aware output with plain ASCII fallback
- [x] Version flag (`-version`)
- [ ] Full cross-platform CI coverage

## Contributing

Pull requests are welcome.

## License

[MIT](https://github.com/AnuchitO/re/blob/master/LICENSE)
