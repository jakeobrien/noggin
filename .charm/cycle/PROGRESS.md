# Progress — C1 read-across-roots

## Cursor

Current task: M2.2 (next: M2.3)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5 (below) has not yet been through review/validate — its Status row is blank
in both columns.

## Handoff

M2.1 is done (D18, `findings/gitignore-matching.md`): the readable-file policy and tree walk
live in a new `internal/tree` package. `Walk(root) (*Node, error)` returns a folder/file tree
— dirs always listed (S6: empty/no-readable folders still shown), files only if readable,
children sorted dirs-first then by name, `Path` absolute so M2.3's viewer can load without the
root. The allowlist is **one declaration**: `var readable = struct{exts, names ...}` in
`policy.go` (`.md`/`.markdown`/`.txt` + `README`/`LICENSE`/`CHANGELOG`, per S6). `.git` is
excluded by a built-in name check; `.gitignore` is hand-rolled (`gitignore.go`) — an ignored
directory is *pruned* (never descended), which is also why re-include-under-excluded-dir is
impossible for free.

Evidence: `make check` clean (build/vet/fmt/test); `TestWalkReadableSet` asserts the criterion
set **exactly** via `reflect.DeepEqual` — `{LICENSE, README, .charm/notes.md, docs/guide.md,
notes.txt}`: gitignored `ignored.md` excluded, `.git/` absent, `.charm/notes.md` present,
`main.go` absent, `README` present, a gitignored `build/` dir pruned, `empty/` still listed.
The gitignore matcher has its own 18-case table + negation + nested-precedence tests. Fixture
is built programmatically under `t.TempDir()` (D18) — no committed fixture, no git-add
awkwardness. `make check-pty` not run: pure Go, no TUI/runtime surface touched.

Not yet through review/validate (Status table blank). Also **M1.5 remains unvalidated** — the
M1→M2 gate was crossed with the user's go-ahead to build M2.1; validate M1.5 alongside M2 at
the next checkpoint.

Next: **M2.2 — Files Nav pane.** Render `tree.Node` in the Files Nav pane: open with only
top-level entries (collapsed), move selection across the expanded set, expand/collapse a folder
in place, and still list a folder with no readable content. This is the first consumer of
`internal/tree` and the first TUI surface since M1 — so `make check-pty` is back in play, and
end-to-end pane behaviour follows the `test/pty/` harness recipe. `Walk` gives absolute `Path`
on every node; `Children` is already sorted (dirs first, then name). Note the pane needs an
expansion/selection model layered over the static `Node` tree — the tree itself is immutable
data, not UI state.

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
