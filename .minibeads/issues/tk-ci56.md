---
title: 'Prefactor: one place decides Text vs Markdown Render Mode in the viewer'
status: closed
priority: 2
issue_type: task
labels:
- ready-for-agent
depends_on:
  tk-0ta: parent-child
created_at: 2026-10-14T20:44:14.094393284+00:00
updated_at: 2026-10-14T20:53:14.336879916+00:00
closed_at: 2026-10-14T20:53:14.336879750+00:00
---

# Description

## Parent

tk-0ta (Spec: Markdown/Text Render Mode and File Filter, D140)

## What to build

Collapse the Document Viewer's scattered Text-versus-Markdown branching into one place. Today Find, current-match indication, Grep landing, rendering and the Render Mode toggle each decide the mode with their own check: whether a Text layout exists, a per-match `inText` tag, a was/is-Text pair in the match carry-over, or the viewer's mode-for-document. The viewer should resolve the effective Render Mode for the shown document once. Mode-specific behaviour (find, indicate match, Grep landing) then hangs off that single decision, not off repeated conditionals.

In the same change, give the two functions both named "is text file" honest names. One means "not a Markdown file" (viewer side). The other means "a non-binary regular file" (tree side). The Picker's Starred scope should also call a clearly named "not binary" check, not borrow the `All` filter's predicate.

There is no change users can see.

Came from a code review of the D140 implementation: possible Repeated Switches, possible tag-field Primitive Obsession on find matches, and a misleading duplicate name.

## Acceptance criteria

- [ ] The effective Render Mode for the shown document is decided in one place. Find, match indication, Grep landing and toggle no longer each re-derive it.
- [ ] No function or field acts as a Text/Markdown type tag on a find match.
- [ ] The two "is text file" functions have distinct names that match their meanings (Glossary: Markdown file, Text file).
- [ ] Existing Render Mode, Find, Grep-landing and toggle tests pass unchanged; `go test ./...` is green.

## Blocked by

- None (can start immediately)
