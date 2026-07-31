# Progress — C1 read-across-roots

## Cursor

Current task: M3.1 (next: M3.2)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5 (below) has not yet been through review/validate — its Status row is blank
in both columns.

## Handoff

M2.3 is done: activating a file makes it the active document. `model.activate()`
(`cmd/noggin/model.go`) reads `m.files.selectedNode()`; a nil node or a directory is a no-op
(folder-enter is deliberately left for later, per the plan's own note); a file is passed to
`loadDocument` and, on success, into the existing `model.openDocument(d)` — which already
moved focus to the viewer, so M2.3 added no new focus-routing logic there. A read error
returns `m` unchanged rather than exiting: no footer-message plumbing exists yet (that's
M4.2), so a failed activate is silent-but-safe rather than surfaced — worth revisiting once
the footer can carry transient messages. `keys.Activate` (`enter`) is a new keymap binding,
routed in `model.Update`'s `filesPane` case ahead of `filesNav.handleKey` so it does not need
to flow through the pane's own `handleKey` (which only ever returns a `filesNav`, not a
`model` — `openDocument`/`activate` need model-level access to the viewer). `keymap_test.go`
gained an `"enter"` entry in `namedKeys` (T-Enter round-trips via `tea.KeyEnter`, same pattern
as `tab`/`up`/etc.).

Evidence: `make check` clean. Three new unit tests in `files_test.go`:
`TestActivateOpensSelectedFileAndMovesFocus` (walks the real `testdata/tree` fixture — unlike
the hand-built `fixtureTree`, whose nodes have no backing files, this one needs `loadDocument`
to succeed — activates `README.md`, checks focus is `viewerPane`, the heading renders, the pane
title names the file, and a subsequent `LineDown` moves the viewer's offset rather than a Files
Nav selection), `TestActivateOnAFolderIsANoOp` (activating the initially-selected folder leaves
focus on `filesPane` and the viewer unopened), `TestActivateWithAReadErrorLeavesSessionRunning`
(a node pointing at a non-existent path leaves focus and the viewer untouched — no panic, no
exit). `make check-pty` now 72 checks (was 67): launches the fixture tree, steps onto
`README.md`, presses enter, and checks the footer reports `Document` focus, the file's heading
is on screen, the pane title names it, a half-page-down does not error, and the session still
exits 0.

Not yet through review/validate (Status table blank for M2.1-M2.3). **M1.5 remains
unvalidated** too — the M1→M2 gate was crossed with the user's go-ahead; validate
M1.5/M2.1/M2.2/M2.3 together at the next checkpoint (M2 is now complete).

Next: **M3.1 — resolve the session's roots from the launch argument (S1).** This replaces,
rather than supplements, `main.go`'s current temporary dir/file-argument scaffolding (dir →
`openFolder`, file → `openDocument`, both `os.Exit(1)` on error) — per T7 and the plan's own
note, a bad path must become a rejection message plus a session that still starts, not a
process exit. The table-driven cases the plan lists (bare launch, dir argument, file argument,
already-starred dir, non-existent path, subdirectory of a starred root, parent of a starred
root, more than one argument) imply a starred-roots concept that M3.2 (the store) has not been
built yet — M3.1's acceptance criterion talks about "already a starred root" cases, so either
M3.1 needs a minimal in-memory roots concept ahead of M3.2's persistent store, or those specific
cases wait for M3.2 to land; worth a fresh read of S1's spec text before starting, since PLAN.md
alone under-specifies the roots data shape at this boundary. No footer-message mechanism exists
yet for rejections either (M4.2) — M3.1 will need to decide where a rejection message lives in
the meantime (a temporary field on `model`, most likely) or explicitly punt display to M4.2 and
just ensure the session still starts.

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
| M2.3 | done   | -        | -         | Activate binding (enter): filesNav → loadDocument → openDocument; 3 unit tests + 5 pty checks |
| M3.1 | todo   | -        | -         |      |
| M3.2 | todo   | -        | -         |      |
| M3.3 | todo   | -        | -         |      |
| M3.4 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
