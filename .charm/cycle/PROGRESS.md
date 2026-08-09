# Progress — C3 compare

<!-- Per-cycle. Cursor and statuses ONLY — no work summaries (git history carries those).
     The Handoff section is the single exception: overwrite it each session, never append.
     Statuses: todo | in-progress | done | blocked | deferred.
     Blocked rows reference the DECISIONS.md / TRIAGE.md item they wait on — the set of
     blocked rows is the cycle's blocking-issue list.

     Reviewed / Validated columns: date (yyyy-mm-dd) the task was last covered by code
     review / validate, or `-` if it never has been. Set per task key, by the role that
     performed the check — or the parent that dispatched it as a subagent — for exactly the
     tasks that check actually covered, never assumed for a whole milestone. This is what
     makes reviewed/validated state accurate across replans that insert tasks earlier,
     partial mid-milestone checks, and skipped end-of-milestone gates — a single cycle-wide
     checkpoint can't represent any of those. Clear a cell back to `-` when the task's code
     changes after that date (fix-ups, reopened work) — it needs re-covering. -->

## Cursor

Current task: M5.1

## Handoff

M4.1 is done (this session): confirmed structurally that DV2's existence, Active Document
and Peek/Find state are never persisted or restored — `store.go`'s roots file (written only
by `saveStore`/`toggleStore`) is the *only* on-disk write path anywhere in the package
(grepped every `os.WriteFile` call in `cmd/noggin/*.go`, non-test, to confirm), so there was
no persistence mechanism for viewer state to remove; the task was verification, not a code
change. Added `TestDV2NeverPersistedAcrossRestart` in the new `dv2_persistence_test.go`: a
fixture (`launchFixture`) reproduces main.go's real launch sequence
(`loadStore`-as-`loadStarredRoots` → `newModel().withStorePath(...).applyResolution(resolveLaunch(...))`)
against a real on-disk store file. First "process" stars a root via the real `toggleStar`
action, opens DV1's Active Document, then creates DV2 with its own Active Document plus
committed Find state; asserts the store file is byte-identical before and after DV2's whole
lifetime; then relaunches a brand-new model against the same store and confirms `dv2 == nil`,
DV1 carries no Find/Peek state, and the starred root (genuinely persisted state, gathered
both before and after DV2 existed) survives the restart unaffected. Verified the test bites
by regression injection (temporarily made `newModel()` construct a live `dv2` unconditionally
— test failed with "DV2 exists in a freshly relaunched model — it was persisted"; reverted
clean). `make check` and `make check-pty` both pass (no new keys, so no pty additions
needed).

Next: **M5.1** — relocate the staleness indicator from the Footer's appended hint suffix to
its own dedicated row above the Footer (effort 3).

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.2 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.3 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.3a| done   | 2026-08-09 | 2026-08-09 | rebind d/shift+d→o/ctrl+o, D55 |
| M1.4 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.4a| done   | 2026-08-09 | 2026-08-09 | rebind shift+enter+L/shift+p→ctrl+enter+ctrl+l/ctrl+p, D55 |
| M1.5 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.6 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.6a| done   | -        | -         | TRIAGE T5, D59 |
| M1.7 | done   | -        | -         | TRIAGE T6 |
| M2.1 | done   | -        | -         |      |
| M3.1 | done   | -        | -         |      |
| M4.1 | done   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
