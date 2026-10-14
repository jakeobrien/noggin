# Findings index

<!-- Per-cycle. One line per findings file. Agents scan this index and load only relevant
     topics. Findings files are condensed and cited (filepath:line, URLs) — never raw dumps.
     Prune by moving stale files to .charm/archive/ (/tidy-docs).
     At /charm-close each file is triaged: promoted to product docs, promoted to the durable
     `.charm/knowledge/` tier, or archived with the cycle (the default). Durable knowledge
     for cross-cycle reading lives there, not here — see .charm/knowledge/INDEX.md. -->

- `frontmatter-rendering-options.md` — survey of ways to display or use YAML/TOML frontmatter
  (strip, fenced yaml, table/def-list, lipgloss card, toggle, title, goldmark ext) with cost,
  library comparison, prior art, edge cases and Find/D34/tags interactions; several verified by
  running Glamour v2.0.1. Read before revisiting D138/D10. Outcome: (d) shipped as D138, with (b) as its fallback.
