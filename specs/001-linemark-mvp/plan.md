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

## Security Considerations

### Path Traversal Protection

All file operations must validate that resolved paths remain within the content directory and `.linemark/` control directory. Specifically:
- When reading a directory listing, resolve each file path and reject any that escape the content root (e.g., via `../` segments in crafted filenames)
- `filepath.Rel()` or equivalent must confirm the file is a descendant of the expected root before any read/write/rename
- Symlinks in the content directory should not be followed; use `os.Lstat` instead of `os.Stat` when scanning, and skip symlinks with a validation finding

### YAML Frontmatter Injection

The `title` field is user-provided and stored in YAML frontmatter. Mitigations:
- Always use `yaml.Node.SetString()` when writing the title, which ensures proper quoting and prevents multiline injection (e.g., a title like `"foo\nnew_key: value"` would be safely quoted)
- When reading, use `.Value` on the specific node — never unmarshal the entire frontmatter into `map[string]interface{}` to avoid type confusion (e.g., `title: true` being parsed as bool)
- The yaml.Node approach already selected in R3 handles this correctly; document this as an explicit security property in the frontmatter package

### Lock File Permissions

The `.linemark/lock` file should be created with permissions `0666` (before umask) to allow any user who can write to the content directory to also acquire the lock. Document that advisory locking only protects against other `lmk` processes — it does not protect against direct filesystem manipulation by editors or scripts.

### SID Collision Handling

Although collision probability is negligible (~2^-35.7 per pair at 71 bits of entropy), the SID allocation code must:
- Retry with a new random SID on `O_EXCL` failure (EEXIST), up to a bounded number of attempts (e.g., 10)
- Fail with a clear error if retries are exhausted (not a silent data corruption)
- Never fall back to a weaker generation method

## Edge Cases & Error Handling

### Filename Length Limits

Filenames can exceed filesystem limits (255 bytes on ext4/HFS+/NTFS) with long titles or deep nesting. Each depth level adds 4 characters (`-NNN`), and the base format `<mp>_<sid>_<type>_<slug>.md` requires at minimum ~30 characters before the slug.

Mitigations:
- After computing a filename, check total byte length against a 255-byte limit
- If the filename exceeds the limit, truncate the slug (not the MP, SID, or type) and append a truncation marker (e.g., trailing `--`)
- Emit a warning when slug truncation occurs
- The `compact` and `move` commands must also check resulting filename lengths

### Empty Slug from Special-Character Titles

When a title consists entirely of special characters (e.g., `"!!!"`, `"—"`), the slug becomes empty. Per the filename regex, the slug portion is optional (`(?:_(.*))?`). Validate that:
- The filename format `<mp>_<sid>_<type>.md` (no slug) is handled correctly by the parser
- Renaming a node from a title-with-slug to a no-slug title works (filename component removed)
- `lmk check` does not report an empty slug as an error

### Cycle Detection in Move

`lmk move` must detect and reject moves that would create a cycle — specifically, moving a node to be a descendant of itself. Before executing:
- Check if the target parent's MP starts with the source node's MP
- If so, fail with: `"cannot move node <MP> to its own descendant <target-MP>"`
- This check must happen before any file renames begin

### Sibling Slot Exhaustion

With 001-999 range and 100/10/1 tier insertion:
- When all 999 positions are genuinely occupied (not just gaps), fail with a clear error: `"maximum 999 siblings reached at <parent-MP>"`
- When tier-1 gaps are exhausted but positions remain unused elsewhere, suggest `compact` to redistribute
- `compact` cannot help when all 999 slots are truly used — the error message must distinguish these cases

### Promote With Insufficient Gaps

`lmk delete --promote` moves children to the deleted node's sibling level. If the deleted node has N children but fewer than N sibling gaps are available:
- Compute required gaps before starting the operation
- Fail early with a clear message: `"not enough sibling positions to promote N children; run compact first"`
- Never leave children partially promoted

### Control Directory Bootstrap

On first mutating command, the `.linemark/` directory and `.linemark/ids/` subdirectory may not exist:
- Create them atomically (`os.MkdirAll`) before attempting lock acquisition or SID reservation
- Handle the race where two concurrent first-run commands both try to create the directory (MkdirAll is idempotent for existing dirs)
- `lmk check` and `lmk list` (read-only commands) should work even when `.linemark/` doesn't exist (empty outline)

### Partial Failure and Rollback

Multi-file operations (`move`, `compact`, `delete --promote`, `rename`) must handle partial failure:
- **Strategy**: Collect all planned renames into a list, validate all preconditions upfront, then execute renames in order. If any rename fails:
  - Attempt to reverse already-completed renames (best-effort rollback)
  - If rollback also fails, emit a clear error listing the current state of all files so the user can recover manually
  - Include a `lmk check` suggestion in the error output
- **`lmk add` cleanup**: If the draft file is created but notes file creation fails, delete the draft file and release the SID reservation
- Document that `lmk doctor` can repair most inconsistencies left by interrupted operations

### Disk Full / Permission Errors

- Before multi-file writes, check available disk space is sufficient (or at minimum, handle ENOSPC gracefully per-file)
- For permission errors, check write access to the content directory at operation start rather than discovering failures mid-operation
- All error messages must include the specific file path that failed and the OS error

### Files Modified During Operation

Advisory locking only protects against concurrent `lmk` processes. If a user or editor modifies files between the directory scan and rename execution:
- Accept this as a known limitation (documented in `--verbose` output)
- If a rename fails because the source file no longer exists, report the error but continue with remaining renames
- `lmk check` serves as the recovery tool for any resulting inconsistencies

## Performance Considerations

### Directory Scan Optimization

Every invocation reads the full directory. For 10,000 files:
- Use `os.ReadDir` (not `os.ReadFile` on each file) to get file names without reading content
- Only parse frontmatter when the command actually needs titles (e.g., `lmk check` for slug validation, `lmk rename`). `lmk list` can display titles from filename slugs for performance, with frontmatter as the authoritative source only when `--json` includes a `canonical_title` field
- Compile the filename regex once (package-level `var`) rather than per-file

### Large Subtree Operations

Moving or compacting a subtree with hundreds of descendants:
- Hold the advisory lock for the minimum duration necessary — compute all renames before acquiring the lock, then acquire, execute renames, and release
- Emit progress to stderr when `--verbose` is set and more than 50 files are affected
- The >50 file warning for `compact` (from spec) should also apply to `move` operations

### SID Reservation Directory Growth

`.linemark/ids/` accumulates marker files permanently. Over time, this could reach thousands of files:
- This is acceptable for MVP — directory listing of small files is fast on modern filesystems
- Document this behavior so users know the directory grows monotonically
- A future `lmk gc` command (out of scope for MVP) could prune reservations with no corresponding content files
