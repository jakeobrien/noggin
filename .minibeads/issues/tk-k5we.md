---
title: ctrl+w closes/merges Document Viewers (replaces ctrl+x)
status: open
priority: 2
issue_type: feature
labels:
- ready-for-agent
depends_on:
  tk-95fr: parent-child
created_at: 2026-10-15T16:20:45.714723341+00:00
updated_at: 2026-10-15T16:20:45.714723341+00:00
---

# Description

Parent: tk-95fr ("ctrl+w").

Acceptance: DV1+DV2 -> DV1 takes DV2's state and DV2 closes; DV1 alone -> cleared to empty incl. history; DV2 focused -> DV2 closes, Focus to DV1; nav panes -> no-op.
