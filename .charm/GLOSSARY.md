# Glossary

- **Active Document** — the document currently rendered in the Document Viewer. Independent of
  the Active Root: switching roots does not change it.
- **Active Root** — the one Root whose contents the Files Nav is showing. Exactly one at a time.
- **Ad-hoc Root** — a Root supplied by the command line for this session only, not persisted.
  Listed first because it was named deliberately. At most one per session (C1).
- **Broken Root** — a Root whose path no longer exists on disk. Still listed and its starred
  state kept, but marked and cannot be activated.
- **Committed Find** — a Find query locked in with Activate, as distinct from one still being
  edited. Its match and indication persist until dismissed or the viewport is forced to
  re-render (e.g. a terminal resize or a Navigator toggle).
- **Document Viewer** — the pane that renders the Active Document as styled markdown. Read-only;
  scrolls, has no cursor.
- **Edit** — the handoff: release the terminal to `$EDITOR` on the Active Document, then resume
  and re-render. Not in-app editing.
- **Files Nav** — the lower Navigator pane; the file tree of the Active Root.
- **Find** — the `/`-style incremental search over the Active Document: next/previous
  traversal wraps, case-insensitive unless the query has uppercase. See Committed Find.
- **Focus** — which pane currently receives keyboard input. Exactly one pane at a time; most
  keys route to the focused pane.
- **Footer** — the persistent bottom bar: contextual key hints plus transient messages.
- **Help Modal** — the modal listing the focused pane's key bindings, reachable from the
  footer's help affordance and dismissible with one key.
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
