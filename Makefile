.PHONY: build run test clean docker-build docker-run help

# Variables
BINARY_NAME=llmgate
DOCKER_IMAGE=llmgate:latest
GO=go

# Default target
all: build

## build: Build the binary
build:
	$(GO) build -o $(BINARY_NAME) .

## run: Build and run the server
run: build
	./$(BINARY_NAME) -config config.yaml

## dev: Run in development mode
dev:
	$(GO) run . -config config.yaml

## test: Run all tests
test:
	$(GO) test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

## docker-build: Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

## docker-run: Run Docker container
docker-run:
	docker run -p 8080:8080 \
		-e OPENAI_API_KEY=$${OPENAI_API_KEY} \
		-e ANTHROPIC_API_KEY=$${ANTHROPIC_API_KEY} \
		$(DOCKER_IMAGE)

## fmt: Format Go code
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## lint: Run linter (requires golangci-lint)
lint:
	golangci-lint run

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //g'
