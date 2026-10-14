---
title: File Filter in the Files Nav (a toggle)
status: open
priority: 2
issue_type: feature
labels:
- needs-triage
depends_on:
  tk-0ta: parent-child
  tk-qei: blocks
created_at: 2026-10-14T18:19:25.184343349+00:00
updated_at: 2026-10-14T18:19:25.208143890+00:00
---

# Description

Add a per-Root File Filter, `Markdown` | `All` (D140).

- Session-only, defaulting to `Markdown`. Launching `noggin <file>` on a Text file starts its ad-hoc Root on `All`.
- `a` in the Files Nav toggles the Active Root's filter.
  - With Wikipedia active, `a` does nothing and shows a Footer message.
  - Shown in the status row, e.g. `Tree · - · All`.
- `All` = Markdown files plus Text files (binaries hidden by the sniff). It bounds only the `Tree`. `Recent` and `Starred` ignore it.
- When a toggle hides the selected file, the selection moves to the nearest surviving ancestor (as Refresh does). Expansion and open documents are kept.
- Staleness (D139) compares the filtered listing, so changes to hidden files don't raise the indicator.
- Git status markers apply to every shown file.

**Acceptance**
- Tests cover per-Root memory, the toggle, selection fallback, staleness and the launch case.
- README and the Help modal are updated.
