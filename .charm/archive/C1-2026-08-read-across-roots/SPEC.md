# Spec — C1 read-across-roots

Status: agreed 2026-07-30

`noggin` is a keyboard-driven TUI for reading markdown. This cycle delivers the first
usable slice: several note sources open at once, a tree to browse the active one, a
well-rendered document, find-in-document, and a clean handoff to `$EDITOR`. It is not an
editor and does not modify files.

## Principles (apply to every deliverable)

- **Go with the grain.** Where intent and library behavior differ, try the library's way
  first. Where a spec'd nicety would require authoring a subsystem, drop the nicety and
  raise it for triage instead. Named library behaviors below are illustrative, not binding.
- **Provisional keymap.** Specific bindings are expected to change through dogfooding.
  Bindings named here are indicative; deliverables are defined by their *actions*. Keys must
  be cheap to change at build time. Not user-configurable this cycle.
- **Keyboard-only.** Every action is reachable by keyboard. No mouse support required.
- **Silence is not an error, but confusion is.** Failed or rejected actions say so briefly;
  they never fail invisibly and never crash the session.

## Deliverables

### Roots

- **S1 — Launch contract.** A single optional path argument resolves the session's roots:
  - `noggin` → starred roots only; the first is active.
  - `noggin <dir>` → `<dir>` is an **ad-hoc root**, listed first and active on launch, alongside
    the starred roots.
  - `noggin <file>` → the **cwd** is the ad-hoc root and `<file>` is opened as the active
    document, whether or not it falls inside any root.
  - `noggin <dir>` where `<dir>` **is** a starred root → no duplicate is created; that starred
    root simply becomes active.
  - Rejected, each with a message and a session that still starts: a path that does not
    exist; a path that is a subdirectory of an existing starred root (that covering root
    becomes active instead); a path that is a parent of an existing starred root; more than
    one path argument.
  - Roots are unique by exact path. Two roots may share a base name.
- **S2 — Starred roots persist.** Starring a root records it; it is present in every later
  session without arguments. Stored in one plain, **hand-editable** file under the user's
  config directory, so a moved root can be repaired in an editor. A missing or unparseable
  file yields an empty root list rather than a failure. Paths that no longer exist are kept,
  not pruned.
- **S3 — Roots Nav.** A list pane showing the session's roots: ad-hoc root first, then
  starred roots. Each entry shows the folder's base name prominently plus its parent path,
  left-truncated (`…/work/admin/`), so same-named roots are distinguishable. The active root
  is highlighted; starred roots are marked; roots whose path is missing are marked broken and
  cannot be activated. Actions: move selection, activate the selected root, toggle its star.
- **S4 — Switching roots does not disturb reading.** Activating a different root re-roots the
  Files Nav and leaves the open document untouched — a document from one source stays readable
  while another source is browsed. Tree expansion and selection reset on switch this cycle.

### Browsing

- **S5 — File tree.** The active root rendered as a navigable folder tree, showing folders and
  the files noggin considers readable. Opens with top-level entries only, collapsed. Actions:
  move selection, expand/collapse a folder, activate a file — which makes it the active
  document and moves focus to the viewer. No file or folder is ever created, renamed, moved or
  deleted this cycle.
- **S6 — What's visible.** Readable files are markdown-ish only: `.md`, `.markdown`, `.txt`,
  and conventional extensionless documents (`README`, `LICENSE`, `CHANGELOG`). The set is
  defined in one place and trivially adjustable. `.gitignore` is respected by default with no
  toggle (a minimal built-in exclusion of `.git` if the file walker does not do this natively).
  Dotfiles and hidden directories **are** shown — `.charm/`, `.github/` and the like hold real
  notes. Folders containing nothing readable are still shown.

### Reading

- **S7 — Document Viewer.** The active document rendered as styled markdown, wrapped to pane
  width, at the quality bar Glow sets (Glamour): headings, tables, code fences with syntax
  highlighting, nested and task lists, blockquotes, rules, emphasis. YAML frontmatter is
  **hidden**, not rendered. Navigation is **viewport scrolling only** — line, half-page, page,
  top, bottom. There is no cursor, no selection and no copy. With no document open, the pane
  says so.
