SHELL=bash
MAIN=dp-import-api

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	dis-vulncheck

.PHONY: build
build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) cmd/$(MAIN)/main.go

.PHONY: debug
debug:
	HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: acceptance
acceptance:
	MONGODB_DATABASE=test HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: test build debug
