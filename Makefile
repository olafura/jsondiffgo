SHELL := /bin/bash

.PHONY: all test bench fuzz lint js-compare profile-big profile-medium build-profile

GOCACHE ?= $(PWD)/.gocache

all: test

test:
	GOCACHE=$(GOCACHE) go test ./...

bench:
	GOCACHE=$(GOCACHE) go test -run '^$$' -bench . -benchmem

fuzz:
	GOCACHE=$(GOCACHE) go test -run '^$$' -fuzz=Fuzz -fuzztime=$${FUZZTIME:-10s}

lint:
	XDG_CACHE_HOME=$(GOCACHE) GOLANGCI_LINT_CACHE=$(GOCACHE)/golangci-lint golangci-lint run --timeout=5m

js-compare:
	JSONDIFFGO_COMPARE_JS=1 GOCACHE=$(GOCACHE) go test ./...
