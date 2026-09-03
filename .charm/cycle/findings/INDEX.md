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
- `images-experiment.md` — M3.5's caption/alt/filename placeholder experiment (D104/S6a): no
  new dependency, the `renderImages` pipeline stage and its fallback order, live-sample
  results across all four S5 titles, a pre-existing chrome-stripping gap it surfaced
  (portal bars, sister-project boxes — routed to TRIAGE, not fixed here), and a
  recommendation for a future cycle.
- `m4.1-conversion-record.md` — M4.1's consolidated verdict record: what converts well/badly
  across a fresh post-M3.8 regen (cites `/tmp/s5scratch`, stale repo-root/`tmp/` sample
  copies removed), the final selector/cleanup list, an S6a experiment summary table citing
  `maths-experiment.md`/`images-experiment.md`, and a maintenance-template citation leak
  (TRIAGE T1) found during this task and fixed post-M4.1 as M3.9 (D111) — record updated in
  place to reflect the fix; M4.1 itself is re-opened pending a fresh regen citation.
