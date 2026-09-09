# Findings index

<!-- Per-cycle. One line per findings file. Agents scan this index and load only relevant
     topics. Findings files are condensed and cited (filepath:line, URLs) — never raw dumps.
     Prune by moving stale files to .charm/archive/ (/tidy-docs).
     At /charm-close each file is triaged: promoted to product docs, promoted to the durable
     `.charm/knowledge/` tier, or archived with the cycle (the default). Durable knowledge
     for cross-cycle reading lives there, not here — see .charm/knowledge/INDEX.md. -->

- `wikimedia-api.md` — the title-search endpoint (title + short description) and how the
  page-HTML endpoint's 307 redirect yields the canonical URL/title; verified live 2026-09-07.
  Read before M2.2/M2.3.
- `lipgloss-reflow-sgr-continuity.md` — a third `lipgloss-v2-row-rendering-gotchas.md`-family
  trap: an outer `Style.Render`'s own word-wrap re-asserts SGR state across a `"\n"` a caller's
  content already contains, so a custom attribute left merely "active" (not explicitly cleared)
  at such a boundary leaks into whatever the outer render draws next. Found building M9.1's
  background tint. Reread before any future per-line SGR injection that flows through a second
  `Style.Render` (a `Width()`/`Height()`/border wrapper, `renderBox`'s own shape).
