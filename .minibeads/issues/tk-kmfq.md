---
title: Clean up stale wording and record accepted D140 deviations
status: open
priority: 3
issue_type: chore
labels:
- ready-for-agent
depends_on:
  tk-0ta: parent-child
created_at: 2026-10-14T20:44:32.135453959+00:00
updated_at: 2026-10-14T20:44:32.135453959+00:00
---

# Description

## Parent

tk-0ta (Spec: Markdown/Text Render Mode and File Filter, D140)

## What to build

Bring the docs and vocabulary in line with what shipped:

- **README.** Remove the claim that toggling Render Mode "resets any find in progress". tk-edt replaced that behaviour: the query and match are kept, as the keybindings doc already says. Also correct the claim that only Markdown files are ever searched or shown by File and Grep. Under the `All` File Filter, Text files are too.
- **Retired term.** The Glossary retires "Readable file" (D140). Remove it from test names, test comments and test strings, using Markdown file / Text file / File Filter as appropriate.
- **D140 amendment.** Record the deviations from the spec text that were reviewed and accepted:
  - Text rendering drops a leading BOM and trailing carriage returns, and shows control characters as U+FFFD, instead of "the source exactly as on disk".
  - The Picker's Starred scope ignores the File Filter but still hides binary and non-regular files, since they cannot be opened.
  - In Text Render Mode, a Grep landing falls back to the first match on ripgrep's line when the column does not match a find hit exactly.

## Acceptance criteria

- [ ] The README describes toggling as keeping the find query and current match.
- [ ] The README says what File and Grep search under each File Filter value.
- [ ] No occurrence of "Readable" (in the retired sense) remains in code, tests or docs.
- [ ] D140 records the three accepted deviations above, each with a one-line rationale.
- [ ] `go test ./...` is green.

## Blocked by

- None (can start immediately)
