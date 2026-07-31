# Progress — C1 read-across-roots

## Cursor

Current task: M2.4 (next: M3.5, M3.6, M4.2 — reordered by replan, see below)

Last reviewed: M3.1-M3.4 (1bc7c04), 2026-07-31 — standards + spec-fidelity review of the M3
milestone; findings logged as T16-T26 in `TRIAGE.md`. M3.1-M3.4 have since been validated
(2026-07-31, see Handoff).

## Handoff

M4.1 done (2026-08-01): the edit handoff. `keys.Edit` (`e`, global — not pane-scoped, since S10
names "the active document" not "the focused pane") triggers `model.startEdit()`
(cmd/noggin/model.go), which no-ops into a footer message ("no document open to edit") when
nothing is open, otherwise shells out via `tea.ExecProcess(exec.Command($EDITOR or vi, path),
...)` — v2's release/resume is automatic, no manual terminal-state juggling needed. On exit,
`editFinishedMsg` (path + err) is handled by `model.finishEdit()`, which reuses `loadDocument`
and `openDocument` (same path as M2.3's activate and S1's launch) rather than duplicating the
read; an editor error or a since-vanished file degrades to a footer message, same shape as
every other refusal in the app. D13's no-requery constraint holds without extra work:
`markdownStyle` is a `sync.OnceValue`, so `finishEdit`'s re-render never touches the terminal.
Evidence: `make check` clean (5 new unit tests in `cmd/noggin/edit_test.go` covering both
refusal paths, the reload, and an editor-error/vanished-file pair); `make check-pty` clean (8
new M4.1 checks — a fake `$EDITOR` script appends a line to a real temp file, the key round-
trips it onto screen with the frame intact, the write is confirmed on disk, and the
no-document-open report is confirmed in the footer). Committed.

Replanned 2026-08-01 (charm-replan): TRIAGE.md's 19 open items (T9-T27) routed. Three new
tasks added — **M2.4** (internal/tree hardening: T11-T14, D23's case-sensitivity call),
**M3.5** (roots/messaging correctness fixes: T17, T18, T20), **M3.6** (M2/M3 code-review
cleanup batch: T9, T10, T16, T19, T21-T25, T27) — and **M4.2**'s scope/acceptance grew to fold
in T15 and T26 (effort 2→3). See D24 for the full rationale and D23 for the case-sensitivity
decision. TRIAGE.md is now empty. M2.4/M3.5/M3.6 are cleanup on already-`done` work and can
land whenever convenient before the M2/M3 code is considered closed out.

M3 milestone (M3.1-M3.4) validated fresh-context (2026-07-31, source: charm-validate): all four
PASS, no discrepancies vs. their PROGRESS.md claims. `make check` and `make check-pty` clean.
Strongest checks: M3.2's "survives a restart" ran as two separate real OS processes sharing an
isolated `$HOME`/`$XDG_CONFIG_HOME`; M3.3/M3.4 were re-exercised end-to-end via pty against the
real binary (cross-root byte-diff of the viewer pane), not just unit-level model assertions.

Next: **M2.4 — internal/tree hardening** (T11-T14, D23's case-sensitivity call; see D24 for
the full routing rationale). M4.1's edit handoff is unrelated to this area, so no carryover
context needed beyond what D24/TRIAGE already captured.

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
| M2.4 | todo   | -        | -         | replan C1/2026-08-01: internal/tree hardening (T11-T14, D23) |
| M3.1 | done   | 2026-07-31 | 2026-07-31 | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check |
| M3.2 | done   | 2026-07-31 | 2026-07-31 | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | 2026-07-31 | 2026-07-31 | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | 2026-07-31 | 2026-07-31 | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M3.5 | todo   | -        | -         | replan C1/2026-08-01: roots/messaging correctness fixes (T17, T18, T20) |
| M3.6 | todo   | -        | -         | replan C1/2026-08-01: M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27) |
| M4.1 | done   | -        | -         | edit handoff (S10/D3): tea.ExecProcess release/resume + finishEdit re-render from disk |
| M4.2 | todo   | -        | -         | replan C1/2026-08-01: scope grew (T15, T26 folded in), effort 2→3 |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
