# Makefile for LockAnalyzer

.PHONY: build clean help test cli

# Variables
EXAMPLE_BINARY_NAME=lockanalyzer-example
CLI_BINARY_NAME=lockanalyzer-cli
BUILD_DIR=build
TEST_DIR=testdata

# Cibles principales
.PHONY: all build clean test test-unit test-integration test-coverage run-example run-cli

all: build

# Compilation
build:
	@echo "🔨 Compiling example..."
	go build -o $(BUILD_DIR)/$(EXAMPLE_BINARY_NAME) cmd/example/main.go
	@echo "🔨 Compiling CLI tool..."
	go build -o $(BUILD_DIR)/$(CLI_BINARY_NAME) cmd/lockanalyzer/main.go
	@echo "✅ Compilation completed"

# Clean
clean:
	@echo "🧹 Cleaning build files..."
	rm -rf $(BUILD_DIR)
	rm -f lock_report_*
	@echo "✅ Cleanup completed"

# Unit tests
test-unit:
	@echo "🧪 Running unit tests..."
	go test -v ./lockanalyzer/... -run "Test.*" -timeout 30s

# Integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	go test -v ./lockanalyzer/... -run "TestConcurrentTransactions|TestDetectBlockedTransactionsReal|TestGenerateLocksReportWithRealData|TestLockDetectionWithTriggers|TestPerformanceWithLargeDataset" -timeout 60s

# Formatter tests
test-formatters:
	@echo "🧪 Running formatter tests..."
	go test -v ./formatters/... -timeout 30s

# All tests
test: test-unit test-formatters test-integration

# Tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./lockanalyzer/... ./formatters/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "🎯 Coverage report generated: coverage.html"

# Run example
run-example:
	@echo "🚀 Running example..."
	go run ./cmd/example/main.go

# Run CLI
run-cli:
	@echo "🔍 Running CLI..."
	@if [ ! -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then make build; fi
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Simulate locks
simulate-locks:
	@echo "🔄 Simulating locks..."
	@chmod +x scripts/simulate_locks.sh
	./scripts/simulate_locks.sh

# Generate test reports
test-reports:
	@echo "🎯 Generating test reports..."
	@if [ ! -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then make build; fi
	./$(BUILD_DIR)/$(BINARY_NAME) --dsn "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" --format markdown --output test_report.md
	./$(BUILD_DIR)/$(BINARY_NAME) --dsn "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" --format json --output test_report.json
	./$(BUILD_DIR)/$(BINARY_NAME) --dsn "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" --format text --output test_report.txt

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Code verification
lint:
	@echo "🔍 Verifying code..."
	gofmt -s -w .
	golint ./...
	go vet ./...

# Help
help:
	@echo "🔒 LockAnalyzer - Makefile"
	@echo ""
	@echo "Available commands:"
	@echo "  make build     - Build the application and CLI tool"
	@echo "  make clean     - Clean build files"
	@echo "  make test      - Run tests"
	@echo "  make cli       - Show CLI tool help"
	@echo "  make run       - Run the main application"
	@echo "  make test-unit      - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-formatters - Run formatter tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make run-example    - Run example"
	@echo "  make run-cli        - Run CLI"
	@echo "  make simulate-locks - Simulate locks"
	@echo "  make test-reports   - Generate test reports"
	@echo "  make deps           - Install dependencies"
	@echo "  make lint           - Verify code"
	@echo "  make help           - Show this help"
	@echo ""
	@echo "CLI usage examples:"
	@echo "  ./build/lockanalyzer-cli -help"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -format=markdown"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -format=json -output=report.json"
	@echo "  ./build/lockanalyzer-cli -dsn='postgres://user@localhost:5432/testdb' -interval=10s"

# Run main application
run: build
	@echo "🚀 Running main application..."
	@./build/$(BINARY_NAME)

# Installation (optional)
install: build
	@echo "📦 Installing CLI tool..."
	sudo cp build/$(CLI_BINARY_NAME) /usr/local/bin/
	@echo "✅ Installation completed. Use 'lockanalyzer-cli' from anywhere"

# Uninstall
uninstall:
	@echo "🗑️  Uninstalling CLI tool..."
	sudo rm -f /usr/local/bin/$(CLI_BINARY_NAME)
	@echo "✅ Uninstallation completed"

# Usage examples
example-markdown: build
	@echo "📝 Example: Markdown report to stdout"
	@./build/$(CLI_BINARY_NAME) -dsn="postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" -format=markdown

example-json: build
	@echo "📊 Example: JSON report to file"
	@./build/$(CLI_BINARY_NAME) -dsn="postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" -format=json -output=example_report.json

example-monitoring: build
	@echo "⏰ Example: Real-time monitoring (5 seconds)"
	@./build/$(CLI_BINARY_NAME) -dsn="postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable" -interval=5s 