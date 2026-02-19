# Linemark Glossary

Domain terms used throughout the project. This is the authoritative source for Ubiquitous Language.

| Term | Definition |
| ---- | ---------- |
| Node | A logical entry in the outline, identified by a Materialized Path and a SID. A node may have multiple documents. |
| Materialized Path (MP) | A sequence of three-digit, zero-padded integers separated by dashes (e.g., `001-200-010`) encoding ancestry and sibling order. |
| SID (Short ID) | An 8-12 character alphanumeric identifier assigned to a node. Unique within the repository, stable across moves and renames, and never reused. |
| Document | A Markdown file belonging to a node, distinguished by its document type (draft, notes, characters, etc.). |
| Document Type | A classification label for a document within a node (e.g., `draft`, `notes`). Appears in the filename. |
| Draft | The required document type containing the canonical title in YAML frontmatter. Every node must have one. |
| Outline | The complete tree of nodes within a single flat directory, representing the hierarchical structure of a prose project. |
| Slug | A URL-safe, kebab-case string derived from a node's title, used in filenames. |
| Control Directory | The `.linemark/` directory containing SID reservation markers (`ids/`) and the advisory lock file (`lock`). |
| SID Reservation | A permanent marker file (`.linemark/ids/<sid>`) that prevents SID reuse, even after node deletion. |
| Advisory Lock | A file-level lock (`.linemark/lock`) acquired by mutating commands to serialize concurrent CLI invocations. |
| Compact | The operation of renumbering materialized path segments within a subtree to restore even spacing for future insertions. |
| Selector | A reference to a node, specified as either a Materialized Path, a SID, or an explicit prefix (`mp:` or `sid:`). |
| Slug Drift | A validation finding where the filename slug no longer matches the slugified YAML title. |
| Sibling Spacing | The numbering strategy for siblings: initial spacing at multiples of 100, insertions at 10s then 1s. |
