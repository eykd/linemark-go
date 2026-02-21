# Quickstart: Linemark MVP Development

**Date**: 2026-02-19 | **Branch**: `001-linemark-mvp`

## Prerequisites

- Go 1.23+
- `just` (command runner)
- `staticcheck` (`go install honnef.co/go/tools/cmd/staticcheck@latest`)
- `lefthook` (`go install github.com/evilmartians/lefthook/cmd/lefthook@latest`)

## Setup

```bash
git clone <repo> && cd linemark-go
git checkout 001-linemark-mvp
lefthook install
go mod download
```

## Development Workflow

### ATDD Outer Loop

Every user story follows the acceptance test-driven workflow:

1. Write GWT spec in `specs/US<N>-<title>.txt`
2. Run `just acceptance` — acceptance test fails (Acceptance Red)
3. Inner TDD cycles until acceptance test passes (Acceptance Green)

### Inner TDD Cycle

```bash
# 1. Write failing test
just test              # RED

# 2. Write minimal code to pass
just test              # GREEN

# 3. Refactor, keeping tests green
just test              # Still GREEN

# 4. Check acceptance tests
just acceptance        # GREEN = done, RED = continue inner loop
```

### Quality Gates

```bash
just check             # Runs all gates: fmt, vet, lint, test-cover-check
just test-all          # Both unit tests and acceptance tests
```

## Package Layout

```
linemark-go/
├── main.go                     # Entry point (signal handling, cmd.ExecuteContext)
├── cmd/                        # Cobra command definitions
│   ├── root.go                 # Root command, global flags
│   ├── add.go                  # lmk add
│   ├── list.go                 # lmk list
│   ├── move.go                 # lmk move
│   ├── delete.go               # lmk delete
│   ├── rename.go               # lmk rename
│   ├── check.go                # lmk check
│   ├── doctor.go               # lmk doctor
│   ├── compact.go              # lmk compact
│   └── types.go                # lmk types (list/add/remove subcommands)
├── internal/
│   ├── domain/                 # Pure domain types and business rules
│   │   ├── node.go             # Node, MaterializedPath, Document types
│   │   ├── filename.go         # Filename parsing and generation
│   │   ├── selector.go         # Selector parsing (MP vs SID disambiguation)
│   │   ├── numbering.go        # Sibling numbering and gap-finding
│   │   └── outline.go          # Outline (tree construction from parsed files)
│   ├── sid/                    # SID generation
│   │   └── sid.go              # crypto/rand base62 generation
│   ├── slug/                   # Slug generation
│   │   └── slug.go             # Unicode-aware kebab-case slugging
│   ├── frontmatter/            # YAML frontmatter handling
│   │   └── frontmatter.go      # Parse, get/set title, serialize
│   ├── lock/                   # Advisory file locking
│   │   └── lock.go             # gofrs/flock wrapper
│   └── outline/                # Application service layer
│       └── service.go          # OutlineService (orchestrates operations)
├── acceptance/                 # Acceptance test pipeline
├── specs/                      # GWT specs and plan artifacts
├── docs/                       # Functional spec and glossary
├── justfile                    # Build commands
├── lefthook.yml                # Pre-commit hooks
└── ralph.sh                    # ATDD outer loop script
```

## Key Conventions

- **Interfaces defined by consumers**: Each command or service defines the interfaces it needs
- **Impl pattern**: OS/exec calls wrapped in `*Impl` functions, excluded from coverage
- **Errors to stderr**: Use `fmt.Fprintf(os.Stderr, ...)` for errors and debug output
- **JSON to stdout**: Use `json.NewEncoder(os.Stdout)` for structured output
- **context.Context**: First parameter on all functions that support cancellation
- **Table-driven tests**: For functions with multiple input/output combinations

## New Dependencies to Add

```bash
go get gopkg.in/yaml.v3
go get github.com/gofrs/flock
go get golang.org/x/text
```
