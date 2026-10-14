---
title: File Filter in the Picker (alt+a)
status: closed
priority: 2
issue_type: feature
labels:
- needs-triage
depends_on:
  tk-0ta: parent-child
  tk-3vr: blocks
created_at: 2026-10-14T18:19:25.187205265+00:00
updated_at: 2026-10-14T19:34:23.553916119+00:00
closed_at: 2026-10-14T19:34:23.553915869+00:00
---

# Description

Apply the File Filter to the File and Grep Pickers (D140).

- The Active Root and All Roots scopes use the **Active Root's** filter, applied uniformly to every Root searched.
- With Wikipedia active, the Picker uses the most recently set filter value (or `Markdown`), and `alt+a` toggles that value.
- The Starred scope ignores the filter.
- `alt+a` toggles the filter. It is the same per-Root state as the Files Nav's `a`.
- The Picker title gains the filter, e.g. `Grep · All Roots · All`.
- `All` means non-binary files: filter `rg --files` output with the binary sniff. Grep already skips binaries.

**Acceptance**
- Tests cover scope behaviour, the Wikipedia fallback value and the title.
- README and the Help modal are updated.
