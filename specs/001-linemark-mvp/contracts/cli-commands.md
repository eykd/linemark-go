# CLI Command Contracts: Linemark MVP

**Date**: 2026-02-19 | **Branch**: `001-linemark-mvp`

## Global Flags

All commands:
- `--json` — Structured JSON output to stdout
- `--verbose` / `-v` — Debug logging to stderr

All mutating commands additionally:
- `--dry-run` — Preview changes without executing

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (invalid args, lock failure, I/O error) |
| 2 | Validation findings detected (`check`, `doctor`) |

---

## `lmk add`

**Synopsis**: `lmk add [flags] <title>`

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--child-of` | selector | (none) | Add as child of specified node |
| `--sibling-of` | selector | (none) | Add as sibling of specified node |
| `--before` | selector | (none) | Place before specified sibling |
| `--after` | selector | (none) | Place after specified sibling |

**Behavior**:
- No flags: creates root-level node as last root sibling
- `--child-of`: appends as last child (unless `--before`/`--after`)
- `--sibling-of`: places immediately after reference (unless `--before`/`--after`)
- Allocates SID, creates `draft` (with YAML title) and `notes` files
- Acquires advisory lock

**JSON output** (`--json`):
```json
{
  "node": {
    "mp": "001-200",
    "sid": "A3F7c9Qx7Lm2",
    "title": "Chapter One"
  },
  "files_created": [
    "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md",
    "001-200_A3F7c9Qx7Lm2_notes.md"
  ]
}
```

**Dry-run JSON** (`--json --dry-run`):
```json
{
  "node": {
    "mp": "001-200",
    "sid": "(pending)",
    "title": "Chapter One"
  },
  "files_planned": [
    "001-200_(sid)_draft_chapter-one.md",
    "001-200_(sid)_notes.md"
  ]
}
```

---

## `lmk list`

**Synopsis**: `lmk list [flags]`

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--depth` | int | 0 (unlimited) | Limit tree display depth |
| `--type` | string | (none) | Filter by document type |

**Behavior**:
- Read-only, no lock acquired
- Default: indented tree view to stdout
- `--depth N`: show only N levels deep
- `--type T`: show only nodes containing document type T

**Human output**:
```
Overview (A3F7c9Qx7Lm2)
├── Part One (B8kQ2mNp4Rs1)
│   ├── Chapter 1 (C2xL9pQr5Tm3)
│   └── Chapter 2 (D4yM0rSt6Un4)
└── Part Two (E6zN1sUv7Wo5)
```

**JSON output** (`--json`):
```json
{
  "nodes": [
    {
      "mp": "001",
      "sid": "A3F7c9Qx7Lm2",
      "title": "Overview",
      "depth": 1,
      "types": ["draft", "notes"],
      "children": [
        {
          "mp": "001-100",
          "sid": "B8kQ2mNp4Rs1",
          "title": "Part One",
          "depth": 2,
          "types": ["draft", "notes"],
          "children": []
        }
      ]
    }
  ]
}
```

---

## `lmk move`

