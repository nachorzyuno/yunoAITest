# Documentation Index

Welcome to the FX-Aware Settlement Engine documentation! This directory contains detailed technical documentation to help you understand, use, and contribute to the project.

## Quick Links

### Getting Started
- [Quick Start Guide](../QUICKSTART.md) - Get running in 5 minutes
- [Main README](../README.md) - Comprehensive project overview
- [Test Data Guide](../testdata/README.md) - Understanding test data and output

### Development
- [Architecture Documentation](ARCHITECTURE.md) - System design and technical decisions
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute code
- [API Documentation](#api-documentation) - Godoc for all packages

## Documentation Overview

### 1. User Documentation

#### [README.md](../README.md)
The main project documentation covering:
- Problem statement and solution overview
- Installation and setup instructions
- Usage examples and command reference
- Input/output CSV format specifications
- Design decisions and business rules
- Testing guide

**Audience**: All users, from first-time users to contributors

#### [QUICKSTART.md](../QUICKSTART.md)
A condensed guide to get started quickly:
- 5-minute installation
- First settlement run
- Common commands
- Basic troubleshooting

**Audience**: New users who want to try the system immediately

#### [testdata/README.md](../testdata/README.md)
Detailed guide to test data:
- Test file descriptions
- How to generate test data
- Expected output format
- Edge cases and validation
- Manual testing scenarios

**Audience**: Users testing the system, QA engineers

### 2. Developer Documentation

#### [ARCHITECTURE.md](ARCHITECTURE.md)
Comprehensive architecture guide covering:
- High-level system architecture
- Layer responsibilities and interactions
- Data flow and processing pipeline
- Design decisions and rationale
- Error handling and testing strategies
- Performance considerations
- Extensibility points

**Audience**: Developers, architects, technical reviewers

#### [CONTRIBUTING.md](../CONTRIBUTING.md)
Guidelines for contributing:
- Development workflow
- Code style and standards
- Financial precision requirements
- Testing requirements
- Pull request process
- Common issues and solutions

**Audience**: Contributors, maintainers

### 3. API Documentation

#### Godoc Comments
Every package includes comprehensive godoc comments:

- **Package-level documentation**: Purpose and usage examples
- **Type documentation**: Explanation of entities and their fields
- **Function documentation**: Parameters, return values, behavior
- **Examples**: Code snippets showing typical usage

View the documentation:
```bash
# Generate and view locally
go doc -all ./internal/domain
go doc -all ./internal/fxrate
go doc -all ./internal/processor
go doc -all ./internal/settlement
go doc -all ./internal/reporter

# Or start a local doc server
godoc -http=:6060
# Then visit: http://localhost:6060/pkg/github.com/ignacio/solara-settlement/
```

**Audience**: Developers using or extending the codebase

## Documentation by Use Case

### "I want to use the settlement engine"
1. Start with [QUICKSTART.md](../QUICKSTART.md)
2. Read [Input/Output format](../README.md#input-csv-format) in README
3. Check [testdata/README.md](../testdata/README.md) for examples

### "I want to understand how it works"
1. Read [README Overview](../README.md#overview)
2. Study [ARCHITECTURE.md](ARCHITECTURE.md)
3. Review domain models: `go doc ./internal/domain`

### "I want to contribute code"
1. Read [CONTRIBUTING.md](../CONTRIBUTING.md)
2. Review [ARCHITECTURE.md](ARCHITECTURE.md)
3. Check existing code with godoc
4. Run tests: `make test`

### "I need to add a feature"
1. Understand [Architecture](ARCHITECTURE.md)
2. Review [Extensibility Points](ARCHITECTURE.md#extensibility-points)
3. Follow [Contributing Guidelines](../CONTRIBUTING.md)
4. Write tests first (TDD)

### "I found a bug"
1. Check [Troubleshooting](../testdata/README.md#troubleshooting)
2. Review [Common Issues](../CONTRIBUTING.md#common-issues)
3. Open a GitHub issue with:
   - Steps to reproduce
   - Expected vs actual behavior
   - Sample data if applicable

## Key Concepts

### Multi-Currency Settlement
The system processes transactions in local currencies (ARS, BRL, COP, MXN) and converts them to USD using historical exchange rates specific to each transaction's date.

### Financial Precision
All monetary calculations use `decimal.Decimal` instead of floating-point numbers to avoid rounding errors and ensure accurate financial reporting.

### Transaction Types
- **Capture**: Money received (adds to settlement)
- **Refund**: Money returned (subtracts from settlement)
- **Authorization**: Intent only (tracked but not settled)

### Settlement Rules
Only completed captures and refunds are included in settlements. Pending and failed transactions are excluded.

## Project Structure

```
yunoAITest/
├── cmd/
│   └── settlement/          # CLI application
├── internal/
│   ├── domain/              # Business entities
│   ├── fxrate/              # FX rate service
│   ├── processor/           # CSV ingestion
│   ├── settlement/          # Settlement calculation
│   └── reporter/            # Report generation
├── scripts/
│   └── generate_testdata.go # Test data generator
├── testdata/                # Sample data files
├── docs/                    # Documentation (you are here)
│   ├── README.md            # This file
│   └── ARCHITECTURE.md      # Technical architecture
├── README.md                # Main documentation
├── QUICKSTART.md            # Quick start guide
├── CONTRIBUTING.md          # Contribution guidelines
├── Makefile                 # Build automation
└── go.mod                   # Go module definition
```

## Makefile Commands

Quick reference for common operations:

```bash
make help          # Show all available commands
make all           # Run tests and build
make test          # Run all tests
make build         # Build CLI binary
make run           # Generate test data and run settlement
make clean         # Remove generated files
make fmt           # Format Go code
make vet           # Run Go vet linter
make coverage      # Generate coverage report
```

## Additional Resources

### Go Documentation
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Financial Precision
- [shopspring/decimal library](https://github.com/shopspring/decimal)
- [Why not use float64 for money](https://husobee.github.io/money/float/2016/09/23/never-use-floats-for-currency.html)

### CSV Standards
- [RFC 4180 - CSV Format](https://tools.ietf.org/html/rfc4180)

### Testing
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

## Keeping Documentation Updated

When making changes to the code:
1. Update relevant documentation files
2. Update godoc comments for affected packages
3. Add examples for new features
4. Update README if user-facing changes
5. Update ARCHITECTURE.md if design changes

## Feedback

Documentation improvements are always welcome! If you find:
- Missing information
- Unclear explanations
- Broken links
- Outdated content

Please open a GitHub issue or submit a pull request.

---

**Last Updated**: 2026-02-24
**Version**: 1.0.0
**Maintainers**: Solara Travel Engineering Team
