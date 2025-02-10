# Contributing to sql2csv

First off, thank you for considering contributing to sql2csv! It's people like you that make sql2csv such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* Use a clear and descriptive title
* Describe the exact steps which reproduce the problem
* Provide specific examples to demonstrate the steps
* Describe the behavior you observed after following the steps
* Explain which behavior you expected to see instead and why
* Include details about your configuration and environment

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* Use a clear and descriptive title
* Provide a step-by-step description of the suggested enhancement
* Provide specific examples to demonstrate the steps
* Describe the current behavior and explain which behavior you expected to see instead
* Explain why this enhancement would be useful

### Pull Requests

* Fork the repo and create your branch from `main`
* If you've added code that should be tested, add tests
* If you've changed APIs, update the documentation
* Ensure the test suite passes
* Make sure your code lints
* Issue that pull request!

## Development Setup

1. Fork and clone the repo
2. Run `go mod download` to install dependencies
3. Create a branch for your changes

### Prerequisites

* Go 1.16 or later
* CGO enabled (required for SQLite support)
* Access to test databases:
  * PostgreSQL
  * MySQL
  * SQLite

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests for a specific package
go test ./pkg/database
```

### Test Database Setup

For running integration tests, you'll need:

1. PostgreSQL database:
   ```bash
   export POSTGRES_HOST=localhost
   export POSTGRES_PORT=5432
   export POSTGRES_USER=postgres
   export POSTGRES_PASSWORD=postgres
   export POSTGRES_DB=test_db
   ```

2. MySQL database:
   ```bash
   export MYSQL_HOST=localhost
   export MYSQL_PORT=3306
   export MYSQL_USER=root
   export MYSQL_PASSWORD=root
   export MYSQL_DB=test_db
   ```

## Style Guide

### Go Code Style

* Follow the standard Go formatting rules (use `gofmt`)
* Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
* Document all exported functions, types, and constants
* Keep functions focused and small
* Use meaningful variable names

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line

### Documentation Style

* Use clear and consistent terminology
* Provide examples where appropriate
* Keep explanations concise but complete
* Update README.md if adding new features

## Release Process

1. Update version number in relevant files
2. Create a new tag following semantic versioning
3. Push the tag to trigger the release workflow
4. The GitHub Action will automatically:
   * Run tests
   * Build binaries for all platforms
   * Create a GitHub release
   * Upload artifacts
   * Update Homebrew tap

## Questions?

Feel free to open an issue with your question or contact the maintainers directly. 