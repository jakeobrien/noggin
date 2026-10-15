# noggin

A TUI for viewing markdown files with an integrated file/folder explorer and manager.


## Agent skills

### Issue tracker

Issues and specs live in minibeads (`mb` CLI, prefix `tk`, stored in `.minibeads/issues/`). See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: root `GLOSSARY.md` and `docs/adr/`. See `docs/agents/domain.md`.

### Developer notes

Gotchas for third-party libraries (Glamour, lipgloss, bubbles viewport), the pty harness, async commands, ignore rules and Wikipedia conversion live in `docs/dev/`. Read the matching note (indexed in `docs/dev/README.md`) before touching that area.
