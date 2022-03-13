GO ?= go
GOSEC ?= gosec

SRC_PKG_PATH=github.com/rafaelespinoza/logg

.PHONY: all deps build vet test gosec

all: deps build vet test

deps:
	$(GO) mod download && $(GO) mod tidy

build:
	$(GO) build $(SRC_PKG_PATH)

vet:
	$(GO) vet $(ARGS) $(SRC_PKG_PATH)

# Specify test flags with ARGS variable. Example:
# make test ARGS='-v -count=1 -failfast'
test:
	$(GO) test $(ARGS) $(SRC_PKG_PATH)

# Run a security scanner over the source code. This Makefile won't install the
# scanner binary for you, so check out the gosec README for instructions:
# https://github.com/securego/gosec
#
# If necessary, specify the path to the built binary with the GOSEC env var.
#
# Also note, the package paths (last positional input to gosec command) should
# be a "relative" package path. That is, starting with a dot.
gosec:
	$(GOSEC) $(ARGS) ./...
