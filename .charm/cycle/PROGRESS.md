# Progress — C1 read-across-roots

## Cursor

Current task: M4.1 (next: M4.2)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5, M2.1, M2.2, M2.3 have since been validated (2026-07-31, see Handoff).

## Handoff

M3.4 done (S4): the acceptance criterion turned out to already hold as a byproduct of M3.3's
`activateRoot` reusing `openFolder` (which touches only `m.files`/`m.focus`, never `m.viewer`)
and `filesNav.setRoot`'s existing reset-to-top-level-collapsed behavior — no production code
changed. The task was writing the test that proves it: `TestActivateRootLeavesOpenDocumentUntouched`
(`cmd/noggin/rootsnav_test.go`) opens the fixture document from a fresh model, scrolls it to a
non-zero offset, opens root A (`testdata/tree`) in the Files Nav with its `docs` folder expanded
and selection off the top, then activates a second root B (a temp tree with a top-level file and
a subfolder) via `m.activateRoot()`. It asserts: the Files Nav's expansion map is empty and
selection is back to 0 (collapsed at top level), B's top-level entries render but its subfolder's
child does not (still collapsed), and the viewer's offset/content/label are all byte-identical to
before — the open document is untouched.

Evidence: `make check` clean. Sanity-checked the test isn't vacuous by temporarily breaking each
half of the guarantee and confirming the new test fails, then reverting: (1) stripped `setRoot`'s
expansion/selection reset — test failed on both the expansion and selection assertions; (2) made
`openFolder` also reset the viewer — test failed on offset, content and label. Both reverts
verified with `diff` against pre-edit backups; `make check` clean again after each revert. No new
findings, no TRIAGE items — this task was pure verification of an existing guarantee, not new
behavior.

Next: **M4.1 — Edit handoff: release the terminal to `$EDITOR` (fallback `vi`), resume on exit and
re-render from disk (D3)** (needs M1.3). Acceptance: with `EDITOR` set to a script that appends a
line, the edit key returns to noggin showing the appended line, and with no document open the
footer reports that instead. Recall the standing constraint from D13/`glamour-in-bubbletea.md`:
`markdownStyle` resolves dark/light once via `lipgloss.HasDarkBackground` before `Run()`; returning
from `$EDITOR` must not re-query the terminal. Likely needs a `tea.ExecProcess`-style release/resume
(check bubbletea v2's current API for this) plus a re-render-from-disk path that reuses `loadDocument`
rather than duplicating it, and a footer message for the no-document case (`renderFooter` already
supports `m.message`, per M3.3/D22).

Still relevant:
- Go with the grain (D1) — now meaning Charm v2.
- D13's rule (no terminal query after `Run()`) stands, but v2 removed its cause. `markdownStyle`
  still resolves dark/light before `Run()` via `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)`;
  `tea.RequestBackgroundColor` is the in-program alternative but is unreliable under tmux/WSL2.
  This still constrains M4.1: returning from `$EDITOR` must not re-query.
- Dependency graph got smaller, not bigger: termenv, muesli/{ansi,reflow,termenv}, go-osc52,
  go-isatty, go-localereader, coninput and x/cellbuf all dropped; ultraviolet (pre-1.0
  pseudo-version), x/termios and x/windows came in. 31 indirect deps down to 24. Noted for
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
| M3.1 | done   | -        | -         | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check |
| M3.2 | done   | -        | -         | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | -        | -         | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | -        | -         | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
