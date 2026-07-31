# Progress — C1 read-across-roots

## Cursor

Current task: M1.5 (next: M2.1)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`.

## Handoff

M1.4 is done: the whole stack is on Charm v2 (`charm.land/{bubbletea,bubbles,lipgloss,glamour}/v2`),
one atomic swap, no design change. ~150 lines across 8 files. `findings/bubbletea-v2-migration.md`
now closes with what that note got right, wrong and missed — read that section rather than
re-deriving. Everything from here is v2-only.

What actually changed, and the three things that will trip you up:

- **`renderPane` stopped pre-deducting the border.** lipgloss v2's `Style.Width()/Height()`
  include it, so the pane box now takes the region's outer size straight through. `layout.go`
  is untouched and still correct: `rect.content()` is the box a pane's *contents* get, which is a
  different quantity from the box's own size. Verified directly — a bordered `Width(20).Height(5)`
  measures 20x5.
- **`keymap.go` grew a hand-rolled key parser, `keyPress`.** `key.Matches` compares
  `msg.String()` to the binding's key strings, and v2 ships no public parser for its own key
  names (`Key.Keystroke` formats one way only; ultraviolet's name table is internal). D2 says
  tests must not spell key literals, so synthesising a key from a binding's name means inverting
  `Keystroke` by hand. It is guarded by `TestKeyPressRoundTrip` (every key in every binding round
  trips, and `key.Matches` accepts the result) plus `TestEveryBindingIsCovered`, a reflection
  check so a new binding cannot escape the round trip. **Add non-printable keys to `namedKeys`**
  — `keyPress` panics with instructions if you name one it does not know.
- **`viewport` needs `FillHeight = true`.** v2 stopped padding short content out to the
  viewport's height, and the pane contract is a content box of exactly its size.

`expandTabs` survives, and D14's open question is settled the other way from what it guessed
(D16). Glamour v2 still emits literal tabs; lipgloss v2 does expand them in `Render` but
`lipgloss.Width` still measures a raw tab as one cell, so the viewport under-slices the line and
the pane box wraps the overflow onto an extra row — **the frame ends up one row taller than the
terminal**, which on a real screen scrolls the top border off. A single screenful cannot see it.
`TestFrameStaysWithinTheTerminalWhileScrolling` walks the whole document a page at a time to
catch it. Removing the `expandTabs` call is the standing negative control and fails both layers.

**The pty harness is checked in now (D15): `test/pty/{harness.py,checks.py}`, `make check-pty`.**
It had been rebuilt from `/tmp` twice. Re-derived against v2 and much less trivial than the v1
recipe implied: v2 sends different startup queries (BEL-terminated OSC 11, Primary DA, two
DECRQM, *no* cursor-position report), and its ultraviolet renderer diffs cells and edits rows in
place — blank runs arrive as ECH+CUF, cursor moves as literal tabs, shrinking text as DCH. So
stripping escapes is wrong in both directions and the harness carries a small screen emulator.
A missing DCH in it read the footer as `Roots nt · q quit` and looked exactly like a product
focus bug; budget for that class of confusion. Assert **text and geometry off the cell grid,
colour off the raw byte stream**. Full recipe in `findings/pty-verification.md`.

Verified: `make check` clean; 43 test cases pass (17 tests, 2 new); `make check-pty` 59 checks at
80x24 and 100x30. The criterion's clauses, each with evidence: zero `github.com/charmbracelet/`
imports of the four modules (rg, no matches); zero key literals outside `keymap.go` (rg over all
`.go`, no matches); the frame exactly 80x24 and 100x30 with all three regions, both border
colours on the wire, the full focus cycle read off the footer in both directions, and resize
reflow; the fixture rendered with all three frontmatter lines absent, no literal tab on screen,
box-drawing table, `ctrl+d` reaching the code fence and `G` the last line with the frame's border
intact throughout; all five scroll actions through all 8 bindings; exit 0 and terminal restored
on both `q` and `ctrl+c`. An independent code review found no functional defect; its two
actionable findings (D2 spirit in the harness's key bytes, and untested modifier-key paths) are
addressed — the harness now routes every key byte through one documented `KEY` table and
exercises `shift+tab` and `ctrl+b` against the real v2 parser.

M1 validated 2026-07-31 (fresh-context subagent, independent of the M1.4 build session): M1.1-M1.4
all PASS, every quantitative claim in this handoff and the findings (59 pty checks, 43/17 test
counts, zero charmbracelet imports, zero key literals outside `keymap.go`, the `expandTabs`
negative control) reproduced exactly on independent re-run. One residual gap, not a defect: the
pty harness proxies terminal-restore behaviour via escape codes (`\x1b[?25h`, `\x1b[?2004l`) but
a human has not visually confirmed a real terminal returns cleanly after `q`/`ctrl+c` — worth a
manual spot-check during the dogfood week rather than blocking here.

**Replan 2026-08-01:** triaged the M1.1-M1.4 checkpoint review's T1-T8 (`TRIAGE.md` is now
empty). T3-T5 (test-only key helpers compiled into the shipped binary; tests reaching into
`bubbles/viewport` internals; `pane.title()`/glossary drift) became `M1.5`, a small cleanup
task (D17) inserted ahead of M2.1 — cheap and fresh from review. T6 (bare module path) went to
`ROADMAP.md` Unscheduled. T7 amended M3.1's wording to make explicit that it *replaces* (not
supplements) `main.go`'s pre-S1 `os.Exit(1)` scaffolding. T8's Status-table regression is fixed
below. T1/T2 dropped as already-handled (self-documenting non-choices, no action pending).

Next: **M1.5** first (see PLAN.md), then **M2.1 — readable-file policy and tree walk.** Nothing exists for it: no filesystem
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
| M1.5 | todo   | -        | -         | added by replan 2026-08-01 (T3-T5, D17) |
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