- **S8 — Find in document.** A `/`-style incremental search over the open document with
  next/previous match traversal that wraps, and dismissal that leaves the view where it is.
  Case-insensitive unless the query contains uppercase. A match is **located**: the view
  scrolls so the match's line is on screen and that line is indicated. Character-exact
  highlighting is desirable but not required — line-level indication is an acceptable outcome.
- **S9 — Navigator toggle.** One key hides and restores the whole left sidebar, giving the
  document the full width.

### Handoff

- **S10 — Edit.** One key hands the active document to `$EDITOR` (falling back to `vi`):
  noggin releases the terminal, the editor owns it, and on exit noggin resumes and re-renders
  the document from disk so edits are immediately visible. No file watching. With no document
  open, it reports that rather than doing nothing. Editing alongside a live noggin (tmux
  panes) is not this cycle.

### Orientation

- **S11 — Footer.** A persistent bottom bar showing the key hints relevant to the focused
  pane, always including quit. Hints truncate gracefully when they do not fit, advertising the
  help affordance. It is also where transient messages appear (rejected launch argument,
  failed action), clearing on the next keypress.
- **S12 — Help.** A modal listing the available keys for the current context, reachable from
  the footer affordance and dismissible with one key. This is the discoverability backstop:
  noggin must be usable without documentation.

## Out of scope

Each exclusion below is a roadmap entry, not a promise inside this cycle.

- Filename find and grep, and any search beyond the open document — **C2**.
- The finder's preview pane — **C2**.
- More than one document on screen; viewer targeting modifiers; follow mode — **C3**.
- Document history (back/forward), starred documents, recent, browse-by-starred,
  browse-by-recent, wiki-links, URL opening — **C4**.
- Tags and the tag browser; the outline pane; creating, renaming, moving, copying or deleting
  files and folders — **C5**.
- Viewing source code and other non-markdown text files — *Unscheduled*.
- A `nogginignore` file; hidden-file and ignored-file toggles — *Unscheduled*.
- Editing in a tmux pane beside a live noggin — *Unscheduled*.
- Remembering tree expansion and selection across root switches — *Unscheduled*
  (accepted friction).
- Multiple path arguments at launch; adding or removing roots in-session — *Unscheduled*.
- User-configurable keybindings; modifier-multiplied navigation acceleration (5×/15×/50×) —
  *Unscheduled*.
- Cursor, select mode and copy in the viewer — *Unscheduled*.
- Mouse support, images, non-text files, file watching, sync, encryption.

## Acceptance (cycle-level)

- **The dogfood week.** For one week, noggin is how notes get read: launch in a project repo,
  jump to personal notes and back, walk a tree to a document, read it, find a phrase in it,
  open it in nvim and return to see the change — without reaching for Glow or Obsidian.
- `noggin .` inside a project shows the cwd first and the starred roots after it; `noggin`
  with no argument shows the starred roots; a first run with neither explains what to do next.
- Starring a root in one session, quitting, and running bare `noggin` shows that root.
- Activating a second root leaves the open document rendered and unchanged.
- A phrase near the end of a long document is found and brought on screen by search.
- Editing the open document and quitting the editor returns to noggin showing the new content.
- Every one of the above is done from the keyboard alone, and every key used is discoverable
  from the footer or the help modal.

## If the cycle overruns

Pre-agreed cut order, most expendable first: **S12 help modal → S9 navigator toggle → S8
find-in-document.** Everything above those is the cycle.

## Notes

- Frontmatter is hidden by pre-stripping, as Glow does; see
  `findings/frontmatter-rendering.md`. Rendering it unstripped is not a neutral fallback —
  CommonMark turns the closing `---` into a heading underline.
- `NOTES.md` has `j` mapped to *up* and `k` to *down*, inverted from vim convention. Read as a
  typo; bindings follow vim.
- Open judgment questions parked in `.charm/DECISIONS.md`: whether noggin ever becomes a
  general text-file viewer (D9), and the eventual tag notation (D10).
