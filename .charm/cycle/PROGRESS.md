# Progress — C1 read-across-roots

## Cursor

Current task: M2.1

## Handoff

M1.3 is done: the Document Viewer renders. Two new files. `cmd/noggin/document.go` holds the
document type (`path` + `src`; `src` is kept so re-wrapping never touches the disk),
`loadDocument`, `stripFrontmatter`, `expandTabs` and `renderMarkdown`. `cmd/noggin/viewer.go`
holds the `viewer` (a bubbles `viewport` plus the document, the width it was last rendered at,
and a render error), the empty-state text, and `scrollActions` — the binding→viewport-move
table that `Update` walks and S12's help modal can reuse.

Things worth knowing before touching this:
- **Glamour's `WithAutoStyle` hangs the program.** It re-queries the terminal background on
  every render and Bubble Tea eats the reply. The style is now resolved once in `main` before
  the program starts (D13, `findings/glamour-in-bubbletea.md`). Never let a terminal query
  happen after `Run()` — this directly constrains M4.1's editor handoff.
- **Pane geometry moved into `layout.go`**: `paneBorderWidth`, `paneTitleHeight` and
  `rect.content()`. A pane is border + title row + content; `renderPane(p, r, content)` now
  takes content sized to `r.content()`. M2.2's file tree should use the same helper rather than
  subtracting 2s by hand.
- Glamour leaks literal tabs out of code fences and lipgloss measures them as one cell, so
  tab-indented code drew through the pane border. `expandTabs` (4-column stops, CommonMark's
  own figure) fixes it; the test asserts zero tabs in rendered output.
- `stripFrontmatter` requires the opening `---` at offset 0, unlike Glow, which takes the first
  two fences found anywhere and so eats the head of a document whose thematic breaks are `---`.
  There is a regression test for that case.
- Scroll keys only act when `focus == viewerPane`, so `j`/`k`/arrows are free for M2.2.
- The viewer pane title becomes `Document: <basename>` when a document is open; the empty pane
  is still just `Document`, which is what `frame_test.go` asserts.
- Launching with a path focuses the viewer (`model.openDocument`) — reading is the intent. A
  bad path still exits 1; M3.1 replaces that with the S1 contract where every rejection leaves
  a running session (D11).

Verified: `make check` clean, 40 test cases pass (`frame_test.go` unchanged + 8 new tests in
`viewer_test.go`): frontmatter stripping (8 cases incl. CRLF, unclosed, mid-document `---`),
tab expansion (6 cases), the fixture rendered with no frontmatter and every construct present
and no line over the wrap width, `View()` exactly 80x24 naming the open file, empty state with
no document, all five scroll moves via all 8 bindings from set offsets, scroll keys inert
outside the viewer, and re-wrap on resize (line count grows 120→80 columns). Key-literal grep
outside `keymap.go`: 0 matches. Nineteen end-to-end pty checks at 80x24 and 100x30 all pass:
real styled frame (colour escapes, box-drawing tables, focused border 205), frontmatter
absent, `G` reaches the document's last line, `ctrl+d` reaches the code fence with no row
exceeding the terminal width and no tabs on screen, empty state on bare launch, exit 0 with the
terminal restored.

Next step: M2.1 — readable-file policy and tree walk. Nothing exists for it yet: no filesystem
package, no `.gitignore` handling, no allowlist. The criterion wants a unit test over a fixture
tree (gitignored `.md` excluded, `.git/` absent, `.charm/notes.md` present, `main.go` absent,
`README` present) and the allowlist in one declaration. `cmd/noggin/testdata/` already exists
for the fixture document, so a fixture tree can live beside it. This is pure Go with no TUI
surface, so it needs no pty work. Decide early whether to take a `.gitignore` dependency
(`go-gitignore`, or shell out) or hand-roll — worth a findings note either way.

Still relevant:
- Bubble Tea v1.3.10 (v2 still pre-release); go with the grain (D1).
- Glamour v1.0.0 dragged the direct `lipgloss` requirement up to the pseudo-version
  `v1.1.1-0.20250404203927-76690c660834` and added chroma/goldmark/bluemonday/`x/net`. Noted
  for security-review time, not a problem now.
- Sandbox Go toolchain fully working as of 2026-07-31; the Makefile's `unexport GOROOT` is
  defensive-only (`findings/go-toolchain-env.md`).

No blockers.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | done   | -        | -         |      |
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
