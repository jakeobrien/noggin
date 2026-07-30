# Glossary

- **Active Document** — the document currently rendered in the Document Viewer. Independent of
  the Active Root: switching roots does not change it.
- **Active Root** — the one Root whose contents the Files Nav is showing. Exactly one at a time.
- **Ad-hoc Root** — a Root supplied by the command line for this session only, not persisted.
  Listed first because it was named deliberately. At most one per session (C1).
- **Document Viewer** — the pane that renders the Active Document as styled markdown. Read-only;
  scrolls, has no cursor.
- **Edit** — the handoff: release the terminal to `$EDITOR` on the Active Document, then resume
  and re-render. Not in-app editing.
- **Files Nav** — the lower Navigator pane; the file tree of the Active Root.
- **Footer** — the persistent bottom bar: contextual key hints plus transient messages.
- **Nav Keys** — the movement keys shared by every pane (`h`/`j`/`k`/`l` and arrows, `g`/`G`),
  vim-oriented. Provisional, like all bindings.
- **Navigator** — the left sidebar: Roots Nav above, Files Nav below. Hideable.
- **Readable file** — a file noggin will show and render. Markdown-ish only in C1: `.md`,
  `.markdown`, `.txt`, `README`, `LICENSE`, `CHANGELOG`.
- **Root** — a folder that bounds file visibility. Several exist per session; identified by
  exact path; may not nest inside another Root.
- **Roots Nav** — the upper Navigator pane; the list of the session's Roots.
- **Starred** — bookmarked and persisted across sessions. Applies to Roots in C1, to documents
  in C4.
