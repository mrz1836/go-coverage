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

## Update version number in multiple locations
.PHONY: update-version
update-version: ## Update version number (usage: make update-version version=1.0.1)
	@if [ -z "$(version)" ]; then \
		echo "Error: version parameter is required. Usage: make update-version version=1.0.1"; \
		exit 1; \
	fi; \
	echo "Updating version to $(version)..."; \
	\
	printf "Updating Version in root.go: "; \
	if grep -E 'Version:.*"[0-9]+\.[0-9]+\.[0-9]+"' cmd/go-coverage/cmd/root.go >/dev/null 2>&1; then \
		sed -i '' -E 's/Version:.*"[0-9]+\.[0-9]+\.[0-9]+"/Version: "$(version)"/' cmd/go-coverage/cmd/root.go && \
		echo "✓ updated"; \
	else \
		echo "not found"; \
	fi; \
	\
	printf "Updating GeneratorVersion in complete.go: "; \
	if grep -E 'GeneratorVersion:.*"[0-9]+\.[0-9]+\.[0-9]+"' cmd/go-coverage/cmd/complete.go >/dev/null 2>&1; then \
		sed -i '' -E 's/GeneratorVersion:.*"[0-9]+\.[0-9]+\.[0-9]+"/GeneratorVersion: "$(version)"/' cmd/go-coverage/cmd/complete.go && \
		echo "✓ updated"; \
	else \
		echo "not found"; \
	fi; \
	\
	printf "Updating version in README.md: "; \
	if grep -E '# Go Coverage v[0-9]+\.[0-9]+\.[0-9]+' README.md >/dev/null 2>&1; then \
		sed -i '' -E 's/# Go Coverage v[0-9]+\.[0-9]+\.[0-9]+/# Go Coverage v$(version)/' README.md && \
		echo "✓ updated"; \
	else \
		echo "not found"; \
	fi; \
	\
	printf "Updating CITATION.cff: "; \
	$(MAKE) citation version=$(version) > /dev/null 2>&1 && echo "✓ updated" || echo "⚠️ failed"; \
	\
	printf "Updating GO_COVERAGE_VERSION in .env.shared: "; \
	if grep -E 'GO_COVERAGE_VERSION=v[0-9]+\.[0-9]+\.[0-9]+' .github/.env.shared >/dev/null 2>&1; then \
		sed -i '' -E 's/GO_COVERAGE_VERSION=v[0-9]+\.[0-9]+\.[0-9]+/GO_COVERAGE_VERSION=v$(version)/' .github/.env.shared && \
		echo "✓ updated"; \
	else \
		echo "not found"; \
	fi; \
	\
	echo "Version update complete!"
