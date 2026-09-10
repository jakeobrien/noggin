# Progress — C6 sources

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

Current task: M8.5 done — cycle's milestones (M1-M10) are all `done`

## Handoff

**Build 2026-09-09 (M8.5):** disambiguated every remaining bare `S12` citation across
`cmd/noggin` (picker.go, keymap.go, model.go, help.go, footer.go, files_search.go and their
test files — 21 sites) to `S12/C1` (this cycle's C1-era "Help modal" S12, e.g. helpOpen,
openHelp, helpBindings, the footer's discoverability-backstop comments, the Help tests) or
`S12/C2` (C2-era "Query threshold, result cap, empty states" S12, e.g.
pickerQueryMinLength's comment, the truncation-flag comments in picker.go/files_search.go/
files_search_test.go, pickerModeScope, the Picker status-line constants and their test).
Classification cross-checked against the archived `C1-2026-08-read-across-roots/SPEC.md`
(S12 = Help) and `C2-2026-08-find/SPEC.md` (S12 = Query threshold/result cap/empty states,
delivered by that cycle's M2.5, whose done-condition explicitly bundles the mode/Scope
indicator in with the five status states — resolving the one non-obvious case,
picker_view.go's `pickerModeScope`). `grep -rn "S12" cmd/noggin | grep -v "S12/C"` and the
`rg` equivalent (grep misreports model.go and the `noggin` binary as binary files; `rg`
confirms both are false positives) are empty — M8.5's own AC. `make check` green, `gofmt -l
cmd/noggin` clean.

Next step: M8.5 was the cycle's last open task — all milestones M1-M10 are `done`. TRIAGE.md
is empty (fifth replan pass drained it). This is an all-milestones-done-and-validated-pending
state: M8.5 itself has no Reviewed/Validated stamp yet, so the next session should run
charm-validate (fresh context) — and code-review, since M8.5 is a doc-comment-only change a
reviewer can assess quickly — covering M8.5 (and confirming M8.4/M8.5 together fully resolve
the bare-S12 ambiguity), then `charm-security-review` and `/charm-close` per the all-done
cycle-state path. No blockers.

**Replan 2026-09-09 (fifth pass):** triaged TRIAGE's sole remaining item, T32 (low severity,
non-blocking — M8.4 fixed only the 5 bare `(S12)` citation sites its own AC named; a fuller
sweep found the same now-ambiguous bare `S12` recurring at ~20 more sites across footer.go,
keymap.go, files_search.go, picker.go, picker_view.go and their test files). Accepted rather
than deferred to ROADMAP (a one-line "more citation cleanup" entry wouldn't survive the next
prune) or wontfixed (no real judgment call — it's leftover scope from M8.4's own deliberately
narrowed AC, not a declined idea): added **M8.5**, same disambiguation convention as M8.4
(`S12/C1` / `S12/C2`), effort 2. TRIAGE.md is now empty. No DECISIONS.md entry — severity is
low and the change isn't material.

Next step: build M8.5, then re-run charm-validate (fresh context) covering M8.5 and the whole
cycle before close-out — TRIAGE is empty, so nothing else gates it.

**Build 2026-09-09 (M10.1):** rewrote README.md end-to-end for S1-S13's user-visible
surface: Sources Nav (Wikipedia row, Root renaming), Files Nav Browse Modes/Scopes with
Wikipedia active, the three-mode Picker (File/Grep/Wikipedia, 2-char threshold, per-mode
debounce, Scope skipping Active Root/Source while Wikipedia is active), opening/caching/
refreshing/refusing-edit on a Wikipedia article, the S9/S9a Document Viewer location line
and its CC BY-SA attribution text, background focus tint, half-page pgup/pgdn, starring and
Recent for articles, and the roots/documents on-disk store formats (Name field, article
record shape). Every key literal, status string, and footer/refusal message quoted in the
README was cross-checked against source (keymap.go, picker.go, picker_view.go, rootsnav.go,
browse.go, location.go, article_fetch.go, store.go, docstore.go) — `grep -n "Roots Nav\|No
roots\|at least 3" README.md` is empty (M10.1's own AC). Independently re-verified by a
validator subagent (fresh context): PASS on every spot-check, all internal anchor links
resolve, `make check` green.

Next step: M10.1 was the plan's last task — the cycle's milestones (M1-M10) are all `done`.
One open TRIAGE item remains (T32, low severity, non-blocking, a citation-clarity sweep) —
not a gate. This is a milestone-complete-with-validation-due state per cycle-state rules:
the next session should run charm-validate (fresh context) before closing the cycle, then
handle T32 via replan (fold into a final pass or defer to ROADMAP) as part of close-out.

**Build 2026-09-09 (M9.2):** fixed TRIAGE T29's real bug — `tintPaneBackground`'s SGR-parsing
scan (style.go) tested every `;`-split parameter token for bare equality against `"7"`/`"27"`,
so a 256-colour index or truecolor component equal to 7 or 27 following a `38`/`48`
(extended foreground/background) token was misread as reverse-video set/clear, silently
skipping S11's tint. Fixed by scanning positionally: on a `38`/`48` token, skip the
parameter(s) that belong to it (one for `;5;n`, three for `;2;r;g;b`) before resuming the
scan, so a colour value is never tested against the plain SGR codes at all — confirmed the
existing `38;5;245;7` reverse-video fixture (focus_background_test.go) still correctly treats
its trailing bare `7` as real reverse video (it comes after the color spec's own tokens are
skipped, not inside them).

Added `TestTintPaneBackgroundNotFooledByExtendedColorParams` (4 subtests: 256-colour and
truecolor, foreground and background, index/component 7 and 27) — confirmed 2 of the 4 fail
on pre-fix code (48;5;7 and 48;2;7;100;200 wrongly read as reverse, so untinted) and all 4
pass post-fix. `make check` and `make check-pty` (268 checks) both green.

Next step: cursor moved to M10.1, the cycle's last task (README update covering S1–S13's
user-visible changes). TRIAGE has one open item (T32, low severity, non-blocking — the wider
bare-S12 sweep M8.4 didn't cover) for the next replan pass; not a gate for M10.1. No blockers.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T2) |
| M1.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T3) |
| M2.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T4, T5) |
| M2.5  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T6, T7, T8) |
| M2.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T10) |
| M2.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T11, T12) |
| M1.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T17) |
| M2.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T15) |
| M3.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T18) |
| M1.10 | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T2) |
| M2.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T22) |
| M2.10 | done   | 2026-09-09 | 2026-09-09 | added by replan 2026-09-08 (TRIAGE T2) |
| M2.11 | done   | -        | -         | added by replan 2026-09-09 (TRIAGE T31) |
| M3.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T23) |
| M1.11 | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T3); checks.py:164 fix closes TRIAGE T30 |
| M5.3  | done   | 2026-09-09 | 2026-09-09 | added by replan 2026-09-08 (TRIAGE T4) |
| M6.1  | done   | 2026-09-09 | 2026-09-08 | pre-agreed cut 1 |
| M7.1  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.2  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.3  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.4  | done   | -        | -         | added by replan 2026-09-09 (TRIAGE T24, T25) |
| M8.1  | done   | 2026-09-09 | 2026-09-09 |      |
| M8.2  | done   | 2026-09-09 | 2026-09-09 | scope widened by replan 2026-09-08: S9a attribution line |
| M8.2a | done   | -        | -         | added by replan 2026-09-09 (TRIAGE T26, T28) |
| M8.3  | done   | 2026-09-09 | 2026-09-09 |      |
| M8.4  | done   | -        | -         | added by replan 2026-09-09 (TRIAGE T27) |
| M9.1  | done   | 2026-09-09 | 2026-09-09 | pre-agreed cut 3 |
| M9.2  | done   | -        | -         | added by replan 2026-09-09 (TRIAGE T29) |
| M10.1 | done   | 2026-09-09 | 2026-09-09 |      |
| M8.5  | done   | -        | -         | added by replan 2026-09-09 (TRIAGE T32) |
