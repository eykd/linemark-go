# Linemark Functional Specification (MVP) — SID Edition

## 1. Overview

**Purpose:**
Linemark is a command-line tool for managing hierarchical outlines of Markdown documents using filenames alone. It enables structured, sortable, and human-readable content organization within a flat directory.

**Goals:**
- Represent a full tree hierarchy in lexicographically sortable filenames.
- Support multi-document nodes (`draft`, `notes`, `characters`, etc.).
- Provide safe, intuitive commands for adding, moving, deleting, and listing nodes.
- Maintain compatibility with Git, Finder, VS Code, and Obsidian.
- Use no external database — filenames are the source of truth for structure and content.
- Be agent-native: structured JSON output, deterministic behavior, and CI-friendly exit codes enable reliable automation by scripts, pipelines, and AI agents.

**Guiding Principles:**
- Flat directory; hierarchy encoded in materialized paths.
- Predictable, human-readable filenames.
- Minimal configuration (CLI flags only in MVP).
- Version control expected for recovery and history.
- Deterministic behavior under concurrent CLI invocation.
- JSON-first output model; human-readable is a presentation layer.

---

## 2. Core Concepts

### Node
A logical entry in the outline. Identified by a **Materialized Path (MP)** and a **SID**. A node may have multiple Markdown files of different document types.

### Materialized Path (MP)
A sequence of three-digit, zero-padded integers separated by dashes, representing ancestry and sibling order.

Example:
```
001-200-010
```
- `001` → root index
- `200` → second-level index
- `010` → third-level index

### SID (Short ID)
A **SID** is a short, unique alphanumeric identifier assigned to a node. SIDs are stable across renames and moves.

**Format:**
```
[A-Za-z0-9]{8,12}
```
(Default recommendation: base62-encoded 72-bit cryptographically secure random value, typically ≤ 12 characters.)

**Properties:**
- Globally unique within the repository
- Never reused, even after node deletion
- Independent of Materialized Path and title

### Document Types
Each node may include multiple Markdown documents distinguished by type:

```
<mp>_<sid>_<doc-type>_<optional-title>.md
```

Required in MVP:
- `draft` (must contain YAML title)
- `notes` (may be empty)

Examples:
```
001_A3F7c9Qx7Lm2_draft_Overview.md
001_A3F7c9Qx7Lm2_notes.md
```

### Title Source
The canonical node title is stored in the YAML frontmatter of the `draft` document:

```yaml
---
title: Overview
---
```

The filename slug is derived from the YAML title during creation and rename operations.

### Title/Slug Invariant
- The YAML `title` field in the `draft` document is canonical for the node's display name.
- The filename slug MUST match `slugify(title)` at all times.
- Drift between the YAML title and the filename slug is a `check` finding, repaired by `doctor --apply`.
- Duplicate titles across nodes are allowed — SIDs guarantee uniqueness of identity.

---

## 3. File Naming & Structure

### Canonical Filename Pattern

```
<materialized-path>_<sid>_<doc-type>_<optional-title>.md
```

### Rules
- MP: three-digit segments (001–999) joined by `-`.
- SID: 8–12 character base62 string.
- Doc type: lowercase string (e.g., `draft`, `notes`, `characters`).
- Title: optional, slugified (kebab-case recommended).
- Directory: flat; all content files in one folder.
- Control directory: `.linemark/` for internal assets.

### Control Directory

```
.linemark/
  ids/
  lock
```

- `.linemark/ids/` contains one empty marker file per allocated SID.
- Each filename in `ids/` is the SID itself.
- These marker files are permanent and prevent SID reuse.
- `.linemark/lock` is an advisory lock file acquired by mutating commands to prevent concurrent modification (see Section 4a).

---

## 4. SID Allocation (No Counters, Concurrency-Safe)

Linemark generates SIDs using cryptographically secure randomness (default 72 bits) encoded in base62.

To guarantee uniqueness under concurrent invocations:

1. Generate random SID.
2. Attempt to create `.linemark/ids/<sid>` using exclusive-create semantics.
3. If creation succeeds, the SID is reserved.
4. If the file already exists, retry with a new SID.

Properties:
- No global counter.
- No external database.
- Safe under parallel CLI invocations.
- Reserved SIDs are never deleted.

---

## 4a. Advisory Locking

All mutating commands (`add`, `move`, `delete`, `rename`, `compact --apply`, `doctor --apply`) acquire an advisory lock on `.linemark/lock` before performing filesystem changes.

**Mechanism:**
- Uses `flock`-style advisory file locking (Go: `syscall.Flock` on Unix, cross-platform equivalent on Windows).
- Lock is acquired at the start of the mutating operation and held for its entire duration.
- Lock is released when the operation completes (success or failure).

