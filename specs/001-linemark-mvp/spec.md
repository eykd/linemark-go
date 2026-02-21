# Specification: Linemark MVP

**Feature Branch**: `001-linemark-mvp`
**Created**: 2026-02-18
**Status**: Draft
**Beads Epic**: `linemark-go-4fx`

**Beads Phase Tasks**:
- clarify: `linemark-go-4fx.1`
- plan: `linemark-go-4fx.2`
- red-team: `linemark-go-4fx.3`
- tasks: `linemark-go-4fx.4`
- analyze: `linemark-go-4fx.5`
- implement: `linemark-go-4fx.6`
- security-review: `linemark-go-4fx.7`
- architecture-review: `linemark-go-4fx.8`
- code-quality-review: `linemark-go-4fx.9`

---

## 1. Problem Statement

Writers, worldbuilders, and long-form content creators need a way to organize hierarchical outlines of Markdown documents that works seamlessly with their existing tools — Git, VS Code, Obsidian, and Finder. Current solutions either require a database, lock users into proprietary formats, or lack the structural rigor needed for large projects.

There is no lightweight, file-based system that encodes a full tree hierarchy directly in filenames while remaining human-readable, sortable, and safe for concurrent use.

## 2. Actors

| Actor | Description |
| ----- | ----------- |
| Author | A writer or content creator managing a long-form prose project (novel, worldbuilding bible, documentation suite) |
| Automation Agent | A script, CI pipeline, or AI agent that programmatically manages the outline |

## 3. User Scenarios & Acceptance Criteria

### US1: Initialize a New Outline

**As an** Author
**I want to** create the first node in a new outline
**So that** I can begin structuring my prose project

**Acceptance Criteria:**
- Given an empty directory, when the author adds a node with a title, then two Markdown files are created: a draft file containing the title in YAML frontmatter and an empty notes file.
- Given a new node is created, then a unique stable identifier (SID) is assigned and reserved permanently.
- Given a new node is created, then the filenames are human-readable, containing the hierarchy position, SID, document type, and slugified title.

### US2: Build a Hierarchical Outline

**As an** Author
**I want to** add nodes as children or siblings of existing nodes
**So that** I can build a multi-level outline (e.g., Part > Chapter > Scene)

**Acceptance Criteria:**
- Given an existing node, when the author adds a child node, then the new node's position reflects its parent-child relationship in the filename.
- Given existing sibling nodes, when the author adds a node between them, then the new node sorts lexicographically between its neighbors.
- Given the author specifies `--before` or `--after` placement, then the new node is positioned accordingly among siblings.
- Given no explicit placement, then new children are appended as the last child and new siblings are placed immediately after the reference node.

### US3: View the Outline as a Tree

**As an** Author
**I want to** see my outline displayed as an indented tree
**So that** I can understand the structure of my project at a glance

**Acceptance Criteria:**
- Given a populated outline, when the author lists nodes, then a human-readable indented tree is displayed showing titles and hierarchy.
- Given the `--depth` flag is used, then only nodes up to the specified depth are shown.
- Given the `--type` filter is used, then only nodes containing the specified document type are shown.
- Given the `--json` flag is used, then a nested JSON structure with `children` arrays is output.

### US4: Move Nodes Within the Outline

**As an** Author
**I want to** reorganize nodes by moving them (with their subtrees) to new positions
**So that** I can restructure my outline as the project evolves

**Acceptance Criteria:**
- Given a node with descendants, when the author moves it to a new parent, then the node and all descendants are repositioned.
- Given a move operation, then all SIDs remain unchanged — only positions change.
- Given a move operation, then all affected filenames are updated atomically.
- Given `--dry-run`, then planned renames are displayed without execution.

### US5: Delete Nodes from the Outline

**As an** Author
**I want to** remove nodes I no longer need
**So that** I can keep my outline clean and focused

**Acceptance Criteria:**
- Given a leaf node, when the author deletes it, then all files belonging to that node are removed.
- Given a node with children and `--recursive`, then the node and its entire subtree are removed.
- Given a node with children and `--promote`, then the node is removed and its children are promoted to its former position.
- Given a delete without `--force`, then interactive confirmation is required.
- Given any deletion, then the SID reservation is preserved (SID is never reused).

### US6: Rename a Node

**As an** Author
**I want to** change the title of an existing node
**So that** I can refine my outline's naming as the project develops

**Acceptance Criteria:**
- Given an existing node, when the author provides a new title, then the YAML frontmatter title is updated and all related filenames are re-slugified.
- Given a rename, then the node's SID and position remain unchanged.

