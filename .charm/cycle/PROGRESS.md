# Progress — C1 read-across-roots

## Cursor

Current task: none — M5 (the plan's last milestone) is now all `done`; M5.3 needs
`code-review`/`charm-validate` (fresh context) like M5.1/M5.2 before the cycle can close.

## Handoff

M5.3 (help modal, S12) is done: `keys.Help` (already existed) and `keys.Cancel` open/close
it. New `cmd/noggin/help.go`: `helpBindings(pane)` (a fuller-than-footer per-pane binding
list — reuses `scrollActions` for the viewer, includes what `paneHints` deliberately omits:
PrevPane, and for the viewer Cancel/FindNext/FindPrevious), `helpBox` (renders via
`bubbles/help`'s `FullHelpView`, D1), and `overlayHelp` (true cell compositing via lipgloss
v2's `Layer`/`Compositor`/`Canvas` — undocumented in the plan-time reference, found in
vendored source; D26). `model.go` gained `helpOpen bool`, `openHelp`/`handleHelpInput`
(same exclusivity shape as `handleFindInput` — while open, only Help/Cancel act, everything
else including Quit is swallowed), and `render()` (View's new indirection: `frame()` plus
the overlay when open). 10 unit tests (`help_test.go`) + a new M5.3 pty section (11 checks).
Acceptance verified directly: opens over any pane without moving focus (all three panes
checked, unit + pty), lists exactly that pane's own bindings and not another's (unit
`helpBox` content checks + pty, both Files and Document contexts), and Cancel or Help itself
closes it, restoring the exact previous frame with focus untouched (unit byte-identical
`View().Content` + pty byte-identical screen, mirroring M5.2's round-trip proof).

A fresh-context code-review pass (subagent, both standards and spec-fidelity) on this task
before marking it done found and I fixed in the same session: a `Compositor` sizes its
output canvas to the union of every layer's bounds, not the base layer's own size, so an
unclamped modal taller/wider than the terminal (e.g. the viewer's full 18-binding list on a
short terminal) grew the rendered frame past what `View()`'s no-alt-screen drawing can show
— fixed by clamping the box with `lipgloss.MaxWidth`/`MaxHeight` before compositing
(`TestHelpOverlayNeverExceedsTheTerminal` locks it in); the modal's border color duplicated
`focusedBorderStyle`'s (now a distinct 214); `toggleHelp` was renamed `openHelp` (it never
toggled off, only `handleHelpInput` closes it); and Quit-while-open coverage was added (unit
+ implicitly by the modal's own exclusivity design). No other findings. Not yet
independently reviewed/validated (see Status table) — this was self-review before marking
done, not the fresh-context `code-reviewer`/`charm-validate` pass the Status table still
needs.

T28 (open in TRIAGE.md) still live, unchanged this session: `resolveLaunch` drops an
*absolute*-path launch argument on a `resolveCwd` failure even though it needs no cwd
(`cmd/noggin/roots.go:64-72`); no regression test covers it. Needs replan — not touched by
any M5 task.

All of M1-M5 is now `done`. Next: M5.1-M5.3 each still need `Reviewed`/`Validated` stamps
(M5.1/M5.2 were already self-noted as pending; M5.3 above) before the cycle can close per
cycle-state's "current milestone all done with any row's Validated blank" rule.

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
| M5.1 | done   | 2026-08-01 | -         | find in document (S8); find.go + keymap/model wiring; 8 unit tests + 5 pty checks; found and fixed a real bubbles/v2 SetHighlights bug along the way (see findings); fresh-context two-pass review logged T29-T36 |
| M5.2 | done   | 2026-08-01 | -         | navigator toggle (S9, D25); computeLayout gained hideNav; model.toggleNavigator/resize/layout; ctrl+n; 4 unit tests + 8 pty checks; fresh-context two-pass review logged T29-T36 |
| M5.3 | done   | 2026-08-01 | -         | help modal (S12, D26); help.go (helpBindings/helpBox/overlayHelp, lipgloss v2 Layer/Compositor); model.go helpOpen/openHelp/handleHelpInput; 10 unit tests + 11 pty checks; self fresh-context review found and fixed an unclamped-overlay overflow bug; independent fresh-context two-pass review found no fidelity issues, logged T29-T36 (mostly M5.1-scoped) |