**Behavior:**
- If the lock cannot be acquired (another mutating command is in progress): exit with error code `1` and a message identifying the lock file.
- Read-only commands (`list`, `check`, `types list`) do NOT acquire the lock.
- The lock is advisory — external tools are not prevented from modifying the directory, but concurrent `lmk` invocations are serialized.

---

## 5. Command Reference

### Global Flags

The following flags are available on all commands:

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | | Structured JSON output to stdout |
| `--verbose` | `-v` | Debug logging to stderr |

The following flag is available on all mutating commands (`add`, `move`, `delete`, `rename`, `compact`, `doctor --apply`):

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview changes without executing; outputs planned actions |

When `--json` is active, all output (including `--dry-run` previews) is structured JSON. Human-readable output is the default when `--json` is not specified.

---

### Node Selectors
Commands accept either a Materialized Path (MP) or a SID.

**MP selector pattern:**
```
^\d{3}(?:-\d{3})*$
```

**SID selector pattern:**
```
^[A-Za-z0-9]{8,12}$
```

**Disambiguation rule:**
- If input matches MP pattern → treated as MP.
- Else if input matches SID pattern → treated as SID.
- Otherwise → error.

Optional explicit selectors:
```
mp:<mp>
sid:<sid>
```

---

### `lmk add`
Create a new node.

```
lmk add [--child-of <selector> | --sibling-of <selector>] [--before <selector> | --after <selector>] <title>
```

Behavior:
- Default placement: last child (for `--child-of`) or immediately after (for `--sibling-of`).
- Allocates new SID via reservation system.
- Creates `draft` and `notes` files.
- Prints created filenames.
- `--json`: outputs created file list as JSON array.
- `--dry-run`: outputs planned files and SID without creating them.

---

### `lmk move`
Reorganize node hierarchy.

```
lmk move <selector> --to <selector> [--before|--after]
```

Behavior:
- Moves node and all descendants.
- Updates MP prefixes in all affected filenames.
- SIDs preserved.
- `--json`: outputs rename map (`{"old": "...", "new": "..."}`).
- `--dry-run`: outputs planned renames without executing.

---

### `lmk delete`
Delete a node and optionally descendants.

```
lmk delete <selector> [-r|--recursive] [-p|--promote] [--force]
```

Behavior:
- Deletes all files for target node.
- `-r` removes subtree.
- `-p` promotes children.
- Interactive confirmation unless `--force`.
- SID reservation marker remains (SID never reused).
- `--json`: outputs deleted file list as JSON array.
- `--dry-run`: outputs files that would be deleted without removing them.

---

### `lmk list`
Display the current hierarchy.

```
lmk list [--json] [--depth <n>] [--type <doc-type>]
```

Behavior:
- Default: human-friendly tree view.
- `--json`: outputs nested JSON structure with `children` arrays.
- `--depth <n>`: limits tree display to `n` levels deep.
- `--type <doc-type>`: filters to show only nodes containing the specified document type.

---

### `lmk rename`
Update node title and filenames.

```
lmk rename <selector> <new-title>
```

Behavior:
- Updates `title` in `draft` frontmatter.
- Regenerates slug in all related filenames.
- SID unchanged.
- `--json`: outputs old/new filename pairs.
- `--dry-run`: outputs planned renames without executing.

---

### `lmk types`
Manage document types within a node.

```
lmk types list <selector>
lmk types add <type> <selector>
lmk types remove <type> <selector>
```

Behavior:
- Lists, adds, or removes type files under node.
- Creates files empty in MVP.

---

### `lmk compact`
Reassigns tiered numbering within a subtree.

```
lmk compact [<selector>] [--apply]
```

Behavior:
- Default (no `--apply`): report-only; outputs planned renames without executing.
- `--apply`: executes the renumbering.
- Renumbers using 100/10/1 tier logic.
- SIDs unaffected.
- `--json`: outputs rename map.
- Warning emitted when more than 50 files would be affected.

---

### `lmk check`
Read-only validation of the outline.

```
lmk check [--json]
```

