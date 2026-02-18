# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`lmk` is a CLI application for managing long-form prose projects using organized Markdown files. Written in Go.

## Development Process

**Strict TDD is mandatory.** Follow Red -> Green -> Refactor for all production code:
1. Write failing test first
2. Write minimal code to pass
3. Refactor while keeping tests green

Use table-driven tests where appropriate.

**Coverage Policy**: 100% coverage of testable code is required. The Impl pattern exempts external operations:
- `*Impl` functions wrap OS/exec calls and are excluded from coverage calculation
- See `.claude/skills/go-tdd/references/coverage.md` for exemption guidelines

## Commands (via justfile)

```bash
just test              # Run all tests
just test-cover        # Run tests with coverage report
just test-cover-check  # Verify coverage meets threshold
just vet               # Run go vet
just lint              # Run staticcheck
just fmt               # Format code with gofmt
just check             # Run all quality gates (test, vet, lint, fmt check)
```

Run a single test:
```bash
go test -run TestName ./path/to/package
```

## Quality Gates

All must pass before commit (enforced via lefthook pre-commit hooks and GitHub Actions):
- `gofmt` formatting
- `go vet ./...` (zero warnings)
- `staticcheck ./...` (zero warnings)
- `go test ./...` with 100% coverage required for non-Impl functions

Coverage exemptions (per `.claude/skills/go-tdd/references/coverage.md`):
- `*Impl` functions that wrap external commands (exec.Command, os operations)
- These are tested via integration tests, not unit tests
- The coverage check filters out Impl functions from the calculation

After cloning, install git hooks:
```bash
lefthook install
```

## Code Standards

- All errors must be handled explicitly (no `_` for errors)
- No panics except at process boundaries (e.g., `main()`)
- Interfaces defined by consumers, not producers
- `context.Context` required for cancellation/deadlines
- GoDoc comments on all exported APIs

## Active Technologies
- Go 1.23 + Cobra (CLI framework)
