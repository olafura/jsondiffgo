SHELL := /bin/bash

.PHONY: all test bench fuzz lint js-compare bdd

GOCACHE ?= $(PWD)/.gocache
FUZZTIME ?= 10s

all: test

test:
	GOCACHE=$(GOCACHE) go test ./...

bench:
	GOCACHE=$(GOCACHE) go test -run '^$$' -bench . -benchmem

# Generic pattern for individual fuzz targets
fuzz-%:
	$(eval FUZZ_FUNC := $(shell echo "Fuzz$*" | sed 's/-\([a-z]\)/\u\1/g'))
	GOCACHE=$(GOCACHE) go test -run '^$$' -fuzz=$(FUZZ_FUNC) -fuzztime=${FUZZTIME}

# Run all fuzz tests
fuzz:
	@echo "Running all fuzz tests..."
	@for func in $$(grep -h '^func Fuzz' *_test.go 2>/dev/null | cut -d'(' -f1 | cut -d' ' -f2 | sort -u); do \
		echo "Running fuzz test: $$func"; \
		GOCACHE=$(GOCACHE) go test -run '^$$' -fuzz=$$func -fuzztime=${FUZZTIME} || exit 1; \
	done

lint:
	XDG_CACHE_HOME=$(GOCACHE) GOLANGCI_LINT_CACHE=$(GOCACHE)/golangci-lint golangci-lint run --timeout=5m

js-compare:
	JSONDIFFGO_COMPARE_JS=1 GOCACHE=$(GOCACHE) go test ./...

bdd:
	GOCACHE=$(GOCACHE) go test ./bdd -run TestBDDFeatures -v
