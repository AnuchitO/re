# re

`re` is a program for rerun your command when files changed with a focus on simplicity.

re is just the program take your go command to rerun whenever file change.
re is easy to use just type `re` and follow with your command.

- `re` is easy to type.
- Watch files to run tests.
- Watch your .Go files and restart your processes without any configuration hassle

# Usage
```
re [any command]
```

# examples
## rerun command e.g. `go test` - watching file to re run your command again whenever file changed.
```
re go test -v .
```

![tdd](example_tdd.gif)

## run rest api
- TODO

## Feature
* [ ] run test wherever you want me to change
* [ ] Watching nested files and directories
* [ ] Watching single files or directories
* [ ] reload app when file change
* [ ] configurable optionx

## Contribute
- please send a PR.

## License
[MIT](https://github.com/labstack/echo/blob/master/LICENSE)
