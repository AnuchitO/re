
# re

`re` is a program for rerun your command when files change with a focus on simplicity.

- `re` is easy to type.
- NO dependency use only standard library of go to implement.
- zero config.

# Installation
go version =>1.11
```
go install github.com/AnuchitO/re
```

go version <1.11
```
go get github.com/AnuchitO/re
```

# Usage
re is easy to use just type `re` and follow with your command.
```
re [command]
```
note: make sure you set $GOPATH/bin into your class path. (.zshrc, .bashrc, etc.)

# examples
## rerun command e.g. `go test` - watching file to re run your command again whenever file changed.
```
re go test -v .
```

![go tdd](../assets/example_tdd_i.gif?raw=true)

## rerun api service
- TODO

## Feature
* [x] rerun one time command e.g. `go test`
* [x] interrupt and rerun
* [x] Watching nested files
* [x] Watching single files
* [x] reload app when file change
* [x] configurable option
* [x] skip file in .gitignore
* [x] should clear screen before rerun command
* [ ] Cross platform

## Flags
| Flag | Default | Description |
|---|---|---|
| `-interval` | `800ms` | Polling interval for file changes |
| `-ignore` | `""` | Comma-separated glob patterns to skip (e.g. `*.log,vendor`) |
| `-clear` | `false` | Clear the screen before each rerun |

```
re -clear -interval 500ms go test -v .
re -ignore "*.log,vendor" go test ./...
```

## Contribute
- please send a PR.

## License
[MIT](https://github.com/AnuchitO/re/blob/master/LICENSE)
