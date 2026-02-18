---
description: Review code changes and create beads issues from findings. Use when reviewing feature branch changes and wanting to track findings as tasks.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

1. Run `.specify/scripts/bash/check-prerequisites.sh --json` from repo root and parse FEATURE_DIR. All paths must be absolute. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

2. **Retrieve Beads Epic ID**:

   a. Read the epic ID from spec.md front matter:

   ```bash
   grep "Beads Epic" FEATURE_DIR/spec.md | grep -oE 'workspace-[a-z0-9]+|bd-[a-z0-9]+'
   ```

   b. If not found, the `/code-review` skill will auto-detect it from the git branch

   c. If code-review reports "No epic found", prompt user:
   - Ask if they want to create an epic first using `/sp:01-specify`
   - Or note that review was completed without beads task creation

3. **Run Code Review Skill**:

   **IMPORTANT**: The `/code-review` skill now handles everything automatically:
   - Detects epic from git branch (searches open and closed epics)
   - Reopens epic if it's closed and findings exist
   - Invokes three review subagents in parallel:
     - quality-review (correctness, test quality, simplicity, code standards)
     - security-review (OWASP vulnerabilities, auth, data security)
     - clean-architecture-validator (layer boundaries, dependency violations)
   - Aggregates and deduplicates findings
   - Creates beads tasks automatically with proper priority mapping
   - Generates consolidated report with copy-paste prompt

   a. Invoke the skill with appropriate scope:

   ```
   /code-review --scope branch
   ```

   - Default scope: `branch` (changes vs main branch)
   - If user specified `--scope`, use that scope
   - If user specified `--base`, use that base ref

   b. The skill output includes:
   - Consolidated review report for non-technical managers
   - All findings organized by severity
   - List of beads tasks created (with IDs)
   - Copy-paste prompt to fix all issues

4. **Extract Task Creation Summary**:

   From the code-review output, note:
   - Number of tasks created
   - Epic ID used
   - Whether epic was reopened
   - Any errors during task creation

5. **Close Phase Task and Optionally Epic**:

   After the code-review skill completes, close the phase task.

   a. Find the review phase task:

   ```bash
   npx bd list --parent <epic-id> --json | jq -r '.[] | select(.title | contains("[sp:09-review]")) | .id'
   ```

   b. Close the review task with a completion summary:

   ```bash
   npx bd close <review-task-id>
   ```

   c. Add a comment summarizing the review:

   ```bash
   npx bd comment <review-task-id> "Review complete: <N> findings, <M> tasks created"
   ```

   d. Check if all phase tasks are now closed:

   ```bash
   OPEN_PHASES=$(npx bd list --parent <epic-id> --status open --json | jq '[.[] | select(.title | contains("[sp:"))] | length')
   ```

   e. If all phase tasks are closed:
   - Check if any Critical/High priority tasks were created
   - If NO Critical/High tasks:
     - Ask user: "All phases complete and no critical issues. Close the epic? (yes/no)"
     - If user confirms: `npx bd close <epic-id>`
   - If Critical/High tasks exist:
     - Report: "Phase [sp:09-review] complete. Critical/High priority tasks must be addressed before closing epic."
   - If user declines (e.g., to address findings first), leave epic open

   f. Report: "Phase [sp:09-review] complete. Feature workflow finished."

## Error Handling

- **No epic found**: The `/code-review` skill will report this and generate review without task creation
- **No findings**: The skill will report "No issues found" and exit successfully
- **Beads error**: The skill handles errors and includes them in its report
- **Epic closed**: The skill automatically reopens closed epics when findings exist
- **Subagent failure**: The skill continues with successful reviews and notes failures

## Changes from Previous Version

This command has been simplified to leverage the updated `/code-review` skill which now:

- Automatically detects and reopens closed epics
- Runs three specialized reviews in parallel
- Creates beads tasks with deduplication built-in
- Generates comprehensive reports for non-technical managers
- Provides copy-paste prompts for fixing issues

No manual parsing or task creation is needed in this command anymore.

## Beads Commands Reference

| Action           | Command                                        |
| ---------------- | ---------------------------------------------- |
| View tasks       | `npx bd list --parent <epic-id> --status open` |
| View ready tasks | `npx bd ready --parent <epic-id>`              |
| Close phase task | `npx bd close <review-task-id>`                |
| Comment on task  | `npx bd comment <task-id> "message"`           |
| Close epic       | `npx bd close <epic-id>`                       |

**Note**: Task creation is now handled automatically by the `/code-review` skill.

## Example Usage

```bash
# Review branch changes and create issues
/sp:review

# Review with specific scope
/sp:review --scope staged

# Review against specific base
/sp:review --base main
```
