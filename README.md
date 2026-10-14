# noggin

A keyboard-driven terminal UI for reading markdown. `noggin` keeps several note
sources — personal notes, shared work notes, the docs of whatever project you're
in, and Wikipedia — open and reachable in one session: browse a folder tree,
search and read a Wikipedia article, read well-rendered markdown, search inside
a document, and hand off to your editor when you actually need to write. It
does not edit files itself.

## Install / build

There's no published binary yet. Build from source:

```sh
git clone <this repo>
cd noggin
make build      # -> bin/noggin
make install    # -> 
```

or run without building:

```sh
make run        # go run ./cmd/noggin
```

`go install` isn't a supported path yet — the module's import path (`noggin`) isn't
externally resolvable.

## Launching

```sh
noggin              # open your starred roots (see below); the first is active
noggin <dir>        # open <dir> as an ad-hoc root for this session, active on launch,
                     # alongside your starred roots
noggin <file>       # open <file> directly; your cwd becomes an ad-hoc root
```

A few launch arguments are rejected with a message in the footer rather than
failing the whole session:

- a path that doesn't exist
- a path that is *inside* an already-starred root — that starred root is opened
  instead
- a path that is the *parent* of an already-starred root — the session starts
  unchanged
- more than one path argument

If `<dir>` is already a starred root, no duplicate is created — it just becomes
the active root.

## Sources

A **Source** is anything noggin can read documents from. There are two kinds:

- **Root** — a folder you point noggin at. Plural, browsable as a tree,
  greppable, can be starred, named and removed from the session by unstarring.
- **Wikipedia** — a single, built-in **Query Source**: reached only by
  searching article titles, never browsable as a tree. It's always present,
  can't be created, removed or starred, and is the same for every session.

Exactly one Source is **active** at a time (the **Active Source**) — the one
the Files Nav, Refresh and the Picker's default scope act on. Wherever the
rest of this document used to say "Active Root", it now means "Active
Source" unless a section says otherwise (Grep, in particular, only ever
searches Roots, so it keeps saying "Active Root").

## The window

```
-------------------------------------------------
| Navigator     | Document Viewer                |
|---------------|           (DV1)                |
| Sources Nav   |                                 |
|---------------|                                 |
| Files Nav     |                                 |
|               |                                 |
|---------------------------------------------------|
| Footer                                             |
-------------------------------------------------------
```

