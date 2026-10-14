---
title: Picker Preview uses the Text renderer
status: closed
priority: 2
issue_type: task
labels:
- needs-triage
depends_on:
  tk-0ta: parent-child
  tk-2g3: blocks
created_at: 2026-10-14T18:19:25.181906807+00:00
updated_at: 2026-10-14T19:23:25.110349763+00:00
closed_at: 2026-10-14T19:23:25.110349680+00:00
---

# Description

Replace the Picker Preview's bare `SetContent(raw)` (`syncPreview`) with the Text-mode renderer (D140):

- soft wrap;
- tabs expanded;
- the line-number gutter;
- an exact span highlight on the Grep match instead of reverse-videoing the whole line.

The Preview stays raw text for every file, Markdown files included.

**Acceptance**
- The Preview for a Grep Result matches what a Text-mode viewer shows after activation.
