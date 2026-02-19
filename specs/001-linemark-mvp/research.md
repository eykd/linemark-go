# Research: Linemark MVP

**Date**: 2026-02-19 | **Branch**: `001-linemark-mvp`

## R1: Advisory File Locking

**Decision**: Use `github.com/gofrs/flock`

**Rationale**: The Go standard library has no public advisory locking API (`cmd/go/internal/lockedfile/internal/filelock` is internal). `gofrs/flock` is the de facto standard (704 stars, used by Helm), provides cross-platform support (Linux `fcntl`/OFD locks, macOS `fcntl`, Windows `LockFileEx`), and includes `TryLock()`/`TryLockContext()` for fail-fast CLI behavior.

**Key design decisions**:
- Use `TryLock()` (non-blocking, immediate fail) for CLI — users should get an immediate error, not a hang
- Create `.linemark/lock` file on first mutating command; never delete it
- Always call `Unlock()` explicitly via `defer` — do not rely on process exit
- Wrap in `*Impl` function for OS-level operations

**Alternatives rejected**:
| Option | Why rejected |
|--------|-------------|
| `syscall.Flock` directly | No Windows support, no goroutine safety, requires build tags |
| Vendored `cmd/go/internal/lockedfile` | Internal package, no stability guarantee |
| `danjacques/gofslock` | Less maintained, no `TryLockContext` |
| Custom with `golang.org/x/sys` | Would reimplement what `gofrs/flock` already provides |

---

## R2: SID Generation (Base62)

**Decision**: Custom implementation using `crypto/rand` + per-byte rejection sampling

**Rationale**: No external library generates base62 random IDs directly. A custom implementation is ~30 lines, zero dependencies, provably unbiased, and 10x faster than `math/big`-based approaches.

**Key design decisions**:
- 12-character base62 strings (71.45 bits of entropy, meeting ~72 bit target)
- Rejection sampling threshold: `maxByte = 248` (62 × 4), bytes ≥ 248 discarded
- Alphabet: `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789`
- `crypto/rand.Read` is goroutine-safe — no mutex needed in caller
- Reservation via `O_EXCL` create of `.linemark/ids/<sid>` (Impl pattern)

**Alternatives rejected**:
| Option | Why rejected |
|--------|-------------|
| `math/big` modulo conversion | ~10x slower, 2 heap allocations |
| `jxskiss/base62` | External dep, variable-length output, encoder only |
| `byte % 62` (no rejection) | ~21% relative bias on 8/62 positions |
| `crypto/rand.Text()` (Go 1.24+) | Base32 only, wrong alphabet, 26-char output |

---

## R3: YAML Frontmatter Parsing

**Decision**: String splitting + `gopkg.in/yaml.v3` Node API for round-trip fidelity

**Rationale**: Struct-based unmarshal loses unknown fields on write-back. `map[string]interface{}` loses field order. The `yaml.Node` tree preserves all fields, order, comments, and quoting style. Frontmatter splitting is a text-level concern handled before YAML parsing.

**Key design decisions**:
- Split on `---\n` prefix, find closing `\n---\n` or `\n---` at EOF
- Parse interior with `yaml.Unmarshal` into `yaml.Node` (not struct)
- Update title via `node.Content[i+1].SetString(title)` to ensure `!!str` tag
- Serialize with `yaml.NewEncoder` + `SetIndent(2)`, reassemble: `"---\n" + yaml + "---\n" + body`
- Body is raw string, never touched by YAML parser

**Gotchas addressed**:
- `yaml.v3` does not guarantee byte-identical round-trips (quoting may change)
- `title: true` parses as `!!bool`; use `.Value` for reading, `SetString()` for writing
- Empty frontmatter `---\n---\n` handled as special case in splitter

**Alternatives rejected**:
| Option | Why rejected |
|--------|-------------|
| Struct-based unmarshal | Loses unknown fields on write-back |
| `map[string]interface{}` | Loses field order and comments |
| Regex manipulation | Brittle with multiline values |
| Third-party frontmatter libs | Most use `map[string]interface{}` internally |

---

## R4: Slug Generation

**Decision**: Custom implementation using `golang.org/x/text/unicode/norm` + `golang.org/x/text/transform`

**Rationale**: NFD normalization + Mn-mark stripping is the canonical Go pattern for diacritics removal. Custom implementation is ~30 lines, deterministic, and avoids the mutable global state in `gosimple/slug`. `golang.org/x/text` is maintained by the Go team and has no transitive dependencies.

**Key design decisions**:
- NFD decompose → strip `unicode.Mn` marks → NFC recompose → lowercase → whitespace to `-` → strip non-`[a-z0-9-]` → collapse multiple `-` → trim
- Pure function: no global state, no randomness
- Empty/all-special-chars input returns `""`
- Safe for Git, Finder, VS Code, and Obsidian (all unsafe chars stripped)

**Alternatives rejected**:
| Option | Why rejected |
|--------|-------------|
| `gosimple/slug` | 2 extra deps, mutable global state, MPL-2.0 license |
| Pure stdlib (hand-coded char map) | Misses uncommon precomposed forms, requires ongoing maintenance |
| NFKD instead of NFD | Minor improvement; can switch later if needed |

---

## Summary of New Dependencies

| Package | Purpose | License |
|---------|---------|---------|
| `gopkg.in/yaml.v3` | YAML frontmatter parsing | MIT/Apache-2.0 |
| `github.com/gofrs/flock` | Cross-platform advisory file locking | BSD-3-Clause |
| `golang.org/x/text` | Unicode normalization for slug generation | BSD-3-Clause |

All existing: `github.com/spf13/cobra` (CLI framework)
