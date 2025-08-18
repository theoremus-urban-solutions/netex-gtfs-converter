# Contributing to NETEX GTFS Conveter Go

Thank you for your interest in contributing to NETEX GTFS Conveter Go! This document provides guidelines and information for contributors.

## Quick Start

1. **Fork** the repository
2. **Clone** your fork: `git clone https://github.com/your-username/netex-gtfs-converter-go.git`
3. **Install** Go 1.21+ and dependencies: `go mod download`
4. **Run tests** to ensure everything works: `go test ./...`
5. **Create a branch** for your changes: `git checkout -b feature/your-feature`

## Development Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** (optional) - For using Makefile commands

### Environment Setup

```bash
# Clone the repository
git clone https://github.com/theoremus-urban-solutions/netex-gtfs-converter.git
cd netex-gtfs-converter-go

# Download dependencies
go mod download

# Verify setup works
go test ./...
go build ./cmd/netex-gtfs-converter
```

## Development Workflow

### 1. Create an Issue

Before starting work, create or find an existing issue that describes:
- **Problem**: What issue are you solving?
- **Solution**: How do you plan to solve it?
- **Scope**: What will be changed?

### 2. Development Process

```bash
# Create feature branch
git checkout -b feature/issue-123-new-validator

# Make your changes
# Write tests for new functionality
# Update documentation if needed

# Run tests and linting
make test
make lint

# Commit your changes
git add .
git commit -m "feat: add new validator for X

- Implements validation for Y
- Fixes issue #123
- Includes comprehensive tests"
```

### 3. Testing Requirements

**All contributions must include tests:**

- **Unit tests** for new functions/methods
- **Integration tests** for new validators
- **Benchmark tests** for performance-critical code
- **CLI tests** for command-line changes

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...

# Test specific package
go test ./validator/core/
```

### 4. Code Quality Standards

#### Code Style

- Follow **standard Go conventions**
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write **clear, self-documenting code**
- Add **comments for exported functions**

#### Naming Conventions

```go
// Good
func ValidateRouteType(routeType int) error
type AgencyValidator struct{}

// Bad  
func validate_route_type(rt int) error
type agencyvalidator struct{}
```

#### Error Handling

```go
// Good - wrap errors with context
if err != nil {
    return fmt.Errorf("failed to parse route_type in %s:%d: %w", filename, rowNum, err)
}

// Bad - lose context
if err != nil {
    return err
}
```

## Project Structure

```
.
├── validator.go           # Public API - main entry points
├── implementation.go      # Internal implementation logic
├── doc.go                # Package documentation
├── cmd/                  # CLI applications
│   └── netex-gtfs-converter/   # Main CLI tool
├── examples/             # Usage examples
├── notice/               # Notice/error reporting system
├── parser/               # GTFS file parsing logic
├── report/               # Report generation
├── validator/            # Individual validator implementations
│   ├── core/            # Core validation (required fields, formats)
│   ├── entity/          # Entity validation (routes, stops, etc.)
│   ├── relationship/    # Cross-entity validation
│   ├── business/        # Business logic validation
│   └── accessibility/   # Accessibility validation
├── schema/               # GTFS data structure definitions
└── types/                # Custom types (Date, Time, Color, etc.)
```

## Adding New Validators

### 1. Create Validator File

```go
// validator/core/new_validator.go
package core

import (
    "github.com/theoremus-urban-solutions/netex-gtfs-converter/notice"
    "github.com/theoremus-urban-solutions/netex-gtfs-converter/schema"
)

type NewValidator struct {
    // configuration fields
}

func (v *NewValidator) Validate(feed *schema.Feed, noticeContainer *notice.NoticeContainer) {
    // Implementation
    for _, agency := range feed.Agencies {
        if someCondition {
            noticeContainer.AddNotice(notice.NewSomeErrorNotice(
                agency.AgencyID, 
                agency.RowNumber,
            ))
        }
    }
}

func (v *NewValidator) Name() string {
    return "NewValidator"
}
```

### 2. Add Notice Types

```go
// notice/validation_notices.go

