.PHONY: test test-race test-bench fmt vet lint install-tools fieldalignment

test:
	go test -v ./...

# test-race:
# 	go test -race -v ./...

# test-bench:
# 	go test -bench=. -benchmem ./...

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

lint:
	golangci-lint run

fieldalignment:
	fieldalignment ./...

build:
	go build ./...

clean:
	go clean ./...
	rm -f coverage.out coverage.html

install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
