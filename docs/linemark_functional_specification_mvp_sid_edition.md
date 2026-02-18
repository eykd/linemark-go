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

**Guiding Principles:**
- Flat directory; hierarchy encoded in materialized paths.
- Predictable, human-readable filenames.
- Minimal configuration (CLI flags only in MVP).
- Version control expected for recovery and history.
- Deterministic behavior under concurrent CLI invocation.

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
```

- `.linemark/ids/` contains one empty marker file per allocated SID.
- Each filename in `ids/` is the SID itself.
- These marker files are permanent and prevent SID reuse.

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

## 5. Command Reference

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

---

### `lmk list`
Display the current hierarchy.

```
lmk list [--json]
```

Behavior:
- Default: human-friendly tree view.
- `--json`: outputs nested JSON structure with `children` arrays.

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
lmk compact [<selector>]
```

Behavior:
- Default: whole outline.
- Renumbers using 100/10/1 tier logic.
- Prints summary of renames.
- SIDs unaffected.

---

### `lmk doctor`
Validate and repair the outline.

```
lmk doctor [--apply]
```

Checks:
- Invalid filename patterns.
- Duplicate SIDs across distinct nodes (error).
- Unreserved SID referenced in filenames → reserve it.
- Missing required document types.
- Malformed YAML frontmatter.

Behavior:
- Default: report-only.
- `--apply`: perform safe repairs.

---

## 6. Parsing & Internal Model

### Filename Regex

```
^(\d{3}(?:-\d{3})*)_([A-Za-z0-9]{8,12})_([a-z]+)(?:_(.*))?\.md$
```

Parsed record:

```python
{
  'mp': '001-200-010',
  'sid': 'A3F7c9Qx7Lm2',
  'type': 'draft',
  'slug': 'overview',
  'path_parts': ['001', '200', '010'],
  'depth': 3
}
```

---

## 7. Behavioral Rules

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

---

## 8. Dependencies & Library Recommendations

| Purpose | Library | Notes |
|----------|----------|-------|
| CLI framework | `click` | Subcommand management |
| Random bytes | `secrets` | Cryptographically secure randomness |
| Base62 encoding | custom or lightweight lib | Encode random bytes |
| YAML parsing | `pyyaml` | Frontmatter handling |
| Slug creation | `python-slugify` | Title slugging |
| JSON output | built-in `json` | Pretty-print with indent=2 |
| File handling | `pathlib` | Filesystem ops |
| Regex parsing | built-in `re` | Filename tokenization |

---

## 9. Future Extensions

- Transaction planning (`--dry-run`, `--json` diffs)
- Query subsystem
- Git-aware operations
- Templates for document types
- Subtree export/import

---

**End of SID Edition Specification**

