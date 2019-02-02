# re

`re` is a program for rerun your command when files change with a focus on simplicity.

- `re` is easy to type.
- test your test command w
- zero config is
# Installation
```
go install github.com/AnuchitO/re
```

# Usage
re is easy to use just type `re` and follow with your command.
```
re [command]
```

# examples
## rerun command e.g. `go test` - watching file to re run your command again whenever file changed.
```
re go test -v .
```

[example use for go tdd](https://imgur.com/a/FQgvJny)

## rerun api service
- TODO

## Feature
* [x] rerun one time command e.g. `go test`
* [ ] interrupt and rerun
* [x] Watching nested files
* [x] Watching single files
* [ ] reload app when file change
* [ ] configurable option
* [ ] should clear screen before rerun command

## Contribute
- please send a PR.

## License
[MIT](https://github.com/labstack/echo/blob/master/LICENSE)
