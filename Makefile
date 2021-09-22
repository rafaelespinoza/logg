GO ?= go

.PHONY: deps test

all: deps build test

deps:
	$(GO) mod tidy && $(GO) mod download && $(GO) mod verify

build:
	$(GO) build .

# Specify test flags with T variable. Example:
# make test T='-v -count=1 -failfast'
test:
	$(GO) test . $(T)
