# Developer notes

Third-party quirks, environment constraints and dead ends that have no better home next to the code. How noggin itself works belongs in code comments; why it is shaped that way belongs in `docs/adr/`. Read the note for an area before working in it.

- `pty-harness.md`: how `make check-pty` drives the real binary, the startup handshake, why it needs a cell-grid emulator, and how to write pty checks that don't flake.
- `testing-async-commands.md`: the generation-counter shape for async `tea.Cmd` work and how to test it without sleeps or goroutines.
- `viewport-highlight-drift.md`: why bubbles' `viewport.SetHighlights` is unsafe on styled content and what Find does instead.
- `lipgloss-gotchas.md`: wrap-before-truncate, partial runs cancelling reverse video, and outer renders re-asserting SGR across newlines.
- `glamour-gotchas.md`: no frontmatter support, no goldmark extension hook, the once-per-process Chroma style, and `WithAutoStyle` hanging.
- `ignore-rules.md`: what the Tree's hand-rolled `.gitignore` matcher covers, and the ripgrep flags that keep the Pickers in line.
- `wikipedia.md`: Wikimedia endpoints, Parsoid quirks, the conversion pipeline's maths/image handling, and known limitations.
