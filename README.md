# re

Reruns a command automatically whenever files change. Built for tight feedback loops like TDD.

- No configuration required
- No runtime dependencies — standard library only
- Short to type: just `re`

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
re go test -v .
```

Restart an API server on every file change:
```sh
re go run main.go
```

Clear the screen before each rerun, poll every 300ms, and ignore log files:
```sh
re -clear -interval 300ms -ignore "*.log,vendor" go test ./...
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `-interval` | `800ms` | How often to check for file changes |
| `-ignore` | — | Comma-separated glob patterns to skip (e.g. `*.log,vendor`) |
| `-clear` | `false` | Clear the terminal before each rerun |

> `.gitignore` patterns are respected automatically — no extra configuration needed.

## Features

- [x] Rerun any command on file change
- [x] Interrupt running process and restart cleanly
- [x] Watch nested directories and single files
- [x] Respect `.gitignore`
- [x] Clear screen before rerun
- [x] Configurable via flags
- [ ] Full cross-platform CI coverage

## Contributing

Pull requests are welcome.

## License

[MIT](https://github.com/AnuchitO/re/blob/master/LICENSE)
