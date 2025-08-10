# Common makefile commands & variables between projects
include .make/common.mk

# Common Golang makefile commands & variables between projects
include .make/go.mk

## Set default repository details if not provided
REPO_NAME  ?= go-coverage
REPO_OWNER ?= mrz1836

## Override the default build-go to build from the cmd directory
.PHONY: build-go
build-go: ## Build the Go application (locally)
	@echo "Building Go app..."
	@mkdir -p bin
	@go build -o bin/go-coverage $(TAGS) $(GOFLAGS) ./cmd/go-coverage/

## Override the default install to install from the cmd directory
.PHONY: install
install: ## Install the application binary to GOPATH/bin
	@echo "Installing go-coverage binary..."
	@go install $(TAGS) $(GOFLAGS) ./cmd/go-coverage/
