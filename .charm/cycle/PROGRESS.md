# Progress — <cycle name>

<!-- Per-cycle. Cursor, checkpoints and statuses ONLY — no work summaries (git history
     carries those). The Handoff section is the single exception: overwrite it each session,
     never append. Statuses: todo | in-progress | done | blocked | deferred.
     Blocked rows reference the DECISIONS.md / TRIAGE.md item they wait on — the set of
     blocked rows is the cycle's blocking-issue list. -->

## Cursor

Current task: <key>

## Checkpoints

<!-- Overwrite in place. Set by validate / code review — or by the parent that dispatched
     them as subagents. cycle-state uses these to detect when the milestone gate is due. -->

- Last validated: <milestone key, yyyy-mm-dd> | none
- Last reviewed: <task key or commit, yyyy-mm-dd> | none

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

<current-state snapshot>

## Status

| Key  | Status      | Note                              |
|------|-------------|-----------------------------------|
| M1.1 | done        |                                   |
| M1.2 | in-progress |                                   |
| M1.3 | blocked     | <one-line reason, or D<n> ref>    |
