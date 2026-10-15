---
title: Tab / shift+tab cycle Focus across visible panes
status: open
priority: 2
issue_type: feature
labels:
- ready-for-agent
depends_on:
  tk-95fr: parent-child
created_at: 2026-10-15T16:20:45.710365591+00:00
updated_at: 2026-10-15T16:20:45.710365591+00:00
---

# Description

Parent: tk-95fr ("Tab focus cycle"). Reintroduces cycling that S5 (C3/M1.1) removed; record in D141.

Acceptance: order Sources -> Files -> DV1 -> DV2 forward, reverse on shift+tab; invisible panes skipped; single focusable pane is a no-op; never reveals the sidebar; inert in modal/input modes.
