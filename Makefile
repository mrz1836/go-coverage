# Common makefile commands & variables between projects
include .make/common.mk

# Common Golang makefile commands & variables between projects
include .make/go.mk

## Set default repository details if not provided
REPO_NAME  ?= go-coverage
REPO_OWNER ?= mrz1836

# Variables
BINARY_NAME := go-coverage
BINARY_PATH := ./bin/$(BINARY_NAME)
GO := go
MODULE := github.com/mrz1836/go-coverage

# Build flags with version injection
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC' 2>/dev/null || echo "unknown")

LDFLAGS := -ldflags="-s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(BUILD_DATE)"
BUILD_FLAGS := -trimpath

## Override the default build-go to build from the cmd directory
.PHONY: build-go
build-go: ## Build the Go application (locally)
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p bin
	@$(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_PATH) $(TAGS) $(GOFLAGS) ./cmd/$(BINARY_NAME)
	@echo "Binary built: $(BINARY_PATH) (v$(VERSION))"

## Override the default install to install from the cmd directory
.PHONY: install
install: ## Install the application binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) v$(VERSION)..."
	@$(GO) install $(BUILD_FLAGS) $(LDFLAGS) $(TAGS) $(GOFLAGS) ./cmd/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed (v$(VERSION))"

## build: Build the pre-commit binary (alias for build-go)
.PHONY: build
build: build-go
