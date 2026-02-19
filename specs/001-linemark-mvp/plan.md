# Implementation Plan: Linemark MVP

**Branch**: `001-linemark-mvp` | **Date**: 2026-02-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/001-linemark-mvp/spec.md`

## Summary

Build `lmk`, a CLI tool for managing hierarchical outlines of Markdown documents using filenames. The tool encodes tree hierarchy in materialized paths within a flat directory, uses stable short IDs (SIDs) for identity, and provides commands for adding, moving, deleting, renaming, listing, validating, and repairing outlines. Implementation uses Go with Cobra, follows strict TDD with ATDD outer loop, and targets 100% coverage of testable code.

## Technical Context

**Language/Version**: Go 1.23 (module declares 1.25.6)
**Primary Dependencies**: Cobra (CLI), yaml.v3 (frontmatter), gofrs/flock (locking), golang.org/x/text (slug normalization)
**Storage**: Local filesystem — flat directory of Markdown files + `.linemark/` control directory
**Testing**: `go test` with 100% coverage (non-Impl), acceptance tests via GWT pipeline
**Target Platform**: Linux, macOS, Windows (cross-compilable, no CGO)
**Project Type**: Single Go CLI application
**Performance Goals**: All commands complete in reasonable time for up to 10,000 files; directory parsed on each invocation (O(n))
**Constraints**: No external database, no network, no caching in MVP, advisory locking via flock
**Scale/Scope**: Up to 10,000 files per outline, 999 siblings per parent, no hard depth limit

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

### Pre-Design Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. ATDD | PASS | GWT specs will drive all user stories; inner TDD for all production code |
| II. Static Analysis | PASS | go vet + staticcheck enforced via lefthook; all errors handled explicitly |
| III. Code Quality | PASS | gofmt mandatory; GoDoc on all exports; Go naming conventions followed |
| IV. Pre-commit Gates | PASS | lefthook already configured with fmt, vet, lint, test-cover-check |
| V. Warning Policy | PASS | All warnings addressed immediately, no deferrals |
| VI. Go CLI Target | PASS | Go 1.23, Cobra, no CGO, cross-compilable, path/filepath for portability |
| VII. Simplicity | PASS | YAGNI applied: no caching, no config files, no templates, no query system |

### Post-Design Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. ATDD | PASS | 13 user stories mapped to GWT specs; acceptance pipeline already scaffolded |
| II. Static Analysis | PASS | Consumer-defined interfaces in `cmd/` and `internal/outline/` |
| III. Code Quality | PASS | Package names: short, no stutter (`domain.Node` not `domain.DomainNode`) |
| IV. Pre-commit Gates | PASS | No changes to gate configuration needed |
| V. Warning Policy | PASS | New dependencies (yaml.v3, flock, x/text) are all actively maintained |
| VI. Go CLI Target | PASS | flock cross-platform via gofrs/flock; path/filepath throughout |
| VII. Simplicity | PASS | Domain types are value objects; no ORMs, no DI frameworks, no generics abuse |

No violations. Complexity Tracking section not required.

## Project Structure

### Documentation (this feature)

```text
specs/001-linemark-mvp/
├── plan.md              # This file
├── research.md          # Phase 0 output — technical research decisions
├── data-model.md        # Phase 1 output — entity definitions
├── quickstart.md        # Phase 1 output — development setup guide
├── contracts/
│   └── cli-commands.md  # Phase 1 output — CLI command contracts
└── (tasks created via beads, not markdown)
```

### Source Code (repository root)

```text
linemark-go/
├── main.go                         # Entry point (signal handling)
├── cmd/                            # Cobra command definitions
│   ├── root.go                     # Root command, --verbose, --json globals
│   ├── add.go                      # lmk add
│   ├── list.go                     # lmk list
│   ├── move.go                     # lmk move
│   ├── delete.go                   # lmk delete
│   ├── rename.go                   # lmk rename
│   ├── check.go                    # lmk check
│   ├── doctor.go                   # lmk doctor
│   ├── compact.go                  # lmk compact
│   └── types.go                    # lmk types (list/add/remove)
├── internal/
│   ├── domain/                     # Pure domain types and business rules
│   │   ├── node.go                 # Node, MaterializedPath, Document
│   │   ├── filename.go             # Filename parsing (regex) and generation
│   │   ├── selector.go             # Selector parsing and disambiguation
│   │   ├── numbering.go            # Sibling numbering, gap-finding, compact
│   │   └── outline.go              # Outline tree construction from parsed files
│   ├── sid/                        # SID generation
│   │   └── sid.go                  # crypto/rand base62 with rejection sampling
│   ├── slug/                       # Slug generation
│   │   └── slug.go                 # Unicode-aware kebab-case slugging
│   ├── frontmatter/                # YAML frontmatter handling
│   │   └── frontmatter.go          # Split, parse, get/set title, serialize
│   ├── lock/                       # Advisory locking
│   │   └── lock.go                 # gofrs/flock wrapper (Impl pattern)
│   └── outline/                    # Application service layer
│       └── service.go              # OutlineService orchestrating operations
├── acceptance/                     # Acceptance test pipeline (already exists)
├── specs/                          # GWT specs and plan artifacts
└── docs/                           # Functional spec and glossary
```

**Structure Decision**: Standard Go CLI layout with `internal/` for library packages and `cmd/` for Cobra commands. Domain types in `internal/domain/` are pure (no I/O); the service layer in `internal/outline/` orchestrates domain logic with filesystem operations. Cobra commands are thin wrappers that parse flags, call service methods, and format output.

## Architecture Decisions

### AD1: Domain Layer is Pure

All types in `internal/domain/` are pure functions and value objects with no I/O dependencies. This enables 100% unit test coverage without mocks. The domain layer:
- Parses filenames into structured types
- Constructs the outline tree from a list of parsed files
- Computes new materialized paths for add/move/compact
- Validates outline structure (finding detection)
- Generates filenames from domain types

### AD2: Service Layer Owns I/O

`internal/outline/service.go` depends on interfaces for filesystem operations:
- A `DirectoryReader` interface for listing and reading files
- A `FileWriter` interface for creating, renaming, and deleting files
- A `Locker` interface for advisory locking
- A `SIDReserver` interface for SID allocation

These interfaces are defined by the consumer (the service), not the producer. Production implementations use real filesystem calls (Impl pattern); tests use in-memory fakes.

### AD3: Cobra Commands Are Thin

Each command in `cmd/` does only:
1. Parse and validate flags/args
2. Call `OutlineService` method with domain types
3. Format result as JSON or human-readable text
4. Set exit code

No business logic in commands. Commands are tested via `cmd.Execute()` with argument injection.

### AD4: JSON-First Output Model

All commands produce a structured result type. The `--json` flag selects JSON encoding to stdout; the default selects human-readable formatting. Both paths consume the same result struct. `--dry-run` returns the same result type with a `planned` flag.

### AD5: Impl Pattern for External Calls

All direct OS/filesystem calls are isolated in `*Impl` functions:
- `ReadDirImpl`, `ReadFileImpl`, `WriteFileImpl`, `RenameImpl`, `RemoveImpl`
- `CreateExclusiveImpl` (for SID reservation with O_EXCL)
- `AcquireLockImpl`, `ReleaseLockImpl`

These are excluded from coverage by the existing coverage check script.

## Implementation Phases

### Phase 1: Foundation (Domain Types + Core Libraries)

Build the domain model and supporting libraries with no CLI integration.

1. **internal/domain/filename.go** — Filename regex parsing and generation
2. **internal/domain/node.go** — Node, MaterializedPath, Document types
3. **internal/domain/selector.go** — Selector parsing (MP vs SID disambiguation)
4. **internal/domain/numbering.go** — Sibling numbering (100/10/1 tiers), gap-finding
5. **internal/sid/sid.go** — Base62 SID generation with rejection sampling
6. **internal/slug/slug.go** — Unicode-aware slug generation
7. **internal/frontmatter/frontmatter.go** — YAML frontmatter split/parse/serialize

### Phase 2: Outline Tree + Validation

Build the outline model and read-only operations.

8. **internal/domain/outline.go** — Tree construction from parsed files
9. **internal/outline/service.go** — OutlineService with reader interface
10. **cmd/list.go** — `lmk list` (tree display, `--depth`, `--type`, `--json`)
11. **cmd/check.go** — `lmk check` (validation findings, `--json`, exit code 2)

### Phase 3: Mutating Operations (Core)

Build the primary write operations.

12. **internal/lock/lock.go** — Advisory locking (gofrs/flock wrapper)
13. **cmd/add.go** — `lmk add` (SID allocation, file creation, placement)
14. **cmd/rename.go** — `lmk rename` (frontmatter update, file rename)
15. **cmd/delete.go** — `lmk delete` (recursive, promote, confirmation)

### Phase 4: Advanced Operations

Build the remaining operations.

16. **cmd/move.go** — `lmk move` (subtree relocation, MP prefix rewrite)
17. **cmd/doctor.go** — `lmk doctor` (repair: slug drift, missing notes, unreserved SIDs)
18. **cmd/compact.go** — `lmk compact` (renumbering with 100-spacing)
19. **cmd/types.go** — `lmk types` (list/add/remove document types)

### Phase 5: Polish

20. **Global flags** — Wire `--json` and `--dry-run` flags across all commands
21. **Selector support** — Wire `mp:`/`sid:` prefix handling in all commands
22. **Error messages** — Consistent error formatting to stderr
23. **Cross-platform testing** — Verify on Linux, macOS, Windows
