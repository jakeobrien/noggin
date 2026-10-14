---
title: 'File Filter edge cases: alt+a feedback in Wikipedia mode, launched Text file in a starred Root'
status: open
priority: 2
issue_type: bug
labels:
- ready-for-agent
depends_on:
  tk-0ta: parent-child
  tk-ybh: blocks
created_at: 2026-10-14T20:44:32.138138251+00:00
updated_at: 2026-10-14T20:44:32.140374959+00:00
---

# Description

## Parent

tk-0ta (Spec: Markdown/Text Render Mode and File Filter, D140)

## What to build

Two File Filter cases where the user currently gets no visible result:

1. **`alt+a` in Wikipedia mode.** With Wikipedia active, the Picker uses the most recently set File Filter, and `alt+a` toggles that value (D140). The Picker title deliberately shows no Scope or filter in Wikipedia mode, so the toggle currently changes nothing on screen. Pressing `alt+a` there should show a Footer message naming the new File Filter value. The title stays unchanged.
2. **Launching a Text file inside a starred Root.** D140 says an ad-hoc Root created by launching `noggin <text file>` starts on `All`, so the launched file shows in the Files Nav. When the launch directory is already a starred Root, that Root currently stays on `Markdown`: the Text file opens but never appears in the Files Nav. In that case the Root should switch to `All` for the session, and the Files Nav should show the file. This also sets the most recently set File Filter value, as any other filter change does.

## Acceptance criteria

- [ ] In Wikipedia mode, `alt+a` shows a Footer message with the new File Filter value. The Picker title is unchanged.
- [ ] The toggled value is the one the next File/Grep Picker opened from Wikipedia uses.
- [ ] `noggin <text file>` inside a starred Root starts that Root on `All`, and the file is visible and selected in the Files Nav.
- [ ] Launching a Markdown file inside a starred Root leaves that Root on `Markdown`.
- [ ] Tests cover all of the above. README and Help are updated where they describe these cases.

## Blocked by

- tk-ybh (Prefactor: one store for the File Filter)
