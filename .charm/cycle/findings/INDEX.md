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
  numbers vs rendered viewport lines). Written at plan time.
- `async-cmd-testing.md` — the generation-counter shape of the Picker's debounced async
  search (D41) and how it is tested without sleeps, goroutines or a test seam: synthesize the
  tick messages, treat a non-nil Cmd as "a search was issued", call the search Cmd inline.
- `picker-pty-async-ordering.md` — an ordering hazard (not a product bug) found writing M7.1's
  pty checks: switching the Picker's mode/Scope immediately after a landed async search races
  the pty harness's cell-diffing replay intermittently; switch on an empty Prompt (no search
  in flight) instead. Reread before adding pty coverage for a future async command.
