VERSION := 0.0.1 # $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECT := $(shell basename "$(PWD)")

CGO_ENABLED ?= 0
GOOS ?= 'linux'
GOARCH ?= 'amd64'
GOBIN ?= /go/bin
GOFILES := $(wildcard *.go)

MAKEFILE := $(abspath $(lastword $(MAKEFILE_LIST)))
MAKEFLAGS += --silent
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
STDERR := /tmp/$(PROJECT).error

## build: Compiles the binary.
build:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -f $(MAKEFILE) -s go-build 2> $(STDERR)
	@cat $(STDERR) | sed 's/make\[.*/ /' | sed '/^/s/^/    /' 1>&2

## clean: Clean build files.
clean:
	@-rm $(GOBIN)/$(PROJECT) 2> /dev/null
	@-$(MAKE) -f $(MAKEFILE) go-clean

go-build:
	@GOBIN=$(GOBIN) CGO_ENABLED=$(CGO) GOOS=$(OS) GOARCH=$(ARCH) \
		go build $(LDFLAGS) -o $(GOBIN)/$(PROJECT) $(GOFILES)

go-clean:
	@GOBIN=$(GOBIN) go clean

.PHONY: help
all: help
help: Makefile
		@echo 'Choose a command to run in "$(PROJECT)":'
		@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/  /'
