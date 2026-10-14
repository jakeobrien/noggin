# Glossary

- **Activate** — commit to a Picker's selected Result: the file becomes the Active Document
  (a Grep Result also lands the Document Viewer on the matched line). Does not change the
  Active Root or the Files Nav tree (C2).
- **Active Document** — the document currently rendered in the Document Viewer. Independent of
  the Active Root: switching roots does not change it.
- **Active Root** — the one Root whose contents the Files Nav is showing. Exactly one at a time.
  Generalized to **Active Source** in C6: wherever a Query Source can be the active one, the
  term (and scope label) is Active Source; "Active Root" survives where only Roots qualify
  (Grep).
- **Ad-hoc Root** — a Root supplied by the command line for this session only, not persisted.
  Listed first because it was named deliberately. At most one per session (C1).
- **Article** — a document served by the Wikipedia Source. Identified by its canonical URL
  (redirects resolved) and title; noggin stores that pointer plus star/last-viewed, never the
  content (C6).
- **Broken Root** — a Root whose path no longer exists on disk. Still listed and its starred
  state kept, but marked and cannot be activated. A **broken starred document** gets the same
  treatment in a Starred list (C4).
- **Browse Mode** — how the Files Nav lists documents: `Tree`, `Recent` or `Starred`. Cycled
  by key from the Files Nav, remembered per Root for the session (C4).
- **Browse Scope** — what a Browse Mode is drawn from. `Tree` has none; `Recent` is Active Root
  or All Roots; `Starred` is Active Root, All Starred or Orphans. Remembered per (Root, Mode)
  for the session (C4). Distinct from a Picker's Scope. C6 relabels `Recent`'s scopes
  `Active Source` | `All Sources` and `Starred`'s first scope `Active Source`; with a Query
  Source active there is no `Tree` and no `Orphans`.
- **Chrome** — everything a fetched Wikipedia page carries that is not the article itself:
  navboxes, sidebars (e.g. "Part of a series on X" boxes; CSS class `sidebar`, distinct from
  navbox), infoboxes, edit-section links, citation superscripts, reference lists, metadata
  tables, images, portal bars (a trailing list of linked subject-area icons, CSS class
  `portal-bar`), sister-project side-boxes (a boxed "Wikimedia Commons has media related
  to..."-style template, CSS class `sistersitebox`), and inline sister-project links (the
  same family of template rendered as a bare `<li>`/`<p>` with no wrapping box element,
  matched on wording rather than a selector). Stripped before conversion; content tables are
  not chrome. Images are chrome too, but as of C4a's amendment (D104) leave a text
  placeholder — caption, falling back to alt-text, falling back to filename — rather than
  vanishing without trace (C4a).