### US7: Manage Document Types

**As an** Author
**I want to** add or remove document types (e.g., characters, locations) within a node
**So that** I can attach different kinds of content to the same outline entry

**Acceptance Criteria:**
- Given an existing node, when the author lists types, then all document types for that node are displayed.
- Given an existing node, when the author adds a new type, then a new empty Markdown file of that type is created.
- Given an existing node, when the author removes a type, then the corresponding file is deleted.

### US8: Validate the Outline

**As an** Author
**I want to** check my outline for structural problems
**So that** I can catch and fix issues before they compound

**Acceptance Criteria:**
- Given an outline with no issues, then validation reports success with exit code 0.
- Given an outline with issues (invalid filenames, duplicate SIDs, slug drift, missing required document types, malformed frontmatter, orphaned SID reservations), then validation reports each finding with exit code 2.
- Given `--json`, then findings are structured with type, severity, message, and path fields.
- Validation never modifies any files.

### US9: Repair the Outline

**As an** Author
**I want to** automatically fix common outline problems
**So that** I can maintain a healthy outline without manual file manipulation

**Acceptance Criteria:**
- Given slug drift (filename doesn't match YAML title), when `--apply` is used, then files are renamed to match the canonical title.
- Given unreserved SIDs (content exists but SID marker is missing), when `--apply` is used, then the missing markers are created.
- Given missing notes files, when `--apply` is used, then empty notes files are created.
- Given duplicate SIDs, then they are reported but never auto-repaired.
- Without `--apply`, doctor behaves identically to check (report-only).

### US10: Compact Numbering

**As an** Author
**I want to** reorganize the numbering in my outline when gaps become unwieldy
**So that** filenames remain clean and well-spaced for future insertions

**Acceptance Criteria:**
- Given an outline with numbering gaps, when compact is run with `--apply`, then positions are renumbered with even spacing.
- Given compact without `--apply`, then planned renames are reported without execution.
- Given more than 50 files would be affected, then a warning is emitted.
- SIDs are never changed during compaction.

### US11: Structured Output for Automation

**As an** Automation Agent
**I want to** receive structured JSON output from all commands
**So that** I can reliably parse and act on the results programmatically

**Acceptance Criteria:**
- Given `--json` on any command, then output is valid JSON to stdout.
- Given `--dry-run` with `--json`, then planned changes are output as structured JSON.
- Exit codes are consistent: 0 = success, 1 = error, 2 = validation findings.

### US12: Concurrent Safety

**As an** Automation Agent
**I want to** run multiple `lmk` commands safely in parallel
**So that** automated pipelines don't corrupt the outline

**Acceptance Criteria:**
- Given two mutating commands run concurrently, then only one proceeds and the other fails with a clear error message and exit code 1.
- Given a read-only command and a mutating command run concurrently, then the read-only command is not blocked.
- Given a mutating command fails, then the advisory lock is released.

### US13: Node Selection by SID or Position

**As an** Author
**I want to** reference nodes by either their position or their stable identifier
**So that** I can use whichever is more convenient in context

**Acceptance Criteria:**
- Given a position pattern (e.g., `001-200`), then the node at that position is selected.
- Given a SID (e.g., `A3F7c9Qx7Lm2`), then the node with that identifier is selected.
- Given an ambiguous input, then position patterns take precedence over SID patterns.
- Given explicit prefixes (`mp:` or `sid:`), then the specified interpretation is used regardless of pattern matching.

## 4. Functional Requirements

### FR1: File-Based Hierarchy
The system must represent a full tree hierarchy using only filenames within a single flat directory. No external database or index file is used — filenames are the source of truth.

### FR2: Canonical Filename Format
Every content file must follow the pattern `<materialized-path>_<sid>_<doc-type>_<optional-title>.md` where the materialized path is a sequence of three-digit zero-padded integers separated by dashes.

### FR3: Stable Identity
Each node must have a unique SID (8-12 alphanumeric characters) that never changes across moves or renames and is never reused, even after deletion.

### FR4: Required Document Types
Every node must include at minimum a `draft` document (containing the canonical title in YAML frontmatter) and a `notes` document.

### FR5: Title/Slug Invariant
The YAML `title` field in the draft document is canonical. The filename slug must always match the slugified title. Drift is a validation finding.

### FR6: Sibling Numbering
Initial sibling spacing uses multiples of 100. Insertions fill gaps at 10s, then 1s. When no gap exists, the command fails with a suggestion to run `compact`.

### FR7: Advisory Locking
All mutating commands must acquire an advisory file lock before making changes. The lock is held for the duration of the operation and released on completion or failure.

### FR8: Deterministic Behavior
The same set of files must always produce the same tree. Sort order is byte-wise ASCII on filenames. No hidden state affects hierarchy beyond filenames and the control directory.

### FR9: Consistent Exit Codes
Exit code 0 means success, 1 means general error, 2 means validation findings detected.

### FR10: JSON-First Output
All commands must support `--json` for structured output. Human-readable display is a presentation layer over the same data.

### FR11: Dry-Run Preview
All mutating commands must support `--dry-run` to preview changes without executing them.

### FR12: Structural Limits
Position segments range from 001-999 (maximum 999 siblings per parent). No hard depth limit is enforced.

## 5. Non-Functional Requirements

### NFR1: Performance
All commands must complete in reasonable time for outlines up to 10,000 files. The directory is parsed on each invocation.

### NFR2: Compatibility
The system must produce filenames compatible with Git, macOS Finder, VS Code, and Obsidian. Filenames are treated as case-sensitive.

### NFR3: No External Dependencies at Runtime
The tool requires no database, server, or network connection. It operates entirely on the local filesystem.

### NFR4: Portability
The tool must work on macOS, Linux, and Windows.

## 6. Success Criteria

| Criterion | Measure |
| --------- | ------- |
| Authors can build a 3-level outline from scratch | An author creates a Part > Chapter > Scene hierarchy in under 2 minutes using only `lmk add` commands |
| Outline survives restructuring | Moving a subtree preserves all content and identifiers; no data loss occurs |
| Concurrent safety holds | Two simultaneous mutating commands never corrupt the outline |
| Automation agents can drive the tool | A script using `--json` output can parse all command responses and chain operations without human intervention |
| Validation catches all specified problems | `lmk check` detects every category of finding listed in US8 with zero false negatives |
| Repair is safe and deterministic | Running `lmk doctor --apply` twice in a row produces no changes on the second run |
| Files remain human-readable | An author unfamiliar with the tool can read filenames and understand the hierarchy |
| Tool integrates with version control | All operations produce clean Git diffs; no binary files or hidden state outside `.linemark/` |

## 7. Key Entities

| Entity | Description |
| ------ | ----------- |
| Node | A logical entry in the outline, identified by a Materialized Path and a SID. Contains one or more documents. |
| Materialized Path (MP) | A sequence of three-digit zero-padded integers (e.g., `001-200-010`) encoding ancestry and sibling order. |
| SID (Short ID) | An 8-12 character alphanumeric stable identifier, unique within the repository and never reused. |
| Document | A Markdown file belonging to a node, distinguished by type (draft, notes, characters, etc.). |
| Draft Document | The required document type containing the canonical title in YAML frontmatter. |
| Outline | The complete set of nodes forming a tree hierarchy within a single flat directory. |
| Control Directory | The `.linemark/` directory containing SID reservation markers and the advisory lock file. |

## 8. Scope & Boundaries

### In Scope
- All commands described above: add, move, delete, rename, list, check, doctor, compact, types
- Flat-directory hierarchy via materialized paths
- SID allocation with concurrency-safe reservation
- Advisory locking for mutating commands
- JSON and human-readable output modes
- Dry-run preview for all mutating commands
- YAML frontmatter for canonical titles

### Out of Scope
- Query/search subsystem
- Link rewriting on move/rename
- Git-aware operations (auto-commit, branch management)
- Templates for document types
- Subtree export/import
- Editor integration (no automatic editor launch)
- GUI or web interface
- Configuration files (CLI flags only)
- Caching or performance optimization beyond direct directory scanning

## 9. Assumptions

- The user has Git installed and uses it for version history and recovery.
- The content directory contains only Markdown files managed by `lmk`. Unrecognized files are ignored by all commands.
- The filesystem supports exclusive-create semantics (O_EXCL) for SID reservation.
- The filesystem supports advisory file locking (flock).
- Users on case-insensitive filesystems will avoid titles that differ only in case.
- Each outline lives in a single flat directory; nested directories are not used for content.

## 10. Dependencies

- Filesystem with advisory locking and exclusive-create support
- YAML frontmatter parsing capability
- Cryptographically secure random number generation for SID creation

## Interview

### Open Questions

_(No open questions — specification derived from approved functional specification document.)_

### Answer Log

| # | Question | Answer | Date |
| - | -------- | ------ | ---- |
| 1 | Source of requirements | Derived from `docs/linemark_functional_specification_mvp_sid_edition.md` — the approved functional specification | 2026-02-18 |
