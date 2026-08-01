# Progress — C1 read-across-roots

## Cursor

Current task: M5.3

## Handoff

M5.2 (navigator toggle, S9) is done, not yet reviewed/validated: `keys.Navigator` (ctrl+n,
D25) in keymap.go; `model.toggleNavigator`/`resize`/`layout` in model.go; `computeLayout`
grew a `hideNav bool` third parameter (layout.go) so a hidden sidebar and a too-narrow
terminal share the same width computation; `footer.go`'s `paneHints` advertises it on every
pane alongside Edit. 4 unit tests (`navigator_test.go`) plus a new M5.2 pty section (8
checks). Acceptance verified directly: the key hides the whole sidebar and the viewer's
border now spans the full terminal width (unit layout-width assertions + pty border-position
check), the open document actually re-wraps to fewer lines at the wider width (unit
`lineCount` comparison + pty column-shift check), a second press restores the exact previous
layout (unit struct equality + pty byte-identical screen), and the Files Nav's own selection
survives a hide/restore cycle (unit `selected` index check + pty: after restoring, tabbing
focus back to Files and activating still opens the file the moved selection pointed at, not
the reset-to-top default). Toggling moves focus off a Navigator pane it can no longer land in
(the same `resolveFocus` rule a narrow-terminal resize already follows) but does not restore
focus on the way back — S9 only promises the sidebar's own selection survives, not focus, and
no test depends on the latter.

T28 (open in TRIAGE.md) still live, unchanged this session: `resolveLaunch` drops an
*absolute*-path launch argument on a `resolveCwd` failure even though it needs no cwd
(`cmd/noggin/roots.go:64-72`); no regression test covers it. Needs replan before or alongside
remaining M5 work — not touched by M5.1 or M5.2.

M5.3 (help modal) is next — the last milestone and the cut candidate (D8): M5.1 and M5.2 are
both done and no longer cuttable; if the cycle overruns from here the only lever left is
cutting M5.3 itself.

Still relevant:
- Go with the grain (D1) now means Charm v2.
- D13's rule (no terminal query after `Run()`) stands; v2 removed its original cause, but
  `markdownStyle` still resolves dark/light via `lipgloss.HasDarkBackground` before `Run()`.
  M4.1's edit handoff already respects this; M5 work must too.
- Dependency graph shrank on the v2 migration (31 → 24 indirect deps) — noted for
  security-review time.
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
| M5.1 | done   | -        | -         | find in document (S8); find.go + keymap/model wiring; 8 unit tests + 5 pty checks; found and fixed a real bubbles/v2 SetHighlights bug along the way (see findings) |
| M5.2 | done   | -        | -         | navigator toggle (S9, D25); computeLayout gained hideNav; model.toggleNavigator/resize/layout; ctrl+n; 4 unit tests + 8 pty checks |
| M5.3 | todo   | -        | -         |      |