type SomeErrorNotice struct {
    *BaseNotice
    AgencyID  string `json:"agencyId"`
    RowNumber int    `json:"rowNumber"`
}

func NewSomeErrorNotice(agencyID string, rowNumber int) *SomeErrorNotice {
    return &SomeErrorNotice{
        BaseNotice: NewBaseNotice(
            "some_error",
            ERROR,
            map[string]interface{}{
                "agencyId":  agencyID,
                "rowNumber": rowNumber,
            },
        ),
        AgencyID:  agencyID,
        RowNumber: rowNumber,
    }
}
```

### 3. Write Tests

```go
// validator/core/new_validator_test.go
package core_test

import (
    "testing"
    "github.com/theoremus-urban-solutions/netex-gtfs-converter/validator/core"
    "github.com/theoremus-urban-solutions/netex-gtfs-converter/notice"
    "github.com/theoremus-urban-solutions/netex-gtfs-converter/schema"
)

func TestNewValidator(t *testing.T) {
    tests := []struct {
        name           string
        feed           *schema.Feed
        expectedNotices []string
    }{
        {
            name: "valid case",
            feed: &schema.Feed{
                Agencies: []schema.Agency{
                    {AgencyID: "agency1", AgencyName: "Valid Agency"},
                },
            },
            expectedNotices: nil,
        },
        {
            name: "invalid case",
            feed: &schema.Feed{
                Agencies: []schema.Agency{
                    {AgencyID: "", AgencyName: "Invalid Agency"},
                },
            },
            expectedNotices: []string{"some_error"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := &core.NewValidator{}
            container := notice.NewNoticeContainer()
            
            validator.Validate(tt.feed, container)
            
            notices := container.GetNotices()
            if len(tt.expectedNotices) != len(notices) {
                t.Errorf("Expected %d notices, got %d", len(tt.expectedNotices), len(notices))
            }
            
            for i, expectedCode := range tt.expectedNotices {
                if notices[i].Code() != expectedCode {
                    t.Errorf("Expected notice code %s, got %s", expectedCode, notices[i].Code())
                }
            }
        })
    }
}
```

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks

### Examples

```bash
feat(validator): add route color contrast validator

Implements validation for route_color and route_text_color combinations
to ensure sufficient contrast for accessibility.

Fixes #123

fix(parser): handle UTF-8 BOM in CSV files

CSV files with UTF-8 BOM were not being parsed correctly.
Now strips BOM from the first header field.

docs: update API examples in README

- Add context cancellation example  
- Fix typo in quick start section
- Update CLI usage examples
```

## Pull Request Process

### 1. Before Submitting

- [ ] Tests pass: `go test ./...`
- [ ] Code is formatted: `gofmt -s -w .`
- [ ] Linting passes: `golangci-lint run`
- [ ] Documentation updated if needed
- [ ] Commit messages follow convention

### 2. PR Description Template

```markdown
## Summary
Brief description of changes

## Changes
- Added/fixed/updated X
- Removed Y
- Refactored Z

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing performed

## Documentation
- [ ] README updated
- [ ] API docs updated
- [ ] Examples updated

Fixes #issue-number
```

### 3. Review Process

1. **Automated checks** must pass (CI, tests, linting)
2. **Code review** by maintainers
3. **Discussion** and iteration if needed
4. **Merge** after approval

## Performance Guidelines

- **Benchmark** performance-critical code
- **Profile** memory usage for large datasets
- **Avoid** unnecessary allocations in hot paths
- **Use** appropriate data structures
- **Consider** parallel processing for CPU-intensive work

## Compatibility

- **Go version**: Support Go 1.21+
- **Breaking changes**: Avoid in minor versions
- **Dependencies**: Minimize external dependencies
- **APIs**: Keep public APIs stable

## Getting Help

- **Issues**: Create GitHub issues for bugs/features
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check README and examples first

## Recognition

Contributors will be recognized in:
- Release notes
- CONTRIBUTORS file
- GitHub contributor stats

Thank you for contributing to NETEX GTFS Conveter Go! 🚀