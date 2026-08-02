# Progress — C1 read-across-roots

## Cursor

Current task: none — replan (2026-08-02) triaged the M5-review TRIAGE backlog into M5.4 and
M5.5. Next up: M5.4, then M5.5, then the still-outstanding M5.1-M5.3 `Validated` stamps,
before the cycle can close.

## Handoff

Replan (2026-08-02) triaged all 9 open TRIAGE items (T28, carried over from the M3.5 review;
T29-T36, from the M5.1-M5.3 fresh-context two-pass review):

- Accepted as **M5.4** (correctness/UX, effort 2): T28 (`resolveLaunch` drops an
  absolute-path launch argument on a `resolveCwd` failure it doesn't need), T29
  (`findMatches`' smart-case path can misattribute a match's line when a rune's case fold
  changes its byte length), T34 (a resize mid-find-edit silently drops the query with no
  footer message).
- Accepted as **M5.5** (standards cleanup, effort 1): T30 (13 literal `\u2014` escapes
  instead of real em-dashes), T32 (stale future-tense help-modal comments in keymap.go now
  that M5.3 shipped), T33 (`viewer.find` field shadows its own type's name), T35
  (undocumented intent: FindNext/Previous dispatch globally like Find).
- Deferred to `ROADMAP.md` Unscheduled: T31 (no shared exclusive-input-mode abstraction
  across modals) — real observation, but only bites if a third modal is ever added, and none
  is planned in C1.
- Dropped: T36 (M5.1's new indirect dependency) — already covered by this doc's existing
  dependency-graph note for security-review time; folded the specific detail in below rather
  than opening a separate item.

Logged as D27. `TRIAGE.md` is now empty. M5.4/M5.5 touch files (`roots.go`, `find.go`,
`viewer.go`, `model.go`, `keymap.go`, `navigator_test.go`) already `Reviewed` under
M3.1/M3.5/M4.2/M5.1/M5.2/M5.3 — once built, check whether those tasks' `Reviewed` cells
still hold or need clearing back to `-` per the doc-map rule (code changed since the stamp).
M5.1-M5.3 still need their fresh-context `Validated` stamps before the cycle can close,
independent of M5.4/M5.5.

Still relevant:
- Go with the grain (D1) now means Charm v2.
- D13's rule (no terminal query after `Run()`) stands; v2 removed its original cause, but
  `markdownStyle` still resolves dark/light via `lipgloss.HasDarkBackground` before `Run()`.
  M4.1's edit handoff already respects this; M5 work must too.
- Dependency graph shrank on the v2 migration (31 → 24 indirect deps); M5.1 added
  `charmbracelet/x/ansi` as a direct dependency and pulled in `atotto/clipboard` indirectly
  via `bubbles/v2/textinput` (unused) — noted for security-review time (T36, dropped as its
  own item, folded in here).
- Sandbox Go toolchain fully working; the Makefile's `unexport GOROOT` is defensive-only.

No blockers.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.2 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.3 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.4 | done   | 2026-08-01 | 2026-07-31 | M1 complete and validated |
| M1.5 | done   | 2026-07-31 | 2026-07-31 | two-pass review (T9, T10 logged); validated (fresh-context subagent) |
| M2.1 | done   | 2026-07-31 | 2026-07-31 | internal/tree: policy + walk + hand-rolled gitignore (D18); review found T11-T14; validated |
| M2.2 | done   | 2026-07-31 | 2026-07-31 | Files Nav pane: filesNav over immutable tree (D19); 8 unit tests + 8 pty checks; validated |
| M2.3 | done   | 2026-07-31 | 2026-07-31 | Activate binding (enter): filesNav → loadDocument → openDocument; 3 unit tests + 5 pty checks; review found T15; validated |
| M2.4 | done   | 2026-08-01 | 2026-08-01 | internal/tree hardening (T11-T14, D23): dead branch dropped, doc comments aligned, name matching case-insensitive, escape test added |
| M3.1 | done   | 2026-07-31 | 2026-07-31 | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check |
| M3.2 | done   | 2026-07-31 | 2026-07-31 | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | 2026-07-31 | 2026-07-31 | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | 2026-07-31 | 2026-07-31 | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M3.5 | done   | 2026-08-01 | 2026-08-01 | replan C1/2026-08-01: roots/messaging correctness fixes (T17, T18, T20); review found T28 (confirmed still open by validate) |
| M3.6 | done   | 2026-08-01 | 2026-08-01 | replan C1/2026-08-01: M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27) |
| M4.1 | done   | 2026-08-01 | 2026-08-01 | edit handoff (S10/D3): tea.ExecProcess release/resume + finishEdit re-render from disk |
| M4.2 | done   | 2026-08-01 | 2026-08-01 | footer key hints (S11); T15 (activate read-error message) and T26 (empty-Roots-Nav guidance) folded in |
| M5.1 | done   | 2026-08-01 | -         | find in document (S8); find.go + keymap/model wiring; 8 unit tests + 5 pty checks; found and fixed a real bubbles/v2 SetHighlights bug along the way (see findings); fresh-context two-pass review logged T29-T36 |
| M5.2 | done   | 2026-08-01 | -         | navigator toggle (S9, D25); computeLayout gained hideNav; model.toggleNavigator/resize/layout; ctrl+n; 4 unit tests + 8 pty checks; fresh-context two-pass review logged T29-T36 |
| M5.3 | done   | 2026-08-01 | -         | help modal (S12, D26); help.go (helpBindings/helpBox/overlayHelp, lipgloss v2 Layer/Compositor); model.go helpOpen/openHelp/handleHelpInput; 10 unit tests + 11 pty checks; self fresh-context review found and fixed an unclamped-overlay overflow bug; independent fresh-context two-pass review found no fidelity issues, logged T29-T36 (mostly M5.1-scoped) |
| M5.4 | planned | -      | -         | replan C1/2026-08-02 (D27): M5-review correctness/UX fixes (T28, T29, T34) |
| M5.5 | planned | -      | -         | replan C1/2026-08-02 (D27): M5-review standards cleanup (T30, T32, T33, T35) |
