# Glossary

- **Activate** — commit to a Picker's selected Result: the file becomes the Active Document
  (a Grep Result also lands the Document Viewer on the matched line). Does not change the
  Active Root or the Files Nav tree (C2).
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
- **File Picker** — the Picker whose Results are files, matched fuzzily (C2).
- **Files Nav** — the lower Navigator pane; the file tree of the Active Root.
- **Find** — the `/`-style incremental search over the Active Document: next/previous
  traversal wraps, case-insensitive unless the query has uppercase. See Committed Find.
  *Find is always within one document*; searching across files is a Picker.
- **Focus** — which pane currently receives keyboard input. Exactly one pane at a time; most
  keys route to the focused pane.
- **Footer** — the persistent bottom bar: contextual key hints plus transient messages.
- **Grep Picker** — the Picker whose Results are matching lines across files (C2).
- **Help Modal** — the modal listing the focused pane's key bindings, reachable from the
  footer's help affordance and dismissible with one key.
- **Peek** — render the Files Nav's selected document in the Document Viewer *without* making
  it the Active Document and without moving Focus. Transient and read-only: cannot be scrolled
  or searched, marked by an inverted header, cleared by any selection change, Focus change,
  Picker open or Refresh — restoring the Active Document at its previous scroll position.
  `<enter>` is the only promotion path. Distinct from the Picker's Preview pane (C2a).

- **Nav Keys** — the movement keys shared by every pane (`h`/`j`/`k`/`l` and arrows, `g`/`G`),
  vim-oriented. Provisional, like all bindings.
- **Navigator** — the left sidebar: Roots Nav above, Files Nav below. Hideable.
- **Picker** — the modal fuzzy-finding UI: a Prompt, a Results list, and a Preview pane.
  Comes in modes — File Picker and Grep Picker (C2). "Finder" is deliberately avoided; it
  means the OS file manager.
- **Preview pane** — the part of a Picker that shows the selected Result's file. Lives inside
  the Picker only; it never changes the Active Document (C2).
- **Prompt** — a Picker's input field; Results update as it is typed (C2).
- **Refresh** — the on-demand re-read of the Active Root's file tree and the Active Document
  from disk, preserving tree expansion, selection and scroll position. Manual only; noggin never
  watches the filesystem (C2a).

- **Readable file** — a file noggin will show and render. Markdown-ish only in C1: `.md`,
  `.markdown`, `.txt`, `README`, `LICENSE`, `CHANGELOG`. Bounds every Picker too (C2).
- **Result** — one row in a Picker: a file (File Picker) or a file-and-line (Grep Picker) (C2).
- **Results list** — the Picker pane listing its Results, one row per Result, below the
  Prompt (C2).
- **Root** — a folder that bounds file visibility. Several exist per session; identified by
  exact path; may not nest inside another Root.
- **Roots Nav** — the upper Navigator pane; the list of the session's Roots.
- **Scope** — what a Picker searches: the Active Root, or All Roots (C2).
- **Staleness indicator** — a passive UI signal that the Active Document or the Active Root's
  tree changed on disk. Reports only that *something* changed, and never refreshes on its own
  (C2a).

- **Starred** — bookmarked and persisted across sessions. Applies to Roots in C1, to documents
  in C4.
