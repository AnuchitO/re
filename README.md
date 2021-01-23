
# re

`re` is a program for rerun your command when files change with a focus on simplicity.

- `re` is easy to type.
- NO dependency use only standard library of go to implement.
- zero config.

# Installation
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
* [ ] configurable option
* [ ] skip file in .gitignore
* [ ] should clear screen before rerun command
* [ ] Cross platform

## Contribute
- please send a PR.

## License
[MIT](https://github.com/AnuchitO/re/blob/master/LICENSE)