Checks:
- Invalid filename patterns (files that don't match the canonical regex).
- Duplicate SIDs across distinct nodes.
- Slug drift (filename slug does not match `slugify(title)` in YAML frontmatter).
- Missing required document types (e.g., node without `draft`).
- Malformed YAML frontmatter in `draft` files.
- Orphaned SID reservations (marker file in `.linemark/ids/` with no corresponding content file).

Behavior:
- Always read-only; never modifies files.
- Exit code `0`: no findings (outline is clean).
- Exit code `2`: one or more findings detected.
- `--json`: outputs structured diagnostics as JSON array, each entry with `type`, `severity`, `message`, and `path` fields.

---

### `lmk doctor`
Validate and repair the outline.

```
lmk doctor [--apply] [--json]
```

Checks (same as `lmk check`):
- Invalid filename patterns.
- Duplicate SIDs across distinct nodes (error — not auto-repairable).
- Slug drift.
- Missing required document types.
- Malformed YAML frontmatter.
- Unreserved SIDs (content file references a SID with no marker in `.linemark/ids/`).

Behavior:
- Default (no `--apply`): report-only, identical to `check`. Exit code `2` if findings exist.
- `--apply`: perform safe repairs:
  - Reserve unreserved SIDs (create missing `.linemark/ids/<sid>` markers).
  - Fix slug drift (rename files to match `slugify(title)` from YAML frontmatter).
  - Create missing `notes` files for nodes that lack them.
- Repairs are deterministic: the same input always produces the same output.
- `--json`: outputs repair plan (without `--apply`) or repair results (with `--apply`).
- Duplicate SIDs are reported but never auto-repaired — manual intervention required.

---

## 6. Exit Codes

All commands follow a consistent exit code convention:

| Code | Meaning |
|------|---------|
| `0` | Success — command completed without errors or findings |
| `1` | General error — invalid arguments, lock acquisition failure, I/O error, or other runtime failure |
| `2` | Validation findings — `check` or `doctor` detected issues in the outline |

Exit codes are stable and suitable for use in scripts, CI pipelines, and agent workflows.

---

## 7. Parsing & Internal Model

### Filename Regex

```
^(\d{3}(?:-\d{3})*)_([A-Za-z0-9]{8,12})_([a-z]+)(?:_(.*))?\.md$
```

Parsed record:

```go
type ParsedFile struct {
    MP        string   // "001-200-010"
    SID       string   // "A3F7c9Qx7Lm2"
    DocType   string   // "draft"
    Slug      string   // "overview"
    PathParts []string // ["001", "200", "010"]
    Depth     int      // 3
}
```

---

## 8. Behavioral Rules

### Numbering
- Three-digit segments (001–999).
- Initial sibling spacing: multiples of 100.
- Insertions fill 10s, then 1s.
- `compact` condenses numbering.

### SID Rules
- Exactly one SID per node.
- SID appears in all files belonging to that node.
- SID never changes.
- SID reservation marker never deleted.

### Node Creation
- Allocates SID.
- Creates required document types.
- Does not launch editor.

### Moves
- Cascade renames based on MP prefix changes.
- SIDs preserved.

### Deletes
- Removes files only.
- Leaves SID reservation intact.

### Determinism Guarantees
- Tree construction from directory listing is deterministic: the same set of files always produces the same tree.
- Sort order: byte-wise ASCII sort on filenames (not locale-dependent).
- No hidden state affects hierarchy beyond filenames and the `.linemark/` control directory.
- Filesystem case sensitivity: filenames are treated as case-sensitive. On case-insensitive filesystems (e.g., macOS HFS+/APFS default), users must avoid titles that differ only in case.

### Structural Limits
- MP segment range: 001–999 (maximum 999 siblings per parent).
- No hard depth limit; practical recommendation is 20 levels or fewer.
- Overflow: when no gap exists for insertion at the target position, the command fails with an error and a suggestion to run `compact`.

### Scale and Performance
- All commands parse the directory on each invocation — O(n) where n is the file count in the content directory.
- Expected scale: up to 10,000 files per directory.
- No caching in MVP; `.linemark/cache` is reserved for future extension.

---

## 9. Dependencies & Library Recommendations

| Purpose | Library / Package | Notes |
|----------|-------------------|-------|
| CLI framework | `github.com/spf13/cobra` | Subcommand management |
| Random bytes | `crypto/rand` | Cryptographically secure randomness |
| Base62 encoding | custom or lightweight lib | Encode random bytes to alphanumeric |
| YAML parsing | `gopkg.in/yaml.v3` | Frontmatter handling |
| Slug creation | custom or `gosimple/slug` | Title slugging |
| JSON output | `encoding/json` | Structured output with `json.Marshal` |
| File handling | `os` + `path/filepath` | Cross-platform filesystem ops |
| Regex parsing | `regexp` | Filename tokenization |
| Advisory locking | `syscall` or cross-platform flock | File-level advisory locking |

---

## 10. Future Extensions

- Query subsystem (`lmk query <expression>`) for filtering and searching nodes.
- Link rewriting on move/rename (update internal Markdown links when nodes are reorganized).
- Full plan/apply separation (`lmk plan`, `lmk apply`) for complex multi-step operations.
- `.linemark/cache` for directory parsing performance at scale.
- Git-aware operations.
- Templates for document types.
- Subtree export/import.

---

**End of SID Edition Specification**
