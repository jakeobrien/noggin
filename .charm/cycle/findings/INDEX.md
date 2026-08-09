# Findings index

<!-- Per-cycle. One line per findings file. Agents scan this index and load only relevant
     topics. Findings files are condensed and cited (filepath:line, URLs) — never raw dumps.
     Prune by moving stale files to .charm/archive/ (/tidy-docs).
     At /charm-close each file is triaged: promoted to product docs, promoted to the durable
     `.charm/knowledge/` tier, or archived with the cycle (the default). Durable knowledge
     for cross-cycle reading lives there, not here — see .charm/knowledge/INDEX.md. -->

- `pty-ctrl-enter-encoding.md` — how to put a `ctrl+enter`-shaped chord on a pty's wire
  when it has no raw C0 byte: the `CSI u` form (`\x1b[13;5u`) that ultraviolet's decoder
  parses unconditionally, regardless of Kitty/CSI-u negotiation; cited by
  `test/pty/checks.py`'s `ActivateDV2CtrlEnter` KEY entry (C3/M1.7).
