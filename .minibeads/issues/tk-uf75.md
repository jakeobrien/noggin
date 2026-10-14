---
title: Toggling Render Mode while Peeking keeps your place
status: closed
priority: 2
issue_type: bug
labels:
- ready-for-agent
depends_on:
  tk-0ta: parent-child
  tk-ci56: blocks
created_at: 2026-10-14T20:44:32.122698126+00:00
updated_at: 2026-10-14T20:58:13.933619934+00:00
closed_at: 2026-10-14T20:58:13.933619767+00:00
---

# Description

## Parent

tk-0ta (Spec: Markdown/Text Render Mode and File Filter, D140)

## What to build

D140 makes Render Mode sticky "across every document opened, Peeked or walked to", and toggling should carry over what the reader was looking at. During a Peek, `m` currently refuses based on the Peeked document (correct). But the scroll and find carry-over reads the Active Document's state, and the Peek re-renders scrolled back to the top.

After this, pressing `m` while Peeking a Markdown file re-renders the Peek in the other mode at the same proportional scroll position. The Active Document underneath keeps its own scroll position and find state, so ending the Peek returns the reader exactly where they were, in the new mode.

## Acceptance criteria

- [ ] Toggling Render Mode during a Peek keeps the Peek's scroll position as a proportion of the document; it does not jump to the top.
- [ ] The Active Document under the Peek keeps its scroll position and its find query and current match across the toggle, carried as D140 specifies for a normal toggle.
- [ ] `m` on a Peeked Text file is still refused with the existing message.
- [ ] Tests cover the Peek toggle in both DV1 and DV2.

## Blocked by

- tk-ci56 (Prefactor: one place decides Text vs Markdown Render Mode in the viewer)
