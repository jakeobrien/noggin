# Findings index

<!-- Per-cycle. One line per findings file. Agents scan this index and load only relevant
     topics. Findings files are condensed and cited (filepath:line, URLs) — never raw dumps.
     Prune by moving stale files to .charm/archive/ (/tidy-docs).
     At /charm-close each file is triaged: promoted to product docs, promoted to the durable
     `.charm/knowledge/` tier, or archived with the cycle (the default). Durable knowledge
     for cross-cycle reading lives there, not here — see .charm/knowledge/INDEX.md. -->

- `sample-judging.md` — M3.2's per-article A1–A5 verdict on the four S5 samples (only
  `E (mathematical constant)` fails A3, on maths); cited by PROGRESS.md and feeds M3.3/M3.4.
- `maths-experiment.md` — M3.4's Unicode-vs-raw-text maths experiment (D104/S6a): the
  `github.com/doug/termtex` dependency, the per-formula single-line-or-fallback rule, the
  61%/39% Unicode/raw-text tally on a live fetch, and a recommendation for a future cycle.
