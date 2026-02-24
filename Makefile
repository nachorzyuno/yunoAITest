.PHONY: all test build clean generate-data run fmt vet coverage help

# Default target
all: test build

# Display help information
help:
	@echo "FX-Aware Settlement Engine - Available Make Targets"
	@echo ""
	@echo "  make all             - Run tests and build binary (default)"
	@echo "  make test            - Run all tests"
	@echo "  make build           - Build the CLI binary"
	@echo "  make clean           - Remove built binaries and generated files"
	@echo "  make generate-data   - Generate test transaction data"
	@echo "  make run             - Generate test data and run settlement engine"
	@echo "  make fmt             - Format all Go code"
	@echo "  make vet             - Run Go vet linter"
	@echo "  make coverage        - Generate HTML coverage report"
	@echo "  make help            - Show this help message"
	@echo ""

# Run all tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

# Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the CLI binary
build:
	@echo "Building settlement CLI..."
	go build -o settlement cmd/settlement/main.go
	@echo "Binary created: ./settlement"

# Clean built binaries and generated files
clean:
	@echo "Cleaning up..."
	rm -f settlement
	rm -f testdata/transactions.csv
	rm -f settlements.csv
	rm -f coverage.out
	rm -f coverage.html
	@echo "Clean complete"

# Generate test data
generate-data:
	@echo "Generating test data..."
	go run scripts/generate_testdata.go --output testdata/transactions.csv
	@echo "Test data created: testdata/transactions.csv"

# Run the settlement engine with test data
run: generate-data
	@echo "Running settlement engine..."
	go run cmd/settlement/main.go --input testdata/transactions.csv --output settlements.csv
	@echo "Settlement report created: settlements.csv"

# Run the settlement engine with the built binary
run-binary: build generate-data
	@echo "Running settlement engine (binary)..."
	./settlement --input testdata/transactions.csv --output settlements.csv
	@echo "Settlement report created: settlements.csv"

# Format all Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Format complete"

# Run Go vet linter
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete"

# Run all quality checks (format, vet, test)
check: fmt vet test
	@echo "All quality checks passed"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"
