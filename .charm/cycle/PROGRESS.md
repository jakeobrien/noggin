# Progress — C1 read-across-roots

## Cursor

Current task: M2.1 (next: M2.2)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5 (below) has not yet been through review/validate — its Status row is blank
in both columns.

## Handoff

M1.5 is done (T3-T5, D17): the M1.1-M1.4 checkpoint review's cleanup items are closed.
`keyPress`/`modifiers`/`namedKeys` moved out of `keymap.go` into a new `keymap_test.go` — `nm` on
the built binary confirms neither symbol is linked in, and `keymap.go` itself compiles with no
test-only imports left. `viewer_test.go` no longer touches `bubbles/viewport.Model` methods
directly; `viewer` (`viewer.go`) now exposes `offset()`, `setOffset()`, `atBottom()`,
`lineCount()`, `height()` and every test call site was switched to them (`rg '\.vp\.'
*_test.go` — no matches). `pane.title()` (`model.go`) now carries a comment explaining it is a
terse display label, not the `GLOSSARY.md` term ("Roots Nav"/"Files Nav"/"Document Viewer").
Evidence: `make check` clean (build/vet/fmt/test, same 43 cases); `make check-pty` still 59/59 —
neither change touches runtime behaviour. Not yet through review/validate (Status table blank);
no functional risk expected since this was a pure refactor, but flag for the next checkpoint.

Next: **M2.1 — readable-file policy and tree walk.** Nothing exists for it: no filesystem
package, no `.gitignore` handling, no allowlist. The criterion wants a unit test over a fixture
tree (gitignored `.md` excluded, `.git/` absent, `.charm/notes.md` present, `main.go` absent,
`README` present) and the allowlist in one declaration. `cmd/noggin/testdata/` already holds the
fixture document, so a fixture tree can live beside it — but note a gitignored file inside the
repo's own tree is awkward to commit, so decide early how the fixture expresses "gitignored"
(a `.gitignore` in the fixture dir plus a committed file the walker must exclude, rather than a
genuinely untracked file). Pure Go, no TUI surface, so no pty work. Decide early whether to take
a `.gitignore` dependency (`go-gitignore`, or shell out) or hand-roll — worth a findings note
either way.

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
| M1.5 | done   | -        | -         | see Handoff; build-only, not yet reviewed/validated |
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
