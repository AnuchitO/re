BINARY  := re
CMD     ?= go test ./...
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test coverage vet lint lint-install run install clean help

## build: compile the binary to ./re
build:
	go build $(LDFLAGS) -o $(BINARY) .

## test: run all tests
test:
	go test -v ./...

## coverage: run tests and open an HTML coverage report
coverage:
	go test -covermode=count -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## vet: run go vet
vet:
	go vet ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## lint-install: install golangci-lint (requires Go)
lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

## run: watch files and rerun CMD (default: go test ./...)
##      usage: make run CMD="go test -v ."
run: build
	./$(BINARY) $(CMD)

## install: install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) .

## clean: remove build artifacts
clean:
	rm -f $(BINARY) coverage.out

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## /  /'
