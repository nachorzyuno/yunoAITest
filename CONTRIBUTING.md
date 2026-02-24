# Contributing to FX-Aware Settlement Engine

Thank you for your interest in contributing to the FX-Aware Settlement Engine! This document provides guidelines and best practices for contributing to this project.

## Code of Conduct

- Be respectful and professional in all interactions
- Focus on constructive feedback
- Welcome newcomers and help them get started
- Keep discussions focused on technical merits

## Getting Started

### Prerequisites

- Go 1.24.5 or higher
- Git
- Basic understanding of Go and financial systems

### Setup Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/solara-settlement.git
   cd yunoAITest
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run tests to ensure everything works:
   ```bash
   go test ./...
   ```

## Development Workflow

### 1. Create a Branch

Create a feature branch from `main`:

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-new-currency` for new features
- `fix/decimal-rounding-bug` for bug fixes
- `docs/update-readme` for documentation
- `refactor/simplify-processor` for refactoring

### 2. Make Changes

Follow these coding standards:

#### Code Style

- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch common issues
- Follow standard Go naming conventions
- Keep functions small and focused (single responsibility)
- Use meaningful variable and function names

#### Financial Precision

- **ALWAYS** use `decimal.Decimal` for monetary amounts
- **NEVER** use `float32` or `float64` for money
- Use `decimal.Zero` instead of `0` for decimal comparisons
- Use `decimal.NewFromFloat()` only for constants, never for calculations

Example:
```go
// Good
amount := decimal.NewFromFloat(100.50)
total := amount.Add(fees)

// Bad
var amount float64 = 100.50
total := amount + fees  // Floating-point errors!
```

#### Error Handling

- Return meaningful error messages
- Use `fmt.Errorf()` with context
- Don't panic unless it's truly unrecoverable
- Wrap errors with additional context: `fmt.Errorf("failed to process transaction %s: %w", id, err)`

#### Testing

- Write unit tests for all business logic
- Aim for >80% code coverage
- Test edge cases:
  - Zero amounts
  - Negative amounts (refunds)
  - Missing/invalid data
  - Currency conversion edge cases
  - Date/time boundary conditions
- Use table-driven tests for multiple scenarios

Example:
```go
func TestTransaction_Validate(t *testing.T) {
    tests := []struct {
        name    string
        tx      *Transaction
        wantErr bool
    }{
        {
            name: "valid transaction",
            tx:   validTransaction(),
            wantErr: false,
        },
        {
            name: "missing transaction ID",
            tx:   transactionWithoutID(),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.tx.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### Documentation

- Add godoc comments to all exported types and functions
- Include usage examples in package documentation
- Update README.md for user-facing changes
- Document design decisions in code comments

### 3. Run Quality Checks

Before committing, ensure all checks pass:

```bash
# Format code
make fmt

# Run linter
make vet

# Run tests
make test

# Check coverage
make coverage
```

### 4. Commit Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add support for EUR currency conversion

- Add EUR to supported currencies enum
- Update mock provider with EUR base rate
- Add validation tests for EUR transactions
- Update documentation with EUR examples"
```

Commit message format:
- First line: Brief summary (50 chars or less)
- Blank line
- Detailed description with bullet points
- Reference issue numbers if applicable

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub with:
- Clear title describing the change
- Description of what changed and why
- Screenshots/examples if UI or output format changed
- Link to related issues

## Pull Request Guidelines

### PR Checklist

Before submitting, ensure:

- [ ] Code is formatted (`make fmt`)
- [ ] Linter passes (`make vet`)
- [ ] All tests pass (`make test`)
- [ ] New code has unit tests
- [ ] Coverage hasn't decreased
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] No merge conflicts with main

### PR Review Process

1. Automated checks run (tests, linter)
2. Maintainers review code
3. Address feedback and push updates
4. Once approved, maintainers will merge

## Adding New Features

### Adding a New Currency

1. Add currency to `internal/domain/currency.go`:
   ```go
   EUR Currency = "EUR" // Euro
   ```

2. Update validation in `Currency.Validate()`

3. Add base rate to `internal/fxrate/mock_provider.go`

4. Add tests for the new currency

5. Update README with supported currencies

### Adding a New FX Rate Provider

1. Implement the `Provider` interface in `internal/fxrate/`
2. Add configuration for API keys/endpoints
3. Add retry logic and error handling
4. Add integration tests (can be skipped in CI)
5. Document setup and usage

### Adding a New Report Format

1. Create new reporter in `internal/reporter/`
2. Implement formatting logic
3. Add CLI flag to select format
4. Add tests for output format
5. Update README with examples

## Common Issues

### Tests Failing Locally

```bash
# Clean and rebuild
make clean
go mod tidy
go test ./... -v
```

### Import Path Issues

Ensure you're using the correct module path:
```go
import "github.com/ignacio/solara-settlement/internal/domain"
```

### Decimal Precision Issues

Always use `decimal.Decimal` for money:
```go
amount := decimal.NewFromFloat(100.50)  // OK for constants
calculated := price.Mul(quantity)       // Preserves precision
```

## Getting Help

- Open an issue for bugs or feature requests
- Tag maintainers for urgent issues
- Check existing issues before creating new ones
- Provide minimal reproducible examples

## License

By contributing, you agree that your contributions will be part of the Solara Travel technical challenge project.

---

Thank you for contributing to the FX-Aware Settlement Engine!
