# Makefile for Go Cafe project

# Run linter using golangci-lint
.PHONY: lint
lint:
	golangci-lint run ./...

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Run tests
.PHONY: test
test:
	go test -v ./...

# Build application
.PHONY: build
build:
	go build -o bin/auth ./auth/cmd/auth/main.go

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/

# Delete database file
.PHONY: delete-db
delete-db:
	rm -rf storage/*