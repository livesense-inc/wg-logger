BIN = wg-logger
CURRENT_REVISION ?= $(shell git rev-parse --short HEAD)
LDFLAGS = -w -s -X 'main.version=Unknown' -X 'main.gitcommit=$(CURRENT_REVISION)'

all: test build

test:
	go test ./...

build:
	go build -ldflags="$(LDFLAGS)" -trimpath -o bin/$(BIN) ./cmd

clean:
	rm -rf bin

.PHONY: test build clean
