---
description: Cross-artifact consistency and quality analysis across spec.md, plan.md, and beads tasks after task generation. Automatically creates remediation beads tasks under the current epic for any issues found.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Identify inconsistencies, duplications, ambiguities, and underspecified items across spec.md, plan.md, and beads tasks before implementation. This command MUST run only after `/sp:05-tasks` has successfully produced tasks in beads.

**Additionally**, this command queries beads for task status and progress metrics.

## Operating Constraints

**WRITE REMEDIATION TASKS**: Do **not** edit source files directly in this command, but you **MUST** create remediation beads tasks under the current epic for any issues you find. Do not require explicit user approval to add beads remediation tasks.

Task creation rules:

- For each finding, first search existing beads tasks under the epic (open + in_progress) to avoid duplicates.
- Only create a new task if no existing task clearly covers the same issue.
- Each created task MUST:
  - Be parented to the current epic
  - Include a concise title starting with `Remediate:`
  - Include a description that cites: impacted artifact(s), severity, and a concrete fix suggestion
  - Include acceptance criteria
- Assign priority by severity (CRITICAL → p1, MAJOR → p2, MINOR → p3).

You MUST still output a structured analysis report, and include a section listing the remediation tasks you created (IDs + titles).

**Constitution Authority**: The project constitution (`.specify/memory/constitution.md`) is **non-negotiable** within this analysis scope. Constitution conflicts are automatically CRITICAL and require adjustment of the spec, plan, or tasks—not dilution, reinterpretation, or silent ignoring of the principle. If a principle itself needs to change, that must occur in a separate, explicit constitution update outside `/sp:06-analyze`.

## Execution Steps

### 1. Initialize Analysis Context

Run `.specify/scripts/bash/check-prerequisites.sh --json` once from repo root and parse JSON for FEATURE_DIR and AVAILABLE_DOCS. Derive absolute paths:

- SPEC = FEATURE_DIR/spec.md
- PLAN = FEATURE_DIR/plan.md

Abort with an error message if spec.md or plan.md is missing (instruct the user to run missing prerequisite command).
For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

### 2. Retrieve Beads Epic and Task Status

a. Get the epic ID from spec.md:

```bash
grep "Beads Epic" FEATURE_DIR/spec.md | grep -oE 'workspace-[a-z0-9]+|bd-[a-z0-9]+'
```

b. Query beads for task statistics:

```bash
npx bd stats --json
```

c. List all tasks for this feature:

```bash
npx bd list --parent <epic-id> --json
```

d. Get ready tasks:

```bash
npx bd ready --json
```

e. View dependency tree:

```bash
npx bd dep tree <epic-id>
```

### 3. Load Artifacts (Progressive Disclosure)

Load only the minimal necessary context from each artifact:

**From spec.md:**

- Overview/Context
- Functional Requirements
- Non-Functional Requirements
- User Stories
- Edge Cases (if present)

**From plan.md:**

- Architecture/stack choices
- Data Model references
- Phases
- Technical constraints

**From beads tasks** (via `npx bd list --parent <epic-id> --json`):

- Task IDs and titles
- Task descriptions (containing spec refs, skills, acceptance criteria)
- Task hierarchy (parent-child relationships)
- Task status (open, in_progress, closed)
- Dependencies
- Progress metrics

**From constitution:**

- Load `.specify/memory/constitution.md` for principle validation

### 4. Build Semantic Models

Create internal representations (do not include raw artifacts in output):

- **Requirements inventory**: Each functional + non-functional requirement with a stable key (derive slug based on imperative phrase; e.g., "User can upload file" → `user-can-upload-file`)
- **User story/action inventory**: Discrete user actions with acceptance criteria
- **Task coverage mapping**: Map each beads task to one or more requirements or stories using:
  - Explicit `**Spec**:` references in task descriptions
  - Task titles matching user story patterns (e.g., "US1:", "US-1")
  - Keyword inference from task descriptions
- **Constitution rule set**: Extract principle names and MUST/SHOULD normative statements
- **Beads task status**: Map task IDs to their current status (open, in_progress, closed)

### 5. Detection Passes (Token-Efficient Analysis)

Focus on high-signal findings. Limit to 50 findings total; aggregate remainder in overflow summary.

#### A. Duplication Detection

- Identify near-duplicate requirements
- Mark lower-quality phrasing for consolidation

#### B. Ambiguity Detection

- Flag vague adjectives (fast, scalable, secure, intuitive, robust) lacking measurable criteria
- Flag unresolved placeholders (TODO, TKTK, ???, `<placeholder>`, etc.)

