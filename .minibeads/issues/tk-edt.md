---
title: Preserve find match and scroll when toggling Render Mode
status: open
priority: 2
issue_type: feature
labels:
- needs-triage
depends_on:
  tk-0ta: parent-child
  tk-2g3: blocks
created_at: 2026-10-14T18:19:25.178647474+00:00
updated_at: 2026-10-14T18:19:25.202343932+00:00
---

# Description

When `m` toggles Render Mode on an open document (D140):

- **With a find active (in progress or committed):** keep the query, re-run it in the new mode, and keep the current match by occurrence ordinal (D34's mapping, in either direction).
- **Grep landing that fell back:** if the viewer's current match came from a Grep landing that fell back to the first match in Markdown, switching to Text lands on ripgrep's exact location instead. Keep rg's (line, col) with the find state for this.
- **With no find active:** keep the same proportional scroll position.

**Acceptance**
- Tests cover all three cases, in both directions.
