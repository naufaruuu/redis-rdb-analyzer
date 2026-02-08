.PHONY: build test clean install lint help run docker-build

# Binary name
BINARY_NAME=redis-rdb-analyzer
BINARY_OUTPUT=redis-rdb-analyzer

# Version information
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

all: clean build

build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	CGO_ENABLED=1 $(GOBUILD) ${LDFLAGS} -o ${BINARY_OUTPUT} .
	@echo "Build complete: ${BINARY_OUTPUT}"

build-linux: ## Build for Linux (useful for Docker)
	@echo "Building ${BINARY_NAME} for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) ${LDFLAGS} -o ${BINARY_OUTPUT}-linux .
	@echo "Build complete: ${BINARY_OUTPUT}-linux"

run: build ## Build and run the application
	@echo "Starting ${BINARY_NAME}..."
	./${BINARY_OUTPUT}

run-dev: build ## Build and run with .env file (for local development)
	@echo "Starting ${BINARY_NAME} with .env configuration..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | xargs) && ./${BINARY_OUTPUT}; \
	else \
		echo "Warning: .env file not found. Copy .env.example to .env"; \
		echo "Running with defaults..."; \
		./${BINARY_OUTPUT}; \
	fi

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests complete"

coverage: test ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f ${BINARY_OUTPUT}
	rm -f ${BINARY_OUTPUT}-linux
	rm -f coverage.txt
	rm -f coverage.html
	rm -f *.db
	rm -rf tmp/
	@echo "Clean complete"

install: ## Install binary to GOPATH/bin
	@echo "Installing ${BINARY_NAME}..."
	$(GOCMD) install ${LDFLAGS} .
	@echo "Installed to GOPATH/bin"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -w -s .
	@echo "Format complete"

lint: ## Run linters (requires golangci-lint)
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run
	@echo "Lint complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t ${BINARY_NAME}:${VERSION} .
	docker tag ${BINARY_NAME}:${VERSION} ${BINARY_NAME}:latest
	@echo "Docker image built: ${BINARY_NAME}:${VERSION}"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help
