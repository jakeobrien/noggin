# Findings index

<!-- Per-cycle. One line per findings file. Agents scan this index and load only relevant
     topics. Findings files are condensed and cited (filepath:line, URLs) — never raw dumps.
     Prune by moving stale files to .charm/archive/ (/tidy-docs).
     At /charm-close each file is triaged: promoted to product docs, promoted to the durable
     `.charm/knowledge/` tier, or archived with the cycle (the default). Durable knowledge
     for cross-cycle reading lives there, not here — see .charm/knowledge/INDEX.md. -->

- `git-status-gutter-approach.md` — evaluates a wrapped/unwrapped render-mode toggle for a per-line git-status gutter in the Document Viewer; why plain `WithWordWrap(0)` isn't enough, what `WithPreservedNewLines()` fixes, remaining line-mapping and horizontal-scroll gaps.
- `pty-harness-ESC-M-D-gap.md` — the pty harness's `Screen` emulator didn't handle bare `ESC M`/`ESC D` (RI/IND), which two-line row redraws provoke; symptom was indistinguishable from a real rendering bug, and it also masked a pre-existing weak assertion in the M6.2 Grep `alt+h` pty check.
