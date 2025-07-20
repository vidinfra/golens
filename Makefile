// File: Makefile
.PHONY: test test-race test-cover lint fmt clean build install-tools

# Test commands
test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-bench:
	go test -bench=. -benchmem ./...

# Code quality
lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

# Build
build:
	go build ./...

# Clean
clean:
	go clean ./...
	rm -f coverage.out coverage.html

# Install development tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run examples
run-basic:
	cd examples/basic && go run main.go

run-advanced:
	cd examples/advanced && go run main.go

run-integration:
	cd examples/gin-integration && go run main.go

# Development workflow
dev: fmt vet lint test

# Release preparation
release: clean fmt vet lint test-race test-cover