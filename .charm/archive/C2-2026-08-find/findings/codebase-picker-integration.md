# Where a Picker plugs into the C1 codebase

Written at C2 plan time (2026-08-03) by a research subagent reading the tree directly. Facts
and citations only — design choices live in PLAN.md / DECISIONS.md.

## Exclusive-input modals are hand-duplicated `if` gates in `Update`

- `model` holds `focus pane` (`rootsPane|filesPane|viewerPane`, `model.go:19-25`) plus two
  modal flags *outside* the pane concept: `viewer.findState.editing` (`find.go:19-25`) and
  `m.helpOpen` (`model.go:74`). Both documented as never touching `m.focus`
  (`model.go:70-77`).
- `Update` on `tea.KeyPressMsg` (`model.go:262-283`): clear `m.message` → `if
  m.viewer.findEditing() { handleFindInput }` → `if m.helpOpen { handleHelpInput }` → global
  bindings switch → `switch m.focus` for pane-local keys. Order is the whole mechanism; there
  is no modal stack or interface. A Picker is a third hand-copy of this shape.
- `handleFindInput` (`model.go:150-164`) matches only `Cancel` / `Activate`; everything else
  falls through to `updateFindQuery(msg)` which types the raw keystroke — arbitrary text is
  deliberately never routed through `keymap.go`. This is the "always typing" precedent.
- `handleHelpInput` (`model.go:236-241`): only `Help`/`Cancel` do anything; all else swallowed,
  including Quit.

## Compositing and layout

- `model.render()` (`model.go:296-302`) is the single composition point: build `frame()`, then
  if `helpOpen` return `overlayHelp(frame, w, h, focus)`. `View()` just wraps it
  (`model.go:289-291`).
- `overlayHelp` (`help.go:78-91`): clamp the box with `MaxWidth/MaxHeight` first (a
  `Compositor` sizes its canvas to the *union* of layer bounds, so an oversized modal would
  overflow the no-alt-screen draw — `help.go:82-87`), then
  `lipgloss.NewCompositor(NewLayer(frame).Z(0), NewLayer(box).Z(1)).Render()`. True cell
  compositing, so the frame stays visible around the edges (D26).
- `layout.go` (`computeLayout`/`rect.content()`, `layout.go:34-59,68-91`) only splits the
  terminal into Roots/Files/Viewer/Footer; it knows nothing about modals. No precedent exists
  for a multi-pane layout *inside* a modal — Help is a single box.

## The Find surface, and the grep hand-off gap

- `find{editing bool; input textinput.Model; matches []findMatch; current int}`
  (`find.go:19-38`); `findMatch{line, colStart, colEnd}` are positions in the **viewport's
  rendered line space**.
- `commitFind()` (`find.go:87-91`) clears `editing` but keeps matches/current, which is what
  makes post-Enter next/previous work. `cancelFind()` (`find.go:93-98`) resets fully.
- `findMatches` (`find.go:157-224`) substring-searches `ansi.Strip(vp.GetContent())` with
  smart case (`foldLower`/`origSpan`, `find.go:226-259`). `applyCurrentMatch`
  (`find.go:120-136`) is the reusable scroll+highlight mechanic (`vp.EnsureVisible` plus
  `vp.StyleLineFunc`). `nearestMatch` (`find.go:262-273`) picks the first match at/after a
  line.
- Viewer surface: `startFind`, `updateFindQuery`, `commitFind`, `cancelFind`, `findEditing`,
  `hasFindMatches`, `findNext`, `findPrevious`, `findView`.
- **Gap:** nothing takes `(query, line)` and produces a committed find. Worse, `findMatches`
  works on Glamour-rendered, word-wrapped content, while ripgrep reports raw-file line
  numbers — frontmatter stripping, re-flowing and markdown syntax removal mean the two line
  spaces do not correspond. See D34.
- `viewer.render()` resets `findState = newFind()` (`viewer.go:70-73`), so any hand-off must
  run *after* `open(d)`.

## Files, roots, documents

- Allowlist: `internal/tree/policy.go:19-38` — exts `.md/.markdown/.txt`, names
  `README/LICENSE/CHANGELOG`, case-insensitive both ways (D23).
- Hand-rolled `.gitignore` subset in `internal/tree/gitignore.go` (last-match-wins,
  `gitignore.go:150-166`); `.git` excluded independently (`tree.go:53-55`). rg's rules are
  allowed to disagree (D28).
- `root{path, starred}` and the pure `resolution` launch contract in `roots.go`;
  `model.applyResolution` (`model.go:97-118`) does the actual walking/loading.
- Opening a document is always `loadDocument(path)` (`document.go:15-31`) →
  `model.openDocument(d)` (`model.go:120-124`) → `viewer.open` → `render` → `GotoTop`
  (`viewer.go:47-52`). No "open at line N" loader exists; composition is the caller's job.

## Footer, keymap, help registration

- `keymap.go` is the only production file allowed key literals (D2); bindings carry
  `WithHelp`, so footer/help are generated. `footer.go:11-26`: `paneHints(p)` is a curated,
  truncation-ordered subset plus `alwaysHints = {Help, Quit}`. `help.go:16-38`
  `helpBindings(p)` is the fuller list. Modal states bypass `hintText`: `renderFooter`
  special-cases `findEditing()` to render the textinput instead (`model.go:378-397`).

## Tests, and the async gap

- `make check` = build, vet, fmt, test (`Makefile:19`). `make check-pty` (`Makefile:24-27`)
  builds and runs `python3 test/pty/checks.py` against the real binary — a linear
  `check(name, ok, detail)` script per milestone, driven by `test/pty/harness.py`'s pty
  `Session` + hand-rolled cell-grid `Screen` (`harness.py:78-176`), needed because bubbletea
  v2 emits diff-based escapes.
- Go test helpers `sized`, `press`, `typeText`, `concrete`, `openedFixture` live in
  `frame_test.go:186-222`.
- **No async precedent:** zero matches for `tea.Tick` / debounce / goroutine / `time.After`
  across the repo. All `tea.Cmd` use is `nil`, `tea.Quit`, textinput's own cmd, or the single
  `tea.ExecProcess` `$EDITOR` handoff (`model.go:206-227`, `finishEdit` `model.go:244-256`) —
  which is *blocking*, structurally unlike streaming subprocess output.
- **No subprocess use at all besides `$EDITOR`** — no `os/exec` elsewhere.