**Synopsis**: `lmk move <selector> --to <selector> [flags]`

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--to` | selector | Target parent node |
| `--before` | selector | Place before specified sibling |
| `--after` | selector | Place after specified sibling |

**Behavior**:
- Moves node and all descendants
- Updates MP prefixes in all affected filenames
- SIDs preserved
- Acquires advisory lock

**JSON output** (`--json`):
```json
{
  "renames": [
    {"old": "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", "new": "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md"},
    {"old": "001-200_A3F7c9Qx7Lm2_notes.md", "new": "002-100_A3F7c9Qx7Lm2_notes.md"}
  ]
}
```

---

## `lmk delete`

**Synopsis**: `lmk delete <selector> [flags]`

**Flags**:
| Flag | Short | Description |
|------|-------|-------------|
| `--recursive` | `-r` | Remove entire subtree |
| `--promote` | `-p` | Promote children to deleted node's position |
| `--force` | (none) | Skip interactive confirmation |

**Behavior**:
- Removes all files for target node
- `-r`: removes subtree recursively
- `-p`: removes node, promotes children (renumbers their MPs)
- Without `--force`: prompts for confirmation (requires TTY)
- SID reservation markers preserved
- Acquires advisory lock

**JSON output** (`--json`):
```json
{
  "files_deleted": [
    "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md",
    "001-200_A3F7c9Qx7Lm2_notes.md"
  ],
  "sids_preserved": ["A3F7c9Qx7Lm2"]
}
```

---

## `lmk rename`

**Synopsis**: `lmk rename <selector> <new-title>`

**Behavior**:
- Updates `title` in draft YAML frontmatter
- Regenerates slug in all filenames for that node
- SID and MP unchanged
- Acquires advisory lock

**JSON output** (`--json`):
```json
{
  "node": {
    "mp": "001-200",
    "sid": "A3F7c9Qx7Lm2",
    "old_title": "Chapter One",
    "new_title": "The Beginning"
  },
  "renames": [
    {"old": "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", "new": "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md"}
  ]
}
```

---

## `lmk types`

### `lmk types list <selector>`

**Behavior**: Read-only, lists document types for the specified node.

**JSON output**:
```json
{
  "node": {"mp": "001", "sid": "A3F7c9Qx7Lm2"},
  "types": ["draft", "notes", "characters"]
}
```

### `lmk types add <type> <selector>`

**Behavior**: Creates a new empty Markdown file of the specified type. Acquires advisory lock.

### `lmk types remove <type> <selector>`

**Behavior**: Deletes the file of the specified type. Cannot remove `draft`. Acquires advisory lock.

---

## `lmk check`

**Synopsis**: `lmk check [flags]`

**Behavior**:
- Read-only validation, no lock acquired
- Exit 0: no findings
- Exit 2: findings detected

**JSON output** (`--json`):
```json
{
  "findings": [
    {
      "type": "slug_drift",
      "severity": "warning",
      "message": "Filename slug 'chpter-one' does not match title slug 'chapter-one'",
      "path": "001-200_A3F7c9Qx7Lm2_draft_chpter-one.md"
    }
  ],
  "summary": {"errors": 0, "warnings": 1}
}
```

**Finding types**:
| Type | Severity | Description |
|------|----------|-------------|
| `invalid_filename` | error | File doesn't match canonical regex |
| `duplicate_sid` | error | Same SID used by multiple nodes |
| `slug_drift` | warning | Filename slug doesn't match YAML title |
| `missing_draft` | error | Node has no draft document |
| `missing_notes` | warning | Node has no notes document |
| `malformed_frontmatter` | error | Draft YAML frontmatter cannot be parsed |
| `orphaned_reservation` | warning | SID reservation with no content files |

---

## `lmk doctor`

**Synopsis**: `lmk doctor [flags]`

**Flags**:
| Flag | Description |
|------|-------------|
| `--apply` | Execute repairs (acquires advisory lock) |

**Behavior**:
- Without `--apply`: identical to `check` (report-only)
- With `--apply`: performs safe repairs:
  - Fix slug drift (rename files to match YAML title)
  - Reserve unreserved SIDs (create missing markers)
  - Create missing notes files
- Duplicate SIDs: reported, never auto-repaired

**JSON output** (`--json --apply`):
```json
{
  "repairs": [
    {
      "type": "slug_drift",
      "action": "rename",
      "old": "001-200_A3F7c9Qx7Lm2_draft_chpter-one.md",
      "new": "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md"
    }
  ],
  "unrepaired": [],
  "summary": {"repaired": 1, "unrepaired": 0}
}
```

---

## `lmk compact`

**Synopsis**: `lmk compact [<selector>] [flags]`

**Flags**:
| Flag | Description |
|------|-------------|
| `--apply` | Execute compaction (acquires advisory lock) |

**Behavior**:
- Without `--apply`: report-only, shows planned renames
- With `--apply`: renumbers positions with 100-spacing
- Warning emitted when >50 files affected
- SIDs never changed

**JSON output** (`--json`):
```json
{
  "renames": [
    {"old": "001-101_A3F7c_draft_x.md", "new": "001-200_A3F7c_draft_x.md"}
  ],
  "files_affected": 4,
  "warning": null
}
```
