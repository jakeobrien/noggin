---
title: 'Prefactor: one store for the File Filter'
status: open
priority: 2
issue_type: task
labels:
- ready-for-agent
depends_on:
  tk-0ta: parent-child
created_at: 2026-10-14T20:44:14.090354617+00:00
updated_at: 2026-10-14T20:44:14.090354617+00:00
---

# Description

## Parent

tk-0ta (Spec: Markdown/Text Render Mode and File Filter, D140)

## What to build

Make the File Filter live in one place. Today it is held three times: the per-Root filters in the Files Nav, the model's "most recently set" value (the Wikipedia fallback), and the Picker's own copy, which is synced by hand whenever the filter changes. The Files Nav status row, the Picker title, the Picker's search bounds and the Wikipedia fallback should all read the File Filter from one source of truth, so a toggle from either `a` or `alt+a` cannot leave a stale copy behind.

There is no change users can see. This prefactor makes the File Filter edge-case fixes simple.

Came from a code review of the D140 implementation: possible divergent state, with three copies of the File Filter.

## Acceptance criteria

- [ ] The per-Root File Filter, the most-recently-set fallback, and what the Picker shows and searches under come from a single source of truth. No hand-synced copy remains.
- [ ] The existing File Filter tests (Files Nav `a`, Picker `alt+a`, scope behaviour, Wikipedia fallback, title) pass unchanged.
- [ ] `go test ./...` is green.

## Blocked by

- None (can start immediately)