#### C. Underspecification

- Requirements with verbs but missing object or measurable outcome
- User stories missing acceptance criteria alignment
- Tasks referencing files or components not defined in spec/plan

#### D. Constitution Alignment

- Any requirement or plan element conflicting with a MUST principle
- Missing mandated sections or quality gates from constitution

#### E. Coverage Gaps

- Requirements with zero associated tasks
- Tasks with no mapped requirement/story
- Non-functional requirements not reflected in tasks (e.g., performance, security)

#### F. Inconsistency

- Terminology drift (same concept named differently across files)
- Data entities referenced in plan but absent in spec (or vice versa)
- Task ordering contradictions (e.g., integration tasks before foundational setup tasks without dependency note)
- Conflicting requirements (e.g., one requires Next.js while other specifies Vue)

### 6. Severity Assignment

Use this heuristic to prioritize findings:

- **CRITICAL**: Violates constitution MUST, missing core spec artifact, or requirement with zero coverage that blocks baseline functionality
- **HIGH**: Duplicate or conflicting requirement, ambiguous security/performance attribute, untestable acceptance criterion
- **MEDIUM**: Terminology drift, missing non-functional task coverage, underspecified edge case
- **LOW**: Style/wording improvements, minor redundancy not affecting execution order

### 7. Produce Compact Analysis Report

Output a Markdown report (no file writes) with the following structure:

## Specification Analysis Report

| ID  | Category    | Severity | Location(s)      | Summary                      | Recommendation                       |
| --- | ----------- | -------- | ---------------- | ---------------------------- | ------------------------------------ |
| A1  | Duplication | HIGH     | spec.md:L120-134 | Two similar requirements ... | Merge phrasing; keep clearer version |

(Add one row per finding; generate stable IDs prefixed by category initial.)

**Coverage Summary Table:**

| Requirement Key | Has Task? | Task IDs | Notes |
| --------------- | --------- | -------- | ----- |

**Constitution Alignment Issues:** (if any)

**Unmapped Tasks:** (if any)

### 8. Beads Task Status Report

Include a section showing beads task progress:

**Beads Progress Summary:**

| Metric      | Count |
| ----------- | ----- |
| Total Tasks | X     |
| Open        | X     |
| In Progress | X     |
| Closed      | X     |
| Ready Now   | X     |

**Task Hierarchy:**

```text
[Include output from bd dep tree]
```

**Ready Tasks:**

- List tasks currently available for work from `bd ready`

**Blocked Tasks:**

- List tasks waiting on dependencies

**Metrics:**

- Total Requirements
- Total Tasks
- Coverage % (requirements with >=1 task)
- Ambiguity Count
- Duplication Count
- Critical Issues Count
- **Beads Completion %** (closed / total tasks)

### 9. Provide Next Actions

At end of report, output a concise Next Actions block:

- If CRITICAL issues exist: Recommend resolving before `/sp:07-implement`
- If only LOW/MEDIUM: User may proceed, but provide improvement suggestions
- Provide explicit command suggestions: e.g., "Run /sp:01-specify with refinement", "Run /sp:03-plan to adjust architecture", "Manually edit tasks.md to add coverage for 'performance-metrics'"
- **If tasks are ready in beads**: Suggest running `/sp:07-implement` to start work

### 10. Offer Remediation

Ask the user: "Would you like me to suggest concrete remediation edits for the top N issues?" (Do NOT apply them automatically.)

## Operating Principles

### Context Efficiency

- **Minimal high-signal tokens**: Focus on actionable findings, not exhaustive documentation
- **Progressive disclosure**: Load artifacts incrementally; don't dump all content into analysis
- **Token-efficient output**: Limit findings table to 50 rows; summarize overflow
- **Deterministic results**: Rerunning without changes should produce consistent IDs and counts

### Analysis Guidelines

- **NEVER modify files or beads tasks** (this is read-only analysis)
- **NEVER hallucinate missing sections** (if absent, report them accurately)
- **Prioritize constitution violations** (these are always CRITICAL)
- **Use examples over exhaustive rules** (cite specific instances, not generic patterns)
- **Report zero issues gracefully** (emit success report with coverage statistics)

## Beads Commands Reference

| Action          | Command                                 |
| --------------- | --------------------------------------- |
| Get statistics  | `npx bd stats --json`                   |
| List all tasks  | `npx bd list --parent <epic-id> --json` |
| Get ready tasks | `npx bd ready --json`                   |
| View hierarchy  | `npx bd dep tree <epic-id>`             |
| Check cycles    | `npx bd dep cycles`                     |

## Context

$ARGUMENTS
