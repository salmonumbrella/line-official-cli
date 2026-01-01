.PHONY: build install clean test lint fmt fmt-check

BINARY_NAME := line
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/salmonumbrella/line-official-cli/internal/cmd.version=$(VERSION) \
                     -X github.com/salmonumbrella/line-official-cli/internal/cmd.commit=$(COMMIT) \
                     -X github.com/salmonumbrella/line-official-cli/internal/cmd.date=$(DATE)"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/line

install:
	go install $(LDFLAGS) ./cmd/line

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

test:
	go test -v ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Run 'make fmt' to fix formatting"; gofmt -l .; exit 1)

# Development helpers
run:
	go run ./cmd/line $(ARGS)

deps:
	go mod tidy
	go mod download

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  install  - Install to GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linter"
	@echo "  fmt      - Format code"
	@echo "  deps     - Download dependencies"
	@echo "  run      - Run with ARGS='...'"
