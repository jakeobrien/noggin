---
title: Keymap remaps (hjkl, t/a, alt+t/alt+a, ctrl+s, d/D, F, e)
status: closed
priority: 2
issue_type: feature
labels:
- ready-for-agent
depends_on:
  tk-95fr: parent-child
created_at: 2026-10-15T16:20:45.705131257+00:00
updated_at: 2026-10-15T16:57:53.753770736+00:00
closed_at: 2026-10-15T16:57:53.753770486+00:00
---

# Description

Parent: tk-95fr (see its "Remaps" section for the full list). Pure rebinds plus moving `Edit` from the global switch into the viewer case and `Rename` to `e` on the Sources Nav. Picker star merges into `keys.Star` (`ctrl+s`). The Picker's File Filter key moves to `alt+m` here too (its old `alt+a` becomes Picker scope); the Files Nav's `a` File Filter moves to `alt+m` in the toggles ticket.

Acceptance: keymap/help/footer/docs reflect the new keys; freed keys do nothing; `go test ./...` green.
