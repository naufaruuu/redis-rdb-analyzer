# Contributing to Redis RDB Analyzer

Thank you for your interest in contributing to Redis RDB Analyzer! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.18+** installed
- **kubectl** configured with access to a Kubernetes cluster (for testing K8s features)
- **SQLite3** (usually comes with the system)
- **Git** for version control
- **Make** (optional but recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/redis-rdb-analyzer.git
   cd redis-rdb-analyzer
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/naufaruuu/redis-rdb-analyzer.git
   ```

## Development Setup

### Install Dependencies

```bash
make deps
# or manually:
go mod download
```

### Build the Project

```bash
make build
# This creates ./redis-rdb-analyzer binary
```

### Run the Application

```bash
make run
# or manually:
./redis-rdb-analyzer -p 8080
```

Access the web UI at http://localhost:8080

### Project Structure

```
redis-rdb-analyzer/
├── main.go              # Entry point
├── decoder/             # RDB parsing logic
│   ├── decoder.go       # Core data structures
│   ├── hdt_adapter.go   # HDT parser adapter
│   ├── hdt_decode.go    # HDT decoding implementation
│   └── memprofiler.go   # Memory profiling
├── server/                # Web server & business logic
│   ├── show.go          # HTTP server & routes
│   ├── job.go           # Async job processing
│   ├── counter.go       # Statistics aggregation
│   ├── db.go            # SQLite persistence
│   ├── k8s_discovery.go # Kubernetes integration
│   └── utils.go         # Utility functions
├── views/               # HTML templates
└── Makefile             # Build automation
```

## Making Changes

### Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring

### Write Clear Commit Messages

Follow these guidelines:
- Use present tense ("Add feature" not "Added feature")
- First line: short summary (50 chars or less)
- Blank line, then detailed explanation if needed

Example:
```
Add Redis Cluster slot distribution chart

- Implement slot distribution calculation
- Add new chart component to dashboard
- Update counter.go to track slot metrics
```

## Code Style

### Go Formatting

Always format your code before committing:

```bash
make fmt
# or:
gofmt -w -s .
```

### Go Best Practices

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions small and focused
- Handle errors explicitly, don't ignore them

### Code Organization

- Place new HTTP handlers in `server/show.go`
- Add business logic to appropriate files in `server/`
- RDB parsing changes go in `decoder/`
- Keep decoder logic separate from web/API logic

## Testing

### Running Tests

Currently, this project has minimal test coverage. When adding tests:

```bash
make test
```

### Writing Tests

- Place test files next to the code they test: `filename_test.go`
- Use table-driven tests for multiple test cases
- Mock external dependencies (kubectl, file I/O)
- Aim for high coverage on critical paths (RDB parsing, job processing)

Example test structure:
```go
func TestSlotCalculation(t *testing.T) {
    tests := []struct {
        name     string
        key      string
        expected int
    }{
        {"simple key", "user:123", 5259},
        {"hash tag", "user:{123}:profile", 5259},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Slot(tt.key)
            if result != tt.expected {
                t.Errorf("got %d, want %d", result, tt.expected)
            }
        })
    }
}
```

## Submitting Changes

### Before Submitting

1. **Update your branch** with upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/master
   ```

2. **Run tests and checks**:
   ```bash
   make fmt
   make test
   make build
   ```

3. **Update documentation** if needed:
   - Update README.md for user-facing changes
   - Update GEMINI.MD for architectural changes
   - Add/update code comments

### Create a Pull Request

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a Pull Request on GitHub with:
   - Clear title describing the change
   - Detailed description of what changed and why
   - Screenshots for UI changes
   - Reference related issues (e.g., "Fixes #123")

### Pull Request Review Process

- Maintainers will review your PR
- Address feedback and update your PR
- Once approved, a maintainer will merge your changes

## Reporting Issues

### Bug Reports

When reporting bugs, include:
- **Description**: Clear description of the bug
- **Steps to Reproduce**: Numbered steps to reproduce
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**:
  - Go version (`go version`)
  - OS and version
  - kubectl version (for K8s issues)
  - RDB file size (if relevant)
- **Logs**: Relevant error messages or logs

### Feature Requests

For feature requests, describe:
- **Problem**: What problem does this solve?
- **Proposed Solution**: How should it work?
- **Alternatives**: Other solutions you considered
- **Use Case**: Real-world scenario where this helps

## Development Guidelines

### Kubernetes Integration

When working on K8s features:
- Test against a real cluster when possible
- Handle kubectl errors gracefully
- Provide clear error messages for missing permissions
- Don't assume kubectl is in PATH or configured

### RDB Parsing

When modifying decoder logic:
- Test with various RDB versions (v7, v9, v10+)
- Handle large files (>1GB) efficiently
- Maintain memory profiling accuracy
- Document encoding-specific quirks

### Web UI

When updating the UI:
- Maintain Tailwind CSS conventions
- Support both light and dark modes
- Ensure responsive design works on mobile
- Test with large datasets (100k+ keys)

### Security Considerations

- Never execute arbitrary shell commands from user input
- Validate file paths before reading
- Sanitize HTML output to prevent XSS
- Don't log sensitive data (credentials, tokens)

## Questions?

If you have questions about contributing:
- Open an issue with the `question` label
- Check existing issues and pull requests
- Review the documentation in README.md and GEMINI.MD

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to Redis RDB Analyzer! 🚀
