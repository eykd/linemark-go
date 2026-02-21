# ATDD in This Repository: A Comprehensive Guide

Acceptance Test-Driven Development (ATDD) is the outermost development loop used in this project. It anchors every feature to a specification written in the language of the user's world — before a line of implementation code is written — and the feature is not considered complete until observable behavior satisfies that specification. This guide explains every layer of the system from spec authorship through automated verification.

---

## Table of Contents

1. [Why ATDD?](#why-atdd)
2. [Two Test Streams](#two-test-streams)
3. [GWT Spec Format](#gwt-spec-format)
4. [Domain Language Discipline](#domain-language-discipline)
5. [The Acceptance Pipeline](#the-acceptance-pipeline)
6. [Binding Test Stubs](#binding-test-stubs)
7. [The Outer Loop: ATDD Cycle](#the-outer-loop-atdd-cycle)
8. [The Inner Loop: TDD Cycle](#the-inner-loop-tdd-cycle)
9. [Task-to-Spec Linkage](#task-to-spec-linkage)
10. [Definition of Done](#definition-of-done)
11. [Spec Quality: The Spec Guardian](#spec-quality-the-spec-guardian)
12. [Command Reference](#command-reference)
13. [File and Directory Layout](#file-and-directory-layout)
14. [Worked Example](#worked-example)

---

## Why ATDD?

Unit tests verify internal correctness — that the code does what the programmer intended. They answer the question *How does it work?* Acceptance tests verify observable behavior — that the system does what the user needs. They answer the question *What does it do?*

Without acceptance tests, a feature can pass all unit tests and still fail its purpose. The developer's model of what a feature should do can drift silently from what the user actually needs. ATDD closes this gap by making the user story executable: before writing the first unit test, you write a plain-language description of what success looks like from the outside, and you run that description as a test.

The most important property of this system: **acceptance tests define done**. A task is closed not when the implementer believes the work is complete, but when the acceptance tests pass. Human judgment is still used for code quality and refactoring, but the binary question of whether a user story is satisfied is answered by a test, not by an opinion.

---

## Two Test Streams

Two independent test streams run in parallel throughout development:

| Stream | Question answered | Location | Runs with |
|---|---|---|---|
| **Acceptance tests** | *What does the system do?* (user-observable behavior) | `generated-acceptance-tests/` | `just acceptance` |
| **Unit tests** | *How does it do it?* (internal correctness) | Alongside production code | `just test` |

Both streams must pass before a task is closed. They are complementary, not alternatives: unit tests catch regressions at the level of individual functions and packages; acceptance tests catch regressions at the level of end-to-end user scenarios.

Run both together:

```
just test-all
```

---

## GWT Spec Format

Acceptance criteria are written as Given-When-Then (GWT) specs — plain text files that a non-developer can read and understand. They use a structured format derived from Uncle Bob's Fitnesse conventions.

### File Naming

```
specs/US<N>-<kebab-case-title>.txt
```

One file per user story. The `US<N>` prefix must match the user story number used in the task tracker.

### Format

```
;===============================================================
; Scenario: Description of the scenario.
;===============================================================
GIVEN precondition.

WHEN action.

THEN observable outcome.
THEN another observable outcome.
```

**Separator lines** consist only of `;` and `=` characters. They delimit scenario headers.

**Description lines** start with `;` (but are not separators). The line between the opening and closing separator is the scenario's title. Additional `;` lines after the title and before the closing separator are ignored comments.

**Step keywords** — `GIVEN`, `WHEN`, `THEN` — appear in ALL CAPS at the start of a line.

**Every step line ends with a period.**

**Blank lines** between scenarios and between steps are allowed and ignored by the parser.

### Multi-Scenario Files

A single spec file covers all scenarios for one user story. Separate scenarios with a new separator block:

```
;===============================================================
; Scenario: Happy path.
;===============================================================
GIVEN the system is in a ready state.
WHEN the user performs the primary action.
THEN the expected result appears.


;===============================================================
; Scenario: Error case.
;===============================================================
GIVEN the system is missing required data.
WHEN the user performs the primary action.
THEN an error message explains the problem.
```

Each scenario in a file becomes its own independent executable test.

---

## Domain Language Discipline

This is the hardest part of writing good specs, and the most important. Specs must use the vocabulary of the **user's domain**, not the vocabulary of the implementation.

A useful test: could a non-technical domain expert read this spec and understand it? If the answer is no, the spec contains implementation leakage and must be rewritten.

### What Leakage Looks Like

| Category | Examples of leaking language |
|---|---|
| Code references | function names, type names, package paths, variable names |
| Infrastructure | HTTP, REST, database, SQL, endpoint, cache, server |
| Framework language | handler, middleware, context, goroutine, channel |
| Technical protocols | JSON, gRPC, WebSocket, TCP |
| Data structures | array, map, slice, struct, interface |
| File system details | `.yaml`, `.json`, config file, exact file path |
| Programming concepts | nil, null, exception, return value, callback |

### Good vs Bad

**Describing preconditions**

Good:
```
GIVEN a project with two chapters named "Introduction" and "Rising Action".
```

Bad:
```
GIVEN a directory containing intro.md and rising-action.md files.
```

**Describing actions**

Good:
```
WHEN the author lists the outline.
```

Bad:
```
WHEN exec.Command invokes the list subcommand.
```

**Describing outcomes**

Good:
```
THEN the output shows both chapters in order.
```

Bad:
```
THEN stdout contains the JSON-encoded chapter array.
```

**Describing errors**

Good:
```
THEN an error message says the project has no content.
```

Bad:
```
THEN the process exits with code 1 and stderr prints "no chapters found".
```

### Spec Review Checklist

Before committing a spec file:

1. File name matches `specs/US<N>-<kebab-case-title>.txt`
2. Separator syntax uses only `;` and `=` characters
3. Keywords are ALL CAPS (`GIVEN`, `WHEN`, `THEN`)
4. Every step ends with a period
5. No implementation language — no code identifiers, no infrastructure terms
6. Scenarios are independent — each can run without the others
7. Outcomes are observable — describe what the user sees, not internal state
8. Error cases are covered — include a scenario for error preconditions
9. Bootstrap scenarios exist — if a command can be the first interaction, spec the cold-start path

Run `/spec-check` to automate leakage detection on committed specs.

---

## The Acceptance Pipeline

The acceptance pipeline is a three-stage transformation that turns GWT spec files into executable tests:

```
specs/*.txt  →  [parse]  →  IR JSON  →  [generate]  →  generated-acceptance-tests/*_test.go  →  [run]
```

### Stage 1: Parse

The parser reads each `specs/*.txt` file and produces an Intermediate Representation (IR) in JSON format under `acceptance-pipeline/ir/`. The IR captures the structured data from each spec: the feature source path, scenario descriptions, and the ordered list of GIVEN/WHEN/THEN steps.

```
just acceptance-parse
```

### Stage 2: Generate

The generator reads the IR and produces test source files under `generated-acceptance-tests/`. Each scenario becomes one test function. Generated functions are named after their scenario description.

```
just acceptance-generate
```

**Key behavior: bound implementations are preserved.** The generator distinguishes between two states for each test function:

- **Unbound**: Contains the sentinel `t.Fatal("acceptance test not yet bound")`. This is a scaffold that must be filled in.
- **Bound**: Does not contain the sentinel. This is a real test implementation.

When the generator runs on a file that already exists, it extracts all bound functions from the existing file and preserves them. Only unbound stubs are regenerated. This means you can safely re-run the generator after editing a spec — your test implementations survive.

Orphaned bound functions (functions in the existing file with no corresponding scenario in the current spec) are preserved with a warning comment rather than deleted. This prevents accidental data loss when a spec is renamed or a scenario is reworded.

### Stage 3: Run

The runner executes the generated tests. Unbound stubs fail immediately (by design). Bound tests verify the implemented behavior.

```
just acceptance-run
```

### All Three Stages at Once

```
just acceptance
```

This runs the full pipeline: clears the IR artifacts, parses, generates, and runs. It does **not** destroy existing bound implementations.

### Force Regenerate (Destructive)

```
just acceptance-regen
```

This deletes `generated-acceptance-tests/` and `acceptance-pipeline/ir/` before running the pipeline. **All bound implementations are destroyed.** Use this only when you need to reset the test files completely — for example, after significantly restructuring a spec.

---

## Binding Test Stubs

After generation, each test function is a stub: it contains step comments that map the spec's GIVEN/WHEN/THEN steps into commented placeholders, followed by the unbound sentinel. The BIND step replaces the sentinel with a real implementation.

### What a Generated Stub Looks Like

```go
// Author creates the first node in a new outline.
// Source: specs/US01-initialize-new-outline.txt:2
func Test_Author_creates_the_first_node_in_a_new_outline(t *testing.T) {
    // GIVEN an empty content directory.
    // WHEN the author adds a node titled "My Novel".
    // THEN two files are created in the content directory.
    // THEN one file is a draft containing the title "My Novel".
    // THEN the other file is an empty notes document.

    t.Fatal("acceptance test not yet bound")
}
```

### The Binding Pattern

Binding follows the structure of the spec steps directly:

- Each **GIVEN** step → set up the precondition (create files, initialize the system)
- Each **WHEN** step → execute the action under test (run the CLI command)
- Each **THEN** step → assert the observable outcome (check output, check filesystem state)

Acceptance tests are **black-box tests**: they treat the system as an opaque binary and interact with it only through its public interface (the command line, in this project). They never import internal packages.

In this project, acceptance tests:
1. Build the binary once in `TestMain` (shared across all tests in the suite)
2. Create an isolated temporary directory per test
3. Run the binary with subprocess execution
4. Assert on the captured output and filesystem state

Helper functions shared across all acceptance tests live in `generated-acceptance-tests/helpers_test.go` and `generated-acceptance-tests/main_test.go`. These are hand-maintained files and are not touched by the pipeline.

### Editing Generated Files Directly

Edit the generated files in `generated-acceptance-tests/` directly. There is no separate file for your implementations — you replace the sentinel in the generated file. The pipeline's merge logic ensures your implementations survive subsequent regeneration runs (as long as the scenario description in the spec hasn't changed, which would change the function name).

If a scenario is renamed in the spec:
1. The old function becomes an orphan (preserved with a warning comment)
2. A new unbound stub is generated with the new name
3. You must manually migrate the implementation to the new function name and delete the orphan

---

## The Outer Loop: ATDD Cycle

The ATDD outer loop is implemented in `ralph.sh` as `execute_atdd_cycle`. It wraps the inner TDD cycle with acceptance-test verification, creating a two-level feedback structure:

```
OUTER LOOP (acceptance-driven):
  1. Check if acceptance tests already pass → done if yes
  2. BIND: write acceptance test implementations from spec
  3. INNER LOOP (unit-test-driven, up to 15 cycles):
     a. RED:      write smallest failing unit test
     b. GREEN:    write minimal code to pass that test
     c. REFACTOR: improve the code without breaking tests
     d. Check acceptance tests → done if they pass
  4. If inner loop exhausted → mark task BLOCKED
```

### Step 1: Initial Acceptance Check

Before doing any work, the cycle checks whether the acceptance tests already pass. This handles the case where a previous session made partial progress or where acceptance tests were already written and the implementation happens to satisfy them.

### Step 2: BIND

If acceptance tests are not yet passing, the BIND step fires. This is a focused Claude invocation with a specific mission: read the spec file and the generated stubs, then replace every unbound sentinel with a real implementation.

The BIND step runs **before** the inner TDD loop starts. This is intentional: the acceptance tests must be executable (even though they will fail) before unit tests are written. The acceptance test is the objective definition of what must be built. Writing it first ensures the inner TDD loop is steered by the right target.

BIND does not verify that the bound tests pass. Failing acceptance tests are expected at this point — the feature hasn't been built yet. The verification happens after each inner TDD cycle.

### Step 3: Inner TDD Cycles

After BIND, the inner TDD loop runs up to 15 times. After each full RED-GREEN-REFACTOR cycle, the acceptance tests are checked. If they pass, the task is closed. If they fail, the loop continues with the accumulated context of what's still missing.

The limit of 15 inner cycles (vs 5 for pure TDD tasks) reflects the reality that an acceptance test may require several units of functionality to pass — each inner cycle delivers one unit.

### Step 4: Exhaustion

If all 15 inner cycles complete without the acceptance tests passing, the task is marked BLOCKED with a comment explaining what happened. This is a signal for a human to intervene — either to adjust the spec, break the task into smaller pieces, or diagnose a deeper problem.

### Fallback for Tasks Without Specs

If `find_spec_for_task` cannot locate a matching spec file for the task (determined by extracting `US<N>` from the task title), the ATDD cycle falls back to a standard unit TDD cycle. This backward compatibility means the ATDD infrastructure does not require every task to have a spec — only user story tasks need specs. Infrastructure and support tasks can use plain TDD.

---

## The Inner Loop: TDD Cycle

The inner TDD cycle implements Red-Green-Refactor at the unit test level. Each step is a separate Claude invocation with a focused mission. The cycle runs up to 5 times per task for pure TDD, or up to 15 times when nested inside the ATDD outer loop.

### RED: Write a Failing Test

The RED step writes the smallest possible unit test that exercises some aspect of the feature. The test must fail before any new production code is written. This confirms the test is actually testing something.

Constraints:
- One test at a time (the smallest increment)
- The test must target the correct behavior, not just fail for a trivial reason
- The test must pass quality gates after the step (vet, lint, format)

### GREEN: Write Minimal Code to Pass

The GREEN step writes the minimum production code needed to make the RED test pass. "Minimum" is taken seriously: no extra features, no generalization, no handling of cases not covered by the currently-failing test.

Constraints:
- All existing tests must still pass
- Only the tests that were RED before this step may be turned GREEN
- Code is committed after GREEN succeeds (creating a stable checkpoint)

If GREEN produces no commit (because no changes were made), a fallback commit is created automatically.

### REFACTOR: Improve Without Changing Behavior

The REFACTOR step improves code quality without changing observable behavior. This is where duplication is removed, names are improved, and abstractions emerge.

Constraints:
- All tests must still pass after refactoring
- If REFACTOR breaks tests, changes are reverted automatically to preserve the GREEN state
- The refactored result must pass quality gates

### Verification Between Steps

After each step (except REVIEW), `ralph.sh` runs the test suite as a spot-check. If the suite fails, the step is retried up to 3 times with the test output included in the retry context. If all retries fail, the step is marked failed and the task is marked BLOCKED.

### REVIEW: Evaluate Completeness

After each full RED-GREEN-REFACTOR cycle, a REVIEW step evaluates whether the task is complete. The reviewer writes a structured JSON verdict to `.ralph-review.json`:

```json
{
  "complete": true,
  "reason": "All requirements implemented with comprehensive test coverage",
  "remaining_items": [],
  "test_gaps": []
}
```

If `complete` is `false`, the remaining items are passed as context into the next RED step. This steers the next cycle toward what's still missing.

In the ATDD outer loop, the REVIEW step is not the primary completion gate — the acceptance test check after each cycle serves that role. The REVIEW still runs and feeds its remaining items into the next cycle's context.

---

## Task-to-Spec Linkage

The automation infrastructure links tasks to specs by convention: a task whose title contains `US<N>` is matched to `specs/US<N>-*.txt`.

For example, a task titled `"US03: View the Outline as a Tree"` will match `specs/US03-view-outline-as-tree.txt`. The match is case-sensitive on `US` and the numeric suffix.

Tasks without a `US<N>` prefix do not get matched to a spec and fall back to the standard TDD cycle. This is by design: not every task is a user story. Infrastructure tasks, bug fixes, and refactoring tasks often don't have corresponding GWT specs.

If you need a task to use the ATDD loop, ensure:
1. The task title contains `US<N>` (e.g., `US07:`)
2. A corresponding spec file exists at `specs/US<N>-<any-name>.txt`
3. The spec file contains at least one scenario with GIVEN/WHEN/THEN steps

---

## Definition of Done

A task is **done** when:

1. The acceptance tests for the corresponding spec pass (objective gate)
2. All unit tests pass (100% coverage of non-exempt functions)
3. All quality gates pass (format, vet, lint)
4. The bead is closed in the task tracker

The acceptance test gate is the primary arbiter. It is evaluated mechanically by running the pipeline — no human judgment is applied to whether "the spirit of the spec" is satisfied. This is intentional: vague completion criteria are a major source of scope creep and integration failures.

If the acceptance tests are poorly written (testing the wrong thing, or testing implementation details rather than behavior), they will provide a false signal. This is why domain language discipline in specs, and correctness in the BIND step, both matter deeply.

---

## Spec Quality: The Spec Guardian

The spec guardian is an AI agent that audits spec files for implementation leakage. It reads every GIVEN, WHEN, and THEN statement and checks for the categories of leaking language described in the [Domain Language Discipline](#domain-language-discipline) section.

### Running the Audit

Audit all specs:
```
/spec-check
```

Audit a specific file:
```
/spec-check specs/US03-view-outline-as-tree.txt
```

The guardian outputs a table of findings with the file, line number, original statement, leakage category, and a suggested domain-language rewrite. If no violations are found, it reports "All specs use clean domain language."

### When to Run It

- After writing a new spec (before binding)
- After editing a spec in response to clarification
- Before closing a user story epic
- Periodically as a quality audit

### Pre-commit Reminder

A pre-edit hook reminds developers to check whether a spec exists before beginning implementation work. This is a reminder, not a gate — you can proceed without a spec, but the prompt encourages the habit of spec-first development.

---

## Command Reference

### Pipeline

| Command | Description |
|---|---|
| `just acceptance` | Full pipeline: parse specs, generate tests, run tests |
| `just acceptance-parse` | Parse `specs/*.txt` to IR JSON only |
| `just acceptance-generate` | Generate test stubs/files from IR only |
| `just acceptance-run` | Run the generated acceptance tests only |
| `just acceptance-regen` | Force-regenerate all stubs (destroys bound implementations) |
| `just test-all` | Run both unit tests and acceptance tests |

### Spec Quality

| Command | Description |
|---|---|
| `/spec-check` | Audit all specs for implementation leakage |
| `/spec-check <path>` | Audit a specific spec file |

### Pipeline CLI (direct)

```
go run ./acceptance/cmd/pipeline -action=parse
go run ./acceptance/cmd/pipeline -action=generate
go run ./acceptance/cmd/pipeline -action=run
```

---

## File and Directory Layout

```
linemark-go/
├── specs/
│   ├── US01-initialize-new-outline.txt    # GWT spec for US01
│   ├── US02-build-hierarchical-outline.txt
│   └── ...                               # One .txt file per user story
│
├── acceptance/                            # Acceptance pipeline source (tested, not acceptance tests themselves)
│   ├── types.go                           # Step, Scenario, Feature types
│   ├── parser.go                          # GWT spec → Feature (pure function)
│   ├── ir.go                              # Feature → IR JSON (pure function)
│   ├── generator.go                       # IR → Go test source, with merge logic
│   └── cmd/pipeline/main.go              # CLI: parse / generate / run actions
│
├── acceptance-pipeline/
│   └── ir/                               # Generated IR JSON (not committed)
│
├── generated-acceptance-tests/           # Generated + bound test files (committed)
│   ├── main_test.go                      # TestMain: build binary once per suite
│   ├── helpers_test.go                   # Shared test helper functions
│   ├── US01-initialize-new-outline_test.go
│   └── ...                              # One _test.go per spec file
│
└── ralph.sh                             # Automation loop: ATDD + TDD orchestration
```

### What Is and Isn't Committed

| Path | Committed? | Notes |
|---|---|---|
| `specs/*.txt` | Yes | Source of truth; authored by humans |
| `acceptance-pipeline/ir/` | No | Ephemeral; regenerated each run |
| `generated-acceptance-tests/*_test.go` | Yes | Contains bound implementations; must be committed |
| `generated-acceptance-tests/main_test.go` | Yes | Hand-maintained; not touched by pipeline |
| `generated-acceptance-tests/helpers_test.go` | Yes | Hand-maintained; not touched by pipeline |

The decision to commit generated acceptance test files (rather than gitignoring them) is deliberate: bound implementations represent real work and belong in version control. Losing them to a stale regeneration would destroy human-authored test logic.

---

## Worked Example

This section traces a complete user story from spec through passing acceptance tests.

### 1. Write the Spec

A new user story: the author can view the outline as a tree.

Create `specs/US03-view-outline-as-tree.txt`:

```
;===============================================================
; Author views an outline with two top-level nodes.
;===============================================================
GIVEN a project with two nodes titled "Part One" and "Part Two".

WHEN the author lists the outline.

THEN the output shows both nodes.
THEN "Part One" appears before "Part Two".


;===============================================================
; Author views an empty outline.
;===============================================================
GIVEN a project with no nodes.

WHEN the author lists the outline.

THEN the output indicates the outline is empty.
```

Run `/spec-check specs/US03-view-outline-as-tree.txt` to confirm no leakage.

### 2. Generate Stubs

```
just acceptance-generate
```

This creates `generated-acceptance-tests/US03-view-outline-as-tree_test.go` with two unbound stubs.

### 3. BIND: Write Implementations

Replace each sentinel with a real test. For the first scenario (language-neutral description):

- **GIVEN**: Create a temp directory; initialize the project; add two nodes with the correct titles.
- **WHEN**: Run the list command; capture its output.
- **THEN**: Assert the output contains both node titles; assert the first title appears at a lower index than the second.

For the second scenario:

- **GIVEN**: Create a temp directory with an initialized but empty project.
- **WHEN**: Run the list command; capture its output.
- **THEN**: Assert the output contains a message indicating emptiness.

After binding, run `just acceptance-run`. Both tests fail — the feature isn't built yet. This is the expected state.

### 4. Create a Beads Task

```
bd create --title="US03: View the Outline as a Tree" \
  --description="Implement the list command to display the outline hierarchy." \
  --type=feature --priority=1
```

The `US03` prefix in the title means `ralph.sh` will automatically locate `specs/US03-view-outline-as-tree.txt` and use the ATDD outer loop.

### 5. Run the ATDD Loop

```
./ralph.sh --epic <epic-id>
```

Ralph picks up the task, checks acceptance (fails), runs BIND (already done manually in this example), then begins inner TDD cycles: RED writes a unit test for the list functionality, GREEN implements the minimal list logic, REFACTOR cleans up. After each cycle, Ralph runs the acceptance tests.

The loop closes the task automatically when both acceptance tests pass.

### 6. Verify Manually

```
just test-all
```

Both unit tests and acceptance tests should be green.

---

## Summary of Key Principles

1. **Specs before code.** Write the GWT spec before writing any implementation code. The spec is the contract.

2. **Domain language only.** A non-technical user must understand every GIVEN, WHEN, and THEN statement. Any implementation language in a spec is a defect.

3. **Acceptance tests define done.** A task is not complete because the implementer believes it is complete. It is complete when the acceptance tests pass.

4. **BIND before RED.** In the ATDD loop, acceptance tests are bound (given real implementations) before the first unit test is written. This ensures the inner TDD loop is oriented toward the right target.

5. **Preserve bound implementations.** Never use `acceptance-regen` casually. The generated test files contain human-authored logic and are committed to version control.

6. **Two streams, always.** Unit tests and acceptance tests are not alternatives. Unit tests verify the internals; acceptance tests verify the observable surface. Both must be green.
