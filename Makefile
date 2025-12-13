.PHONY: build clean install test lint help init-db

# Build the binary
build:
	go build -o ev-oracle .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o ev-oracle-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o ev-oracle-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o ev-oracle-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o ev-oracle-windows-amd64.exe .

# Clean build artifacts
clean:
	rm -f ev-oracle ev-oracle-*

# Install to GOPATH/bin
install:
	go install .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	go vet ./...
	go fmt ./...

# Initialize the database schema
init-db:
	./ev-oracle init

# Show help
help:
	@echo "EV Oracle Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the binary"
	@echo "  make build-all      Build for multiple platforms"
	@echo "  make clean          Remove build artifacts"
	@echo "  make install        Install to GOPATH/bin"
	@echo "  make test           Run tests"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make lint           Run linters and formatters"
	@echo "  make init-db        Initialize the database schema"
	@echo "  make help           Show this help message"