A second Document Viewer (DV2) can appear beside the first, splitting the
right-hand region 50/50 — see [A second Document Viewer](#a-second-document-viewer)
below:

```
-------------------------------------------------------------
| Navigator     | Document Viewer  | Document Viewer          |
|---------------|      (DV1)       |        (DV2)             |
| Sources Nav   |                  |                           |
|---------------|                  |                           |
| Files Nav     |                  |                           |
|               |                  |                           |
-----------------------------------------------------------------
| (staleness row, only while something is stale)                 |
| Footer                                                          |
--------------------------------------------------------------------
```

- **Sources Nav** — the list of Sources in this session: the ad-hoc root (if
  any) first, then your starred roots, then **Wikipedia** always last. Each
  Root renders on two lines — its name in bold above its full path, unbold —
  so two same-named roots in different places stay distinguishable; moving
  up/down steps one whole entry at a time. Broken roots (path no longer
  exists) are marked and can't be activated. A starred root is marked `★`.
  Wikipedia renders the same two-line shape (`Wikipedia` / `en.wikipedia.org`)
  but is never starred or broken, and always present even with zero Roots.
  See [Renaming a Root](#renaming-a-root) and [Wikipedia](#wikipedia) below.
- **Files Nav** — lists documents from the *Active Source*, in one of the
  [Browse Modes](#browsing-tree-recent-starred): `Tree` (the file tree,
  Root only), `Recent`, or `Starred`. Switching Sources re-scopes this pane
  but leaves whatever document you're reading untouched; each Source's Tree
  expansion/selection (Roots only), Browse Mode and Browse Scope are
  remembered independently for the rest of the session, so switching back
  restores it exactly as you left it. Files with uncommitted changes are
  marked `M` (modified) or `A` (new / untracked) — see [Git status markers](#git-status-markers)
  below.
- **Document Viewer** — a document, rendered as styled markdown (headings,
  tables, syntax-highlighted code fences, lists, blockquotes, rules, emphasis)
  and wrapped to the pane width, with a de-emphasized location line on its
  bottom row (see [Document location line](#document-location-line)). Read-only:
  no cursor, no selection, no copy — just scrolling. A session always has one
  Document Viewer (DV1) and may have a second (DV2) beside it — see
  [A second Document Viewer](#a-second-document-viewer).
- **Footer** — key hints for whatever pane has focus, plus transient messages
  (a rejected launch path, a refused action). A message clears on your next
  keypress. Hints truncate gracefully on a narrow terminal, always keeping
  `help` and `quit`.

Only one pane has focus at a time; most keys route to it. Every pane shows
the key that jumps focus to it, right-aligned on its title row (e.g.
`Files` ... `[f]`, DV2's `[^o]`) — dropped if the pane is too narrow to fit
both. See [Focus](#focus) below. The focused pane is drawn with a thick border
and an accent-coloured title bar, so focus is visible at a glance — see
[Focus](#focus).

### Appearance

noggin has its own colour scheme and does not inherit your terminal's theme: it
paints every cell — background included — from one built-in palette, so it looks
the same in any terminal and its text is always legible against its own
background. It uses 24-bit colour; a terminal that only does 256 colours gets
the nearest match. There is no light/dark detection and no setting for it. (A
terminal that honours `NO_COLOR` still gets no colour.)

## Keybindings

Bindings are provisional (cheap to retune) but consistent across panes: movement
is vim-style (`j`/`k`/arrows, plus `g`/`G` for top/bottom where it applies).

### Global (any non-modal pane)

| Key | Action |
|---|---|
| `?` | open the Help modal (full keybinding list for the focused pane) |
| `ctrl+n` | toggle the Navigator (hide/restore the whole left sidebar) |
| `r` | focus the Sources Nav (reveals the Navigator if hidden) |
| `f` | focus the Files Nav (reveals the Navigator if hidden) |
| `o` | focus DV1 |
| `ctrl+o` | focus DV2 (no-op if DV2 doesn't exist — a bare focus request never creates it) |
| `ctrl+x` | close DV2, wherever focus is (no-op if DV2 doesn't exist); focus moves to DV1 if DV2 held it |
| `R` | refresh the Active Root's file tree and every open Document Viewer's document from disk (re-fetches an open Wikipedia article instead) |
| `e` | edit the active document in `$EDITOR` (refused on a Wikipedia article) |
| `m` | toggle the viewer between Markdown and Text rendering (see [Render Mode](#render-mode)) |
| `/` | start find in the active document |
| `n` / `N` | next / previous find match (once a search has matches) |
| `space` | open the Picker (fuzzy find files, grep across files, or search Wikipedia — see below) |
| `ctrl+space` | reopen the Picker to its last non-empty query, mode and Scope (opens normally if there is none) |
| `q` / `ctrl+c` | quit |
| `esc` | context-dependent cancel (dismiss find, discard a rename, close Help, close Picker) |

There is no pane-cycling key: every pane is reached by a direct focus key
(`r`/`f`/`o`/`ctrl+o`) instead — see [Focus](#focus).

`e`, `m`, `/`, `n` and `N` are reachable from any pane, not just a Document
Viewer's, but always act on DV1's document — unless focus is currently on
DV2, in which case they act on DV2's.

### Sources Nav

| Key | Action |
|---|---|
| `j`/`k` or `↓`/`↑` | move selection |
| `enter` | activate the selected Source (a Root, or Wikipedia) |
| `s` | star / unstar the selected root (no-op with a footer message on Wikipedia) |
| `c` | rename the selected root inline (see [Renaming a Root](#renaming-a-root); refused on Wikipedia and on an unstarred ad-hoc root) |

While a rename is in progress, every other key types into the name instead of
acting on its usual binding: `enter` confirms, `esc` cancels.

### Files Nav

| Key | Action |
|---|---|
| `j`/`k` or `↓`/`↑` | move selection |
| `l`/`→` | expand the selected folder (`Tree` mode) |
| `h`/`←` | collapse the selected folder (`Tree` mode) |
| `enter` | open the selected file into DV1 and move focus there (fetches and opens an article for a Wikipedia row) |
| `ctrl+enter` (fallback `ctrl+l`) | open the selected file into DV2 instead, creating it if needed, and move focus there |
| `p` | peek at the selected row in DV1, without opening it or moving focus (press `p` again, or move the selection, to return) |
| `ctrl+p` | peek at the selected row in DV2 instead, creating it if needed, without moving focus |
| `s` | star / unstar the selected document or article |
| `tab` | cycle Browse Mode: `Tree` → `Recent` → `Starred` → `Tree` … (Wikipedia active: `Recent` ↔ `Starred` only, no `Tree`) |
| `shift+tab` | cycle the current mode's Browse Scope (no-op in `Tree`) |

See [Browsing: Tree, Recent, Starred](#browsing-tree-recent-starred).

### Document Viewer (DV1 or DV2, whichever has focus)

Each viewer acts on its own document only — scrolling, finding, starring,
editing and history in DV1 never touch DV2, and vice versa.

| Key | Action |
|---|---|
| `j`/`k` or `↓`/`↑` | scroll one line |
| `ctrl+u` / `ctrl+d` | half-page up / down |
| `pgup`/`ctrl+b` , `pgdown`/`ctrl+f` | half-page up / down (same distance as `ctrl+u`/`ctrl+d`) |
| `g`/`home` , `G`/`end` | jump to top / bottom |
| `s` | star / unstar this viewer's document |
| `[` / `]` | step back / forward through this viewer's own Document History |

See [Document history](#document-history).

### While finding (`/`)

| Key | Action |
|---|---|
| *(typing)* | updates matches live |
| `enter` | commit the search — matches stay live for `n`/`N` |
| `esc` | dismiss the search entirely |

### While the Picker is open (`space`)

Every printable key types into the Prompt — there's no separate typing mode to
leave. Everything else is on a non-printable key:

| Key | Action |
|---|---|
| *(typing)* | updates the Results list live (once 2+ characters are typed) |
| `↑` / `↓` | move the Results selection; the Preview follows it (File/Grep only) |
| `tab` | switch mode: File → Grep → Wikipedia → File … (the typed query carries over) |
| `shift+tab` | switch Scope (no-op in Wikipedia mode, which has none — see below) |
| `alt+h` | toggle hiding dotfiles from the Results, for the rest of this session (File/Grep only) |
| `alt+s` | star / unstar the selected Result (file or article) |
| `enter` | activate the selected Result into DV1 (fetches an article, then renders it) |
| `ctrl+enter` (fallback `ctrl+l`) | activate the selected Result into DV2 instead, creating it if needed |
| `esc` | dismiss the Picker — nothing about the session changes |

See [Finding across files (the Picker)](#finding-across-files-the-picker) and
[Wikipedia](#wikipedia).

## Focus

Exactly one pane has focus at a time, and every pane is reached by a direct
key from anywhere non-modal: `r` (Sources Nav), `f` (Files Nav), `o` (DV1),
`ctrl+o` (DV2 — a no-op if DV2 doesn't exist). `r`/`f` reveal the Navigator
first if it's hidden; `o`/`ctrl+o` don't, since DV1/DV2 are never inside it.
There's no Tab-cycling. Each pane's own focus key is shown right-aligned on
its title row (`Files` ... `[f]`, DV2's `[^o]`), dropped if the pane is too
narrow for both title and badge.

The focused pane is drawn with a thick border and its title row becomes a
full-width accent-coloured bar; unfocused panes keep the thin rounded border
and a plain bold title. (A Document Viewer that is Peeking keeps its inverted
Peek header even while focused.)

Unfocused panes also recede: a Sources or Files Nav's selected row becomes a dim
grey bar (instead of a bright reverse-video one) and its other text is muted,
and an unfocused Document Viewer shows its document in a desaturated, darker
version of the colour scheme. Focusing a pane restores full colour.

Once a Document Viewer holds focus, its local commands — Find, Edit, star,
scrolling, history — act on that viewer's own document only.

## A second Document Viewer

A session always has one Document Viewer, DV1, which takes the full width of
the Document region alone. A second, DV2, can be created beside it — the
region then splits 50/50, DV1 on the left — and closed again, without ever
affecting DV1's own document, scroll position, find state or history:

- From the Files Nav, `ctrl+enter` (fallback `ctrl+l`, for terminals that
  can't distinguish it from plain `enter`) opens the selected file into DV2,
  creating it first if needed, and moves focus there — the DV2 counterpart
  of `enter`. `ctrl+p` peeks it into DV2 instead, the counterpart of `p`:
  focus doesn't move.
- From the Picker, the same `ctrl+enter`/`ctrl+l` activates the selected
  Result into DV2 instead of DV1.
- `ctrl+x`, from any non-modal pane's focus (including DV1 or DV2
  themselves), closes DV2. A no-op if it doesn't exist. If DV2 held focus,
  focus moves to DV1; from anywhere else, focus is unchanged.

DV1 and DV2 are fully independent: separate Active Document, Peek, Find and
scroll position each, and each walks its own [Document History](#document-history) —
this applies to a Wikipedia article exactly as to a file. DV2 is
session-only — always absent at launch, never persisted.

`R` (refresh) re-reads the Active Root's tree and both viewers' documents
(whichever exist) in one action, regardless of which pane has focus; it
re-fetches instead for a viewer showing a Wikipedia article. The
[staleness indicator](#staleness-indicator) also now spans both viewers, not
just one.

## Renaming a Root

Every persisted Root has a **name**, shown in bold on its Sources Nav row,
its Picker result rows, and cross-Source Files Nav rows; it defaults to the
folder's own base name until you rename it. Press `c` on a Root row in the
Sources Nav to edit its name in place: `enter` confirms, `esc` cancels, and
clearing the name entirely reverts it to the default. The name persists
across sessions alongside the root's starred state.

An unstarred ad-hoc root can't be renamed yet — `c` reports "star this root
before renaming it" in the Footer; starring it first gives it the default
name to rename from. The Wikipedia row can't be renamed at all. Names aren't
required to be unique — the Sources Nav's path line (and, for two same-named
roots, its left-truncated tail) tells them apart regardless.

## Starring roots and documents

Roots, files and Wikipedia articles can each be starred independently of one
another — un-starring a root never touches the star on any document under
it, and vice versa.

**Roots:** press `s` in the Sources Nav to star or unstar the selected root
(a no-op on the Wikipedia row, refused with a footer message). Starred roots
are what a plain `noggin` (no argument) opens next time, marked `★` in the
Sources Nav. They're recorded, one absolute path per line (optionally
followed by a tab and its Name, once renamed — see
[Renaming a Root](#renaming-a-root)), in a plain text file at:

- `$XDG_DATA_HOME/noggin/roots` (or the platform equivalent, e.g.
  `~/.local/share/noggin/roots` on Linux, `~/Library/Application Support/noggin/roots`
  on macOS)

The file is deliberately hand-editable — `.gitignore`-style comments (`#`) and
blank lines are fine, so you can fix a moved root by editing the path directly.
A missing or unreadable file is treated as "no starred roots" rather than an
error, and stale paths (root no longer exists) are kept, not silently pruned —
they just show up broken in the Sources Nav.

**Documents and articles:** press `s` to star or unstar a document or
Wikipedia article from the Files Nav (the selected row), from a focused
Document Viewer (that viewer's own document — DV1 and DV2 can have different
starred documents open at once), or `alt+s` in the Picker (the selected
Result, file or article). A star shows as `★` on Files Nav rows (in every
Browse Mode except `Starred`, where membership already says it), to the left
of the document's title in a Document Viewer, and on Picker rows in every
mode. A starred document or article survives un-starring its root, moving to
a different root, or that root simply not being part of a later session —
it's tracked by absolute path (or, for an article, its canonical URL), not
by root. A **broken starred document** (path no longer on disk) still shows
up starred, marked `(missing)`, but can't be opened; it's reachable from the
Files Nav's `Starred` mode and the Picker's `Starred` scope even when no open
root covers it (see [Browsing](#browsing-tree-recent-starred)). A starred
Wikipedia article can never go "missing" the same way — see
[Wikipedia](#wikipedia) for what happens if it can no longer be fetched.

Document/article stars, along with each one's last-viewed time (used by
[Recent](#browsing-tree-recent-starred)), live in a second plain file
alongside the roots file:

- `$XDG_DATA_HOME/noggin/documents` (same directory as `roots` above)

Each line is tab-separated, in one of two shapes: the plain-file shape
`<0 or 1>\t<RFC3339Nano timestamp or ->\t<absolute path>`, or, for a
Wikipedia article, `<0 or 1>\t<RFC3339Nano timestamp or ->\t<file or
article>\t<title>\t<key>` — `<key>` is the file's absolute path or the
article's canonical URL, matching the pointer noggin actually keeps (never
the article's content). A starred entry is never evicted, only its
timestamp is cleared, if the store grows past its internal cap — files and
articles share one cap and one eviction rule.

Note: both files live under noggin's *data* directory, not its config
directory (`$XDG_CONFIG_HOME`) — a config-directory `roots` file from an
older noggin is not read or migrated.

## What gets shown

noggin classifies every file it meets:

- **Markdown file**: extensions `.md`, `.markdown`, `.mdown`, `.mkd`, or the
  extensionless names `README` and `CHANGELOG` (all case-insensitive).
  `.txt` and `LICENSE` are *not* Markdown.
- **Text file**: any other file that is not binary.
- **Binary file**: a NUL byte in the first ~8KB (the test ripgrep and git use).

For now the Tree and the Pickers list **Markdown files only**; a later
File Filter will widen them to Text files. Whatever is listed, noggin refuses
to open a binary file or one over ~10MB, on every route — the Files Nav, the
Picker, `noggin <file>`, Recent/Starred, history walks and Refresh — with a
Footer message such as `cannot open app.bin: binary file` or `cannot open
dump.md: too large`.

Folders are always shown, even empty ones or ones containing no Markdown
(unless the folder itself is ignored — see below).

`.gitignore` is respected while browsing, with no toggle to turn it off. It
covers the common cases (per-directory `.gitignore` files, `#` comments, `!`
negation, directory-only and anchored patterns, `*`/`?`/`[...]`/`**` globs) but
is a focused subset, not full git semantics — it doesn't consult global excludes
or `.git/info/exclude`. `.git` itself is always excluded. Dotfiles and hidden
directories otherwise *are* shown (e.g. `.github/`), since they often hold real
notes.

## Browsing: Tree, Recent, Starred

The Files Nav has three Browse Modes, cycled with `tab` (Files Nav focused):
`Tree` (the file tree), `Recent`, and `Starred`. A status row below the
pane's title always names the current mode and scope, e.g. `Tree · -`,
`Recent · Active Source`, `Starred · All Starred` — costing the list one row
of height in every mode.

With a **Root** active, all three modes are reachable, defaulting to `Tree`.
With **Wikipedia** active, there is no tree to walk: `Tree` drops out of the
cycle entirely and `tab` toggles directly between `Recent` and `Starred`
instead, landing on `Starred` the first time a session activates Wikipedia.

`shift+tab` cycles the current mode's Browse Scope (a no-op in `Tree`,
which has none):

- **Recent** — `Active Source` ↔ `All Sources`. Lists documents by when
  they were last Activated (opened into a viewer, or fetched by Activating
  a Wikipedia article) or sent to `$EDITOR`; peeking doesn't count.
  `Active Source` shows at most 100 documents; `All Sources` merges every
  Source's list by timestamp with no further cap, labelling a Wikipedia
  article's row `Wikipedia` the way a file's row is labelled by its Root.
- **Starred** — `Active Source` → `All Starred` → `Orphans`. `All Starred`
  is every starred document and article in the store, including ones under
  no current Root, labelling an article's row `Wikipedia`; `Orphans` is
  exactly the starred *documents* under no session Root at all (a Root-only
  idea — articles are never under a Root, so they never appear here, and
  `Orphans` itself drops out of the cycle while Wikipedia is active, leaving
  just `Active Source` ↔ `All Starred`).

`Recent` and `Starred` are flat lists, not trees: each row shows the
document's name on the left and, right-aligned, the covering Root's name (or,
for an orphan, its left-truncated parent path, or, for an article,
`Wikipedia`). `Recent` rows carry the usual `★` star marker; `Starred` rows
omit it (mode already implies it) and instead mark a broken starred document
with `(missing)`. Any mode/scope combination with nothing to show says so
directly (e.g. "No starred documents outside a session Root.").

The current mode is remembered per Source, and the scope is remembered per
(Source, mode) pair, for the rest of the session — returning to a Source
restores what you last had there. Every session starts every Root on `Tree`
and Wikipedia on `Starred`.

## Document History

Each Document Viewer (DV1 and DV2 independently) keeps its own back/forward
Document History of the documents opened into it, walked with `[`/`]` while
that viewer has focus — a Wikipedia article is a document like any other
here. It's browser-style: activating a new document after going back
discards the forward entries, and the same document can appear more than
once but never twice in a row. Going back or forward restores that
document's own most recently used scroll position in that viewer (remembered
per document, not per history entry).

History can walk into documents under other Roots, or a mix of Roots and
Wikipedia, without ever changing the Active Source. The `$EDITOR` handoff and
`R` (refresh) do not add history entries — both re-render the same document
in place instead. History and its scroll memory are session-only for both
viewers, dropped entirely when DV2 is closed or noggin exits.

## Document location line

Each Document Viewer's bottom row shows, de-emphasized, where its document
lives — costing the content above it one row of height, and following a
Peeked document while the Peek lasts:

- a file under a Root: `<Root name> · <path relative to the Root>`
- an orphan file (under no session Root): its full path
- a Wikipedia article: `CC BY-SA 4.0 · <canonical URL> · Reformatted to
  markdown` — Wikipedia's own reuse licence, the article's own URL as
  attribution, and a note that the HTML has been converted to markdown

Every form ends with the document's [Render Mode](#render-mode): `· Markdown`
or `· Text`. On a narrow pane it is the path that is truncated, not the mode.

The row is absent entirely when the viewer has no document open (or none to
Peek).

## Render Mode

Each Document Viewer shows its document in one of two Render Modes, toggled
with `m`:

- **Markdown** (the default): the document rendered as styled markdown.
- **Text**: the raw source exactly as written — frontmatter included — with
  tabs expanded, long lines soft-wrapped to the pane width, a dim line-number
  gutter (blank on continuation rows) and no colour. An unfocused pane shows a
  muted variant.

`m` is reachable from any pane and, like `e` and `/`, acts on DV1's viewer
unless focus is on DV2, in which case it acts on DV2's. The mode belongs to the
viewer and is sticky: it carries across every document opened, Peeked or
walked to with `[`/`]` in that viewer, for the rest of the session. DV1 and DV2
each have their own, and a new DV2 starts on Markdown. Toggling resets any find
in progress.

A Text file (anything that is not a [Markdown file](#what-gets-shown)) is always
shown as text without changing the viewer's mode, so `m` on one is refused with
a Footer message. A Wikipedia article in Text mode shows its converted
markdown.

**Find in Text mode** is exact. `/` matches against the source lines (smart-case
as usual), so a match can span a soft-wrapped row, and it finds what rendering
hides — a link's URL, the `**` around bold. Only the matched span of the current
match is highlighted, not the whole line; `n`/`N` and wrap-around work as in
Markdown mode. Activating a Grep Result into a viewer in Text mode lands on
ripgrep's exact line and column, as the current match of a committed find.
Find in Markdown mode is unchanged: it searches the rendered text and marks the
whole line.

## Editing

Press `e` to hand off to your editor: `$EDITOR` if set, otherwise `vi`. `e` is
reachable from any pane, and always acts on DV1's document unless focus is
currently on DV2, in which case it acts on DV2's. noggin releases the
terminal for the editor and resumes when it exits, re-reading the file from
disk so your edits show up immediately — there's no live file-watching, this
is the one point where it re-reads. With no document open in the targeted
viewer, `e` just reports that instead of doing nothing. A Wikipedia article
has nothing on disk to hand `$EDITOR`, so `e` is refused with a Footer
message ("cannot edit a Wikipedia article") instead.

## Help

Press `?` anywhere for a modal listing every binding relevant to the pane you're
in (a superset of the footer's hints — it also lists things like cancel/next
match/previous match for the Document Viewer). Press `?` or `esc` again to close
it; whatever had focus keeps it.

## Frontmatter

YAML frontmatter at the top of a document is shown as a card above it: one row
per key, in file order, values wrapped beside their keys and lists
comma-joined. The frontmatter goes between `---` fences (closed by `---` or `...`)
and the opening fence must be the file's first line.

Where the card can't show the frontmatter faithfully, it is shown verbatim as a
syntax-highlighted code block instead. That happens with nested values (a map, or
a list of maps or lists), anchors and aliases, invalid YAML, TOML frontmatter
(`+++` fences), or a pane too narrow for a card. An empty block shows as nothing.

## Keeping things in sync

noggin never watches the filesystem — everything on screen is a snapshot from
when the root, document or article was last read.

Press `R` from any pane to refresh: it re-reads the Active Root's file tree
and both DV1's and DV2's documents (whichever exist) from disk in one pass —
a viewer showing a Wikipedia article re-fetches it from Wikipedia instead,
bypassing the session cache, the one time Refresh reaches the network. Files
Nav expansion and selection are preserved (selection moves to the nearest
surviving ancestor if its exact node is gone); each viewer's own scroll
position is preserved independently, clamped to the end if its document
shrank past it; an in-progress or committed find may be lost in either
viewer. If a viewer's document no longer exists or can't be read (or its
article can't be re-fetched), that viewer's document is left untouched and a
Footer message reports the failure. Nothing outside the Active Root's tree
and the two viewers' documents is touched — other roots, other documents,
and Picker results are unaffected. Refresh also recomputes the git status
markers below, and doesn't add a [Document History](#document-history) entry.

### Git status markers

Files with uncommitted changes are marked in the Files Nav: `M` for modified,
`A` for new/untracked. Only those two states are shown — staged vs. unstaged
isn't distinguished, and deleted or renamed files aren't marked. Status is
resolved against the git repository containing the Active Root (its root may
sit above the Root itself), not the Root path directly, so a Root that is a
subdirectory of a repo still shows correct markers. A Root outside any git
repository, or a machine without `git` on `PATH`, shows no markers, silently.
Markers are computed when a root is opened and recomputed on every `R`
refresh — they are never kept live in between. Wikipedia articles never
carry a git status marker.

### Staleness indicator

Every few seconds, noggin quietly checks whether the Active Root's tree, DV1's
document, or DV2's document (if it exists) has changed on disk since it was
last read. If any has, a dedicated highlighted row appears just above the
Footer reading "something changed on disk; press R to refresh" — it occupies
no space at all while nothing is stale. It never says *what* changed, and
never refreshes or re-reads anything on your behalf — pressing `R` is the
only way to act on it. It's one combined signal across tree/DV1/DV2, not
per-source; it clears the next time you refresh and stays cleared until
another change is detected. A Wikipedia article is never tracked by this
indicator — there is no local copy of it to go stale, and `R` always
re-fetches it fresh regardless.

## Finding across files (the Picker)

Press `space` from any pane, including with the Navigator hidden, to open the
Picker: a modal fuzzy-finder over the files in scope, or a Wikipedia title
search. It always opens the same way — empty Prompt, and File mode/Active
Root scope, unless Wikipedia is the Active Source, in which case it opens
straight into Wikipedia mode instead. `space` never remembers anything: the
one exception is `ctrl+space`, which reopens the Picker to the query, mode and
Scope of the last session dismissed with something typed (by `esc` or by
activating a Result), and searches it straight away. A session dismissed with
an empty Prompt doesn't replace it, and with nothing saved `ctrl+space` opens
normally. It lasts for the current session only — it isn't saved between runs.

The Picker has three modes, switched in place with `tab` without losing what
you've typed:

- **File** (the default with a Root active) — fuzzy-matches your query
  against file *names* only, ranked by match strength, with the matched
  characters highlighted in each row. Each row also shows the file's Root
  name and its path relative to that Root, so two same-named files in
  different Roots stay distinguishable.
- **Grep** — searches for your query as literal text (no regex) across file
  *contents*, case-insensitively unless your query itself has an uppercase
  letter. One row per matching line, showing the file, its line number and
  the line's text, with the matched text highlighted. Grep never searches
  Wikipedia, in any scope.
- **Wikipedia** (the default with Wikipedia active) — see
  [Wikipedia](#wikipedia) below.

File and Grep each search one of three Scopes, cycled in place with
`shift+tab`: **Active Root**, **All Roots**, or **Starred** (every starred
document, including orphans, plus — File mode only — starred Wikipedia
articles matched on title and labelled `Wikipedia`; `All Roots` and Grep in
every scope never include an article). With Wikipedia the Active Source,
File and Grep have no Active Root to offer, so their Scope cycles only
`All Roots` ↔ `Starred`. The title above the Prompt names both mode and
scope, e.g. `File · Active Root` or `Grep · Starred`. All Roots silently
skips any Broken root.

A query shorter than 2 characters doesn't search at all ("type at least 2
characters"); a query that matches nothing says "no results"; a File/Grep
result set larger than 200 says so ("first 200 results (truncated)") instead
of implying that's everything. The Preview pane shows the selected Result's
file as raw, unrendered text — the File Picker from the top, the Grep Picker
scrolled to and highlighting the matched line — and you can scroll it
without affecting anything else. `enter` activates the selected Result: it
becomes the Active Document (landing on the matched line for a Grep Result,
with next/previous `n`/`N` traversal working from it, same as a typed `/`
search), without changing the Active Root or touching the Files Nav tree's
expansion or selection. `esc` dismisses the Picker with no side effects at
all — same Active Document, same Active Root, same scroll positions.

Only [Markdown files](#what-gets-shown) are ever searched or shown by File
and Grep.

### Requires `ripgrep`

File and Grep are powered by [`ripgrep`](https://github.com/BurntSushi/ripgrep)
(the `rg` binary on your `PATH`) for both listing file names and searching
text. Without it, noggin still launches and reads documents normally — only
File/Grep search is affected: opening the Picker in either mode shows "Picker
unavailable: ripgrep (rg) not found on PATH" instead of a search. Wikipedia
mode needs no `rg` and is unaffected.

Because `rg` supplies the file list, its own ignore rules apply, not
noggin's Files Nav ignore rules: `.gitignore`/`.ignore` files are respected
(consistent with the Files Nav). Dotfiles and dot-directories (e.g. `.github/`)
are included by default in both modes, matching what the Files Nav tree
already shows — `.git/` itself stays excluded unconditionally, in either
pane. `alt+h` toggles hiding them again, in both modes at once; the title
above the Prompt gains a ` · Hidden` suffix while the toggle is on. It's
session-only — relaunching noggin resets it to showing dotfiles. `alt+h` has
no effect in Wikipedia mode.

## Wikipedia

Wikipedia is always listed as the last row of the Sources Nav (`Wikipedia` /
`en.wikipedia.org`), even with zero Roots starred. Press `enter` on it to
make it the Active Source; with it active, the Files Nav's `Tree` mode drops
out (see [Browsing](#browsing-tree-recent-starred)) and the Picker opens
straight into Wikipedia mode.

**Searching.** Open the Picker (`space`) and switch to Wikipedia mode with
`tab` (or just open it with Wikipedia already active). Typing 2 or more
characters searches Wikipedia's own article-title index — like Wikipedia's
own search box, not a full-text search — after a short pause in typing
(a little longer than File/Grep's, since each search is a network call).
Results show each article's title and, when Wikipedia supplies one, a short
description, with a `★` marker for a starred article. There is no Preview
for Wikipedia mode — the Preview pane shows a short placeholder instead
("No preview — Activate to open the article"). A failed search (offline, a
Wikipedia service error) is reported on the Picker's status line and leaves
the previous Results in place rather than clearing them. `alt+s` stars or
unstars the selected Result exactly as it does a file Result.

**Opening an article.** `enter` fetches and opens the selected Result (or a
Wikipedia row from the Files Nav's `Recent`/`Starred` lists, or a starred
article surfaced in the File Picker's `Starred` scope) into DV1;
`ctrl+enter`/`ctrl+l` opens it into DV2 instead; `p`/`ctrl+p` peek it.
Redirects are followed — the document noggin opens and remembers is the
article's *canonical* title and URL. While the fetch is in flight the UI
stays fully responsive: the targeted viewer's title reads `Fetching:
<title>` and its content area shows a matching placeholder, then renders the
converted article once it lands. A failed fetch (offline, the article no
longer exists, a timeout) reports "cannot fetch \<title\>: \<error\>" in the
Footer and leaves whatever the viewer was already showing untouched.
Fetched articles are cached only for the running session, keyed by their
canonical URL; `R` always re-fetches, bypassing that cache (see
[Keeping things in sync](#keeping-things-in-sync)).

**Stars and Recent.** A Wikipedia article can be starred and un-starred from
the Files Nav, a focused Document Viewer, or the Picker, exactly like a
file, and Activating it records it in `Recent` (Peeking does not) — see
[Starring roots and documents](#starring-roots-and-documents) and
[Browsing](#browsing-tree-recent-starred). Nothing about an article's
*content* is ever persisted — only its canonical URL, title, starred flag
and last-viewed time, the same shape a file record keeps for its path.

**Attribution.** Every open article's [location line](#document-location-line)
reads `CC BY-SA 4.0 · <canonical URL> · Reformatted to markdown`, satisfying
Wikipedia's reuse terms for content noggin has converted from HTML to
markdown.

**Limits.** noggin only reads English Wikipedia, never follows links inside
an article body (they render as plain text), never edits or previews an
article for editing (`e` is refused — see [Editing](#editing)), and never
shows a staleness warning for one (see
[Staleness indicator](#staleness-indicator)). Grep and the File Picker's
`All Roots` scope never include Wikipedia — see
[Finding across files](#finding-across-files-the-picker).

## Scope

noggin does not create, rename, move or delete files or folders, and it isn't an
editor — writing happens by handing off to `$EDITOR`. It also does not follow
links inside a Wikipedia article, edit or preview-for-editing a Wikipedia
article, or read anything but English Wikipedia.
