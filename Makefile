## simple makefile to log workflow
.PHONY: all test clean build install

GOFLAGS ?= $(GOFLAGS:)

all: dep install test

copyright:
	python .scripts/make_copyright_headers.py

build:
	@go build $(GOFLAGS) ./...

dep:
	@go get $(GOFLAGS) ./...

install:
	@go install $(GOFLAGS) ./...

test: install
	@go test -v -cover $(GOFLAGS) ./...

bench: install
	@go test -run=NONE -bench=. $(GOFLAGS) ./...

clean:
	@go clean $(GOFLAGS) -i ./...

## EOF