- **Collection** — north-star language for a class of knowledge noggin admits (Notes,
  Wikipedia, Dictionary). Not an in-app entity: the entity is the **Source** (C6, resolving
  C4a's provisional definition).
- **Committed Find** — a Find query locked in with Activate, as distinct from one still being
  edited. Its match and indication persist until dismissed or the viewport is forced to
  re-render (e.g. a terminal resize or a Navigator toggle).
- **Document History** — a Document Viewer's own back/forward stack of the documents opened in
  it. Per viewer, session-only, browser semantics, never changes the Active Root (C4).
- **Document Store** — the persisted, path-keyed record of every document noggin knows about:
  whether it is Starred and when it was last viewed. Not attached to Roots; per-Root Starred and
  Recent lists are derived by path (C4).
- **Document Viewer** — the pane that shows the Active Document in that viewer's Render Mode
  (styled markdown or raw text, D140). Read-only;
  scrolls, has no cursor. A session has one always-present Document Viewer (DV1) and may have
  a second, optional one (DV2) beside it; the two are fully independent — separate Active
  Document, Peek, Find and scroll position each (C3). See Targeting modifier for how a Files
  Nav or Picker action chooses which one to load into.
- **Edit** — the handoff: release the terminal to `$EDITOR` on the Active Document, then resume
  and re-render. Not in-app editing.
- **File Filter** — which files a Root offers for discovery: `Markdown` (Markdown files only)
  or `All` (every Text file too). Remembered per Root for the session, default `Markdown`.
  Bounds the Files Nav `Tree` and the File/Grep Pickers' Active Root and All Roots scopes;
  never Recent or Starred (D140). _Avoid_: file mode, show-all.
- **File Picker** — the Picker whose Results are files, matched fuzzily (C2).
- **Files Nav** — the lower Navigator pane; the file tree of the Active Root — or, when a
  Query Source is active, its Starred / Recent lists only (C6).
- **Find** — the `/`-style incremental search over the Active Document: next/previous
  traversal wraps, case-insensitive unless the query has uppercase. See Committed Find.
  *Find is always within one document*; searching across files is a Picker. Exact (matched
  span, source positions) only in Text Render Mode; in Markdown it matches the rendered
  output and indicates the whole line (D140).
- **Focus** — which pane currently receives keyboard input. Exactly one pane at a time; most
  keys route to the focused pane. DV1 and DV2 are separately focusable panes (C3): once a
  Document Viewer holds Focus, its local commands (Find, Edit, scrolling) act on that
  viewer's Active Document only.
- **Focus Badge** — the key that focuses a pane, shown right-aligned on that pane's title row
  (`[f]`, `[^o]`). Dropped when the pane is too narrow (C4).
- **Footer** — the persistent bottom bar: contextual key hints plus transient messages.
- **Frontmatter Card** — a document's YAML frontmatter drawn as a bordered key/value box
  above its rendered body. Frontmatter it cannot show faithfully is shown verbatim as a code
  block instead (D138).
- **Grep Picker** — the Picker whose Results are matching lines across files (C2). Never
  includes Query Sources (C6).
- **Help Modal** — the modal listing the focused pane's key bindings, reachable from the
  footer's help affordance and dismissible with one key.
- **Location line** — the de-emphasized bottom row of a Document Viewer naming where its
  document lives: Root name + relative path, full path for an orphan; for a Wikipedia
  article, `CC BY-SA 4.0 · <canonical URL> · Reformatted to markdown` (C6, amended S9a/D98).
  Ends with how the document is shown, `· Markdown` or `· Text` (D140).
- **Nav Keys** — the movement keys shared by every pane (`h`/`j`/`k`/`l` and arrows, `g`/`G`),
  vim-oriented. Provisional, like all bindings.
- **Markdown file** — a file noggin can render as markdown: extension `.md`, `.markdown`,
  `.mdown` or `.mkd`, or an extensionless `README` or `CHANGELOG` (case-insensitive) (D140).
- **Navigator** — the left sidebar: Sources Nav above, Files Nav below. Hideable.
- **Orphan** (orphaned star) — a Starred document that lies under no Root in the current
  session. Reachable from the Picker's `Starred` scope and the Files Nav's `Orphans` scope (C4).
- **Peek** — render the Files Nav's selected document in the targeted Document Viewer
  (DV1 by default, DV2 with the Targeting modifier) *without* making it that viewer's Active
  Document and without moving Focus. Transient and read-only: cannot be scrolled or
  searched, marked by an inverted header, cleared by any selection change, Focus change,
  Picker open or Refresh — restoring the Active Document at its previous scroll position.
  `<enter>` is the only promotion path. Distinct from the Picker's Preview pane (C2a).
- **Picker** — the modal fuzzy-finding UI: a Prompt, a Results list, and a Preview pane.
  Comes in modes — File Picker, Grep Picker (C2) and Wikipedia Picker (C6).
- **Preview pane** — the part of a Picker that shows the selected Result's file. Lives inside
  the Picker only; it never changes the Active Document (C2).
- **Prompt** — a Picker's input field; Results update as it is typed (C2).
- **Query Source** — a Source reached only by search: a remote service, exactly one of each
  kind, always present, never created, removed or starred. Has no Tree and no grep; its
  documents are found by search, then kept via Starred or Recent. Wikipedia is the first
  (C6). "Finder" is deliberately avoided; it means the OS file manager.
- **Readable file** — retired by D140; split into Markdown file and Text file, with the
  File Filter choosing which are offered. (Was: `.md`, `.markdown`, `.txt`, `README`,
  `LICENSE`, `CHANGELOG`, C1/C2.)
- **Recent** — documents ordered by when they were last viewed. Viewing means Activate or the
  `$EDITOR` handoff; Peek does not count. Recency only, never frecency (C4).
- **Refresh** — the on-demand re-read of the Active Root's file tree and, as of C3, every open
  Document Viewer's Active Document from disk in one action, preserving tree expansion,
  selection and each viewer's scroll position. An open Wikipedia article (C6) has nothing on
  disk to re-read: Refresh re-fetches it instead, bypassing the session cache — the one case
  where Refresh reaches the network. Otherwise manual only; noggin never watches the
  filesystem (C2a).
- **Render Mode** — how a Document Viewer shows a Markdown file: `Markdown` (rendered) or
  `Text` (raw source, line-numbered). One per viewer, sticky across the documents opened
  into it, session-only, default `Markdown`. A Text file is always shown as text without
  changing it (D140). _Avoid_: view mode, raw mode.
- **Result** — one row in a Picker: a file (File Picker), a file-and-line (Grep Picker) (C2),
  or an article title-and-description (Wikipedia Picker) — a Wikipedia Result can also carry
  a star, the same as a file's (C6).
- **Results list** — the Picker pane listing its Results, one row per Result, below the
  Prompt (C2).
- **Retune** — adjust what the conversion spike strips and re-judge the sample. Repeatable
  within the cycle; new capability may not be added to buy quality except the two named,
  costed exceptions carved out by C4a's amended retune budget (S6a): a Unicode
  approximation of maths, and a caption/alt/filename-derived image placeholder (C4a, D104).
- **Root** — a folder that bounds file visibility. Several exist per session; identified by
  exact path; may not nest inside another Root. One of the two classes of Source (C6).
- **Root name** — the editable label of a persisted Root, defaulting to its folder's base
  name; shown wherever the Root is labelled. Ad-hoc Roots have none until starred (C6).
- **Scope** — what a Picker searches: the Active Root, All Roots (C2), or Starred (C4). The
  Wikipedia Picker has none; with a Query Source active, File/Grep skip Active Root (C6).
- **Scratch root** — a throwaway Root, outside the repo, holding converted articles purely so
  they can be read in noggin's own Document Viewer at real pane widths. A judging proxy, not
  a feature (C4a).
- **Source** — anything noggin serves documents from; the entity listed in the Sources Nav.
  Two classes: Root and Query Source. Common rule: content is always fetched, never stored —
  noggin keeps pointers, stars and timestamps only (C6).
- **Sources Nav** — the upper Navigator pane: Ad-hoc Root, persisted Roots, then Query
  Sources (C6; formerly Roots Nav).
- **Spike** — a throwaway experiment whose deliverable is an answer, not a feature. C4a is
  one: its output is a go / no-go / retune-exhausted verdict plus findings.
- **Staleness indicator** — a passive UI signal that an open Document Viewer's Active Document
  or the Active Root's tree changed on disk. Reports only that *something* changed across
  those sources, and never refreshes on its own (C2a). Its own dedicated, highlighted row
  above the Footer, shown only while stale (C3) — not a Footer suffix. Checked every few
  seconds and whenever the terminal regains focus; changes to entries the tree doesn't show
  never raise it (D139).
- **Starred** — bookmarked and persisted across sessions. Applies to Roots (C1) and to
  documents (C4). Document stars are a single global collection of paths, independent of the
  Root lifecycle: un-starring a Root never touches them.
- **Targeting modifier** — the `ctrl+` prefix on Files Nav's Activate/Peek and the Picker's
  Activate that routes the action to DV2 instead of DV1, creating DV2 first if it doesn't
  already exist. Unmodified keys always target DV1. The same `ctrl+` prefix is also used by
  S5's focus-jump keys (`ctrl+o` jumps Focus to DV2) for consistency, though that is a
  distinct mechanism with no load/peek routing involved (C3; `shift+` before D55).
- **Text file** — any file that is neither a Markdown file nor binary (a NUL byte early in
  the file). Shown only as raw text. Binary files are never shown or opened (D140).
- **Wikipedia Picker** — the Picker mode that searches Wikipedia article titles; no Scope, no
  Preview; reachable from any Active Source (C6).
- **Wikipedia Source** — the first Query Source: English Wikipedia, listed last in the
  Sources Nav (C6).
