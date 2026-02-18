# Justfile for lmk CLI project
# Run `just --list` to see available commands

# Default recipe: run all quality checks
default: check

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with coverage report (excludes main package)
test-cover:
    go test -coverprofile=coverage.out $(go list ./... | grep -v '^github.com/eykd/linemark-go$')
    go tool cover -func=coverage.out

# Run tests with HTML coverage report
test-cover-html: test-cover
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Check test coverage meets threshold (excludes main package)
# Note: *Impl functions are filtered out because they wrap external commands
# or OS-level operations that require fault injection to test.
# See .claude/skills/go-tdd/references/coverage.md for exemption guidelines
#
# Policy: 100% coverage required for all non-Impl, non-main functions
test-cover-check:
    #!/usr/bin/env bash
    set -uo pipefail
    PACKAGES=$(go list ./... | grep -v '^github.com/eykd/linemark-go$' | grep -v '/cmd/pipeline$')
    go test -coverprofile=coverage.out $PACKAGES || exit 1
    # Check that all non-Impl functions are at 100%
    # Filter out: Impl functions (exempt), main (exempt), total line, and 100% functions
    UNCOVERED=$(go tool cover -func=coverage.out | grep -v "Impl" | grep -v "^.*main.go:" | grep -v "100.0%" | grep -v "^total:" || true)
    if [ -n "$UNCOVERED" ]; then
        echo "The following non-Impl functions are not at 100% coverage:"
        echo "$UNCOVERED"
        exit 1
    fi
    TOTAL=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "Coverage: ${TOTAL} (100% of non-Impl functions covered)"

# Run go vet
vet:
    go vet ./...

# Run staticcheck linter
lint:
    staticcheck ./...

# Format code
fmt:
    gofmt -w .

# Check formatting (fails if not formatted)
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

# Run all quality gates
check: fmt-check vet lint test-cover-check

# Build the binary
build:
    go build -o bin/lmk .

# Clean build artifacts
clean:
    rm -rf bin/ coverage.out coverage.html generated-acceptance-tests/ acceptance-pipeline/ir/

# Install dependencies
deps:
    go mod download
    go mod tidy

# Run a single test by name (usage: just test-one TestName)
test-one NAME:
    go test -v -run {{NAME}} ./...

# Full acceptance pipeline: parse specs -> generate tests -> run
acceptance:
    ./run-acceptance-tests.sh

# Parse specs/*.txt to IR JSON
acceptance-parse:
    go run ./acceptance/cmd/pipeline -action=parse

# Generate Go tests from IR JSON
acceptance-generate:
    go run ./acceptance/cmd/pipeline -action=generate

# Run generated acceptance tests only
acceptance-run:
    go test -v ./generated-acceptance-tests/...

# Run both unit tests and acceptance tests
test-all: test acceptance
