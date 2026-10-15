---
title: View toggles alt+h / alt+i / alt+m shared by Files Nav and Picker, with status indicator
status: closed
priority: 2
issue_type: feature
labels:
- ready-for-agent
depends_on:
  tk-95fr: parent-child
created_at: 2026-10-15T16:20:45.720680382+00:00
updated_at: 2026-10-15T16:57:53.769053278+00:00
closed_at: 2026-10-15T16:57:53.769053195+00:00
---

# Description

Parent: tk-95fr ("View toggles").

Acceptance: session-wide hidden and gitignored flags shared by Files Nav and Picker; File Filter per Root on alt+m; Tree honours both flags (with the heavy-folder exclusion list when gitignored are shown, and muted styling for revealed gitignored entries); ripgrep honours both (--hidden glob / --no-ignore-vcs); selection falls back to nearest visible ancestor; right-aligned `h i  md`-style indicator on the Files Nav status row and in the Picker title.
