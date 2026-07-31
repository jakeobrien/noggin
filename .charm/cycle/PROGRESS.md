# Progress — C1 read-across-roots

## Cursor

Current task: M2.3 (next: M3.1)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5 (below) has not yet been through review/validate — its Status row is blank
in both columns.

## Handoff

M2.2 is done (D19): the Files Nav pane is a `filesNav` struct (`cmd/noggin/files.go`) layered
over the immutable `tree.Node`. Expansion state is a `map[string]bool` keyed by absolute
`Path` (empty = all collapsed, S5); `selected` is an index into `visibleRows()`, which
flattens the tree to just its open part (root's children at depth 0; a folder's children only
when expanded — the root node itself is not a row). `view()` keeps the same content-box
contract as the viewer (exactly width×height, selection kept on screen, reverse-video bar,
empty-state text when no root). Movement reuses `LineUp`/`LineDown`; `Expand` (right/l) and
`Collapse` (left/h) are new keymap bindings; expand/collapse clone the map (value-semantic,
per T10's convention). Focus routing in `model.Update` sends keys to the viewer or the Files
pane by focus. `main.go` grew temporary dir-argument scaffolding (dir → Files Nav, file →
viewer), flagged in-comment as replaced by M3.1/M3.4.

Evidence: `make check` clean. Unit tests (`files_test.go`, 8 tests over a hand-built
`fixtureTree`): opens top-level-only with nested file hidden, empty folder listed (and stays
listed when expanded), expand/collapse in place with selection unmoved, selection walks the
whole expanded set and clamps at both ends, expansion is value-semantic, `view()` fills the
box and marks the selection, empty-state when no root, and model-level key routing through
focus. `make check-pty` now 67 checks (was 59): a committed fixture tree
(`cmd/noggin/testdata/tree/`, empty folder held by a non-readable `.gitkeep`) launched via the
binary — top-level entries shown, nested `guide.md` hidden until docs is expanded then hidden
again on collapse, reverse-video selection bar on the wire, clean exit.

Not yet through review/validate (Status table blank). **M1.5 remains unvalidated** and M2.1
(reviewed/validated blank) too — the M1→M2 gate was crossed with the user's go-ahead; validate
M1.5/M2.1/M2.2 together at the next checkpoint.

Next: **M2.3 — activate a file.** Selecting a fixture file and pressing an activate key renders
it in the viewer with focus moving to the viewer, and scroll keys then act on it. The pieces
are in place: `filesNav.selectedNode()` returns the node under the selection (`Path` absolute,
ready for `loadDocument`); `model.openDocument(d)` already loads a document, moves focus to the
viewer and is what M2.3 should call; the viewer's scroll routing already fires on
`focus == viewerPane`. Needs a new `Activate` binding (enter is the obvious choice, currently
unbound) routed when `focus == filesPane` and the selected node is a file — a folder's enter
could toggle expand, or M2.3 can keep activate file-only and leave folder-enter for later.
Watch: `openDocument` takes a `document`, so activate must `loadDocument(selectedNode().Path)`
and surface a read error rather than exit.

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
| M1.5 | done   | 2026-07-31 | -       | two-pass review (T9, T10 logged); not yet validated |
| M2.1 | done   | -        | -         | internal/tree: policy + walk + hand-rolled gitignore (D18) |
| M2.2 | done   | -        | -         | Files Nav pane: filesNav over immutable tree (D19); 8 unit tests + 8 pty checks |
| M2.3 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M3.2 | todo   | -        | -         |      |
| M3.3 | todo   | -        | -         |      |
| M3.4 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
