# Findings index

<!-- Per-cycle. One line per findings file. Agents scan this index and load only relevant
     topics. Findings files are condensed and cited (filepath:line, URLs) — never raw dumps.
     Prune by moving stale files to .charm/archive/ (/tidy-docs).
     At /charm-close each file is triaged: promoted to product docs, promoted to the durable
     `.charm/knowledge/` tier, or archived with the cycle (the default). Durable knowledge
     for cross-cycle reading lives there, not here — see .charm/knowledge/INDEX.md. -->

- `codebase-picker-integration.md` — where a Picker plugs into C1's code: the hand-duplicated
  modal dispatch, `render()`/Compositor overlay, the Find surface a grep hand-off would reuse,
  test/pty conventions, and the two new-ground gaps (no async `tea.Cmd` precedent; rg line
  numbers vs rendered viewport lines). Written at plan time; superseded by the shipped code
  once the Picker landed. Archived with the cycle, not promoted.

<!-- async-cmd-testing.md and picker-pty-async-ordering.md promoted to .charm/knowledge/ at
     C2 close (2026-08-05) — still true and useful next cycle, no better home. -->
