# Progress — C1 read-across-roots

## Cursor

Current task: M1.3

## Handoff

M1.2 is done: the frame is up. `cmd/noggin/layout.go` owns all proportions (`footerHeight`,
`navigatorPercent` 30, `rootsPercent` 35, min/max width and min pane height) and
`computeLayout(w,h)` returns the four region rects; a terminal narrower than
`minNavigatorWidth+minViewerWidth` drops the Navigator column entirely (groundwork M5.2 can
reuse). `cmd/noggin/keymap.go` is the sole place key literals may appear (D2) — bubbles
`key.Binding`s with help text (`NextPane` tab, `PrevPane` shift+tab, `Quit` q/ctrl+c), plus a
`keyMsg(name)` helper so tests never spell a key. `model.go` tracks size + `focus`, cycles
focus through `paneOrder` (roots → files → viewer, wrapping, skipping hidden panes), and
renders bordered panes (focused border 205, else 240) with a placeholder footer that names the
focused pane and the quit hint. Added deps: bubbles v1.0.0, lipgloss v1.1.0 (direct).

Verified: `make check` clean; 6 tests in `cmd/noggin/frame_test.go` pass — proportions at
80x24 and 120x40, `View()` exactly W×H with the three region titles in their bands, reflow on
resize, full forward focus cycle plus stepping back, navigator-dropped narrow case, and both
quit keys. `rg -g '*.go' -g '!cmd/noggin/keymap.go'` for key literals returns 0 matches.
End-to-end in a pty at 80x24 and 100x30: real frame drawn, focus highlight colour present,
footer switches to `Document` after tab, exit 0 with terminal restored.

Next step: M1.3 — Document Viewer. Add Glamour (v1.0.0), pre-strip frontmatter (D5, see
`findings/frontmatter-rendering.md`), wrap to the viewer pane's inner width
(`l.viewer.width-2` today, since panes are bordered), viewport scrolling by line/half-page/
page/top/bottom via bubbles `viewport`, empty-state text when no document is open, and a
fixture document for the criterion. Scroll keys go in `keymap.go`, nowhere else. The viewer
currently renders only its title placeholder in `model.renderPane`.

Gotchas still relevant:
- Pty verification now needs the OSC 11 + cursor-position startup queries answered or the
  binary hangs; `go test` sees no colour. Recipe in `findings/pty-verification.md`.
- Bubble Tea v1.3.10 (v2 still pre-release); go with the grain (D1).
- M1.3 renders a document from a bare path argument; M3.1 replaces that ad-hoc argument
  handling with the real S1 launch contract (D11).
- Sandbox Go toolchain fully working as of 2026-07-31; the Makefile's `unexport GOROOT` is
  defensive-only (`findings/go-toolchain-env.md`).

No blockers.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M2.2 | todo   | -        | -         |      |
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
