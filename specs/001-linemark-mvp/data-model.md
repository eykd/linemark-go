# Data Model: Linemark MVP

**Date**: 2026-02-19 | **Branch**: `001-linemark-mvp`

## Entities

### Node

A logical entry in the outline. Has no persistent representation beyond its constituent files and SID reservation.

| Field | Type | Description | Rules |
|-------|------|-------------|-------|
| MP | MaterializedPath | Position in the hierarchy | Sequence of 3-digit zero-padded ints joined by `-` |
| SID | SID | Stable unique identifier | 12-char base62, never changes, never reused |
| Title | string | Canonical display name | Stored in draft frontmatter YAML `title` field |
| Documents | []Document | Files belonging to this node | Must include `draft` and `notes` at minimum |

**Invariants**:
- SID is assigned once and never changes across moves or renames
- Title/Slug invariant: `slug(node.Title) == filename slug` at all times
- A Node always has at least a `draft` and `notes` document

### MaterializedPath

A value object encoding ancestry and sibling order.

| Property | Value |
|----------|-------|
| Format | `\d{3}(-\d{3})*` (e.g., `001`, `001-200`, `001-200-010`) |
| Segment range | 001–999 |
| Depth | Number of segments (1-based) |
| Parent | All segments except the last (empty for root nodes) |
| Sort order | Byte-wise ASCII (lexicographic = hierarchical) |

**Numbering rules**:
- Initial sibling spacing: multiples of 100 (100, 200, 300, ...)
- Insertions: fill at 10s (110, 120, ...), then 1s (101, 102, ...)
- When no gap exists: fail with suggestion to run `compact`

### SID (Short ID)

A value object providing stable identity.

| Property | Value |
|----------|-------|
| Format | `[A-Za-z0-9]{12}` (base62) |
| Entropy | ~71.45 bits (12 × log₂(62)) |
| Generation | `crypto/rand` with rejection sampling |
| Reservation | `.linemark/ids/<sid>` marker file (O_EXCL create) |
| Lifetime | Permanent — never deleted, never reused |

### Document

A Markdown file belonging to a node.

| Field | Type | Description |
|-------|------|-------------|
| Type | string | Document classification: `draft`, `notes`, or custom |
| Filename | string | Full canonical filename |
| Content | string | Raw file content (Markdown with optional frontmatter) |

**Required types**: Every node must have `draft` (with YAML frontmatter containing `title`) and `notes`.

### ParsedFile

The result of parsing a filename against the canonical pattern.

| Field | Type | Description |
|-------|------|-------------|
| MP | string | Materialized path (e.g., `001-200-010`) |
| SID | string | Short ID (e.g., `A3F7c9Qx7Lm2`) |
| DocType | string | Document type (e.g., `draft`) |
| Slug | string | Title slug from filename (e.g., `overview`) |
| PathParts | []string | Individual segments (e.g., `["001", "200", "010"]`) |
| Depth | int | Number of segments |

**Filename regex**: `^(\d{3}(?:-\d{3})*)_([A-Za-z0-9]{8,12})_([a-z]+)(?:_(.*))?\.md$`

### Outline

The complete tree of nodes. Not persisted — reconstructed from directory listing on each invocation.

| Property | Description |
|----------|-------------|
| Nodes | All nodes parsed from the content directory |
| Tree | Hierarchical view derived from MaterializedPaths |
| Root Dir | The flat directory containing all content files |
| Control Dir | `.linemark/` directory for SID reservations and lock |

**Determinism guarantee**: Same set of files always produces same tree (byte-wise ASCII sort on filenames).

### Selector

A reference to a node, specified by the user.

| Variant | Pattern | Example |
|---------|---------|---------|
| MP | `^\d{3}(?:-\d{3})*$` | `001-200` |
| SID | `^[A-Za-z0-9]{8,12}$` | `A3F7c9Qx7Lm2` |
| Explicit MP | `mp:<mp>` | `mp:001-200` |
| Explicit SID | `sid:<sid>` | `sid:A3F7c9Qx7Lm2` |

**Disambiguation**: If ambiguous, MP pattern takes precedence over SID.

### Finding (Validation)

A diagnostic result from `check` or `doctor`.

| Field | Type | Description |
|-------|------|-------------|
| Type | string | Finding category (e.g., `invalid_filename`, `duplicate_sid`, `slug_drift`) |
| Severity | string | `error` or `warning` |
| Message | string | Human-readable description |
| Path | string | Affected file path |

## State Transitions

### Node Lifecycle

```
[Created] → add → [Active] → rename/move → [Active] → delete → [Deleted]
                                                                    ↓
                                                          SID reservation persists
```

### Sibling Numbering

```
Empty parent → add first child → 100
Add second child → 200
Add third child → 300
Insert between 100 and 200 → 110 (10s tier)
Insert between 100 and 110 → 101 (1s tier)
No gap available → ERROR: suggest compact
Compact → renumber all siblings at 100-spacing
```

## Relationships

```
Outline 1──* Node
Node    1──* Document
Node    1──1 MaterializedPath
Node    1──1 SID
Node    1──1 SID Reservation (.linemark/ids/<sid>)
Document ──1 ParsedFile (via filename parsing)
```

## Control Directory Structure

```
.linemark/
├── ids/
│   ├── A3F7c9Qx7Lm2    # SID reservation markers (empty files)
│   ├── B8kQ2mNp4Rs1
│   └── ...
└── lock                  # Advisory lock file (flock)
```
