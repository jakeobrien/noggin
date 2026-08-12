# noggin

A keyboard-driven terminal UI for reading markdown. `noggin` keeps several note
sources — personal notes, shared work notes, the docs of whatever project you're
in — open and reachable in one session: browse a folder tree, read well-rendered
markdown, search inside a document, and hand off to your editor when you actually
need to write. It does not edit files itself.

## Install / build

There's no published binary yet. Build from source:

```sh
git clone <this repo>
cd noggin
make build      # -> bin/noggin
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

## The window

```
-------------------------------------------------
| Navigator     | Document Viewer                |
|---------------|           (DV1)                |
| Roots Nav     |                                 |
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
| Roots Nav     |                  |                           |
|---------------|                  |                           |
| Files Nav     |                  |                           |
|               |                  |                           |
-----------------------------------------------------------------
| (staleness row, only while something is stale)                 |
| Footer                                                          |
--------------------------------------------------------------------
```

- **Roots Nav** — the list of roots in this session: the ad-hoc root (if any)
  first, then your starred roots. Each entry renders on two lines — the folder
  name in bold above its full path, unbold — so two same-named roots in
  different places stay distinguishable; moving up/down steps one whole entry
  at a time. Broken roots (path no longer exists) are marked and can't be
  activated. A starred root is marked `★`.
- **Files Nav** — lists documents from the *active* root, in one of three
  [Browse Modes](#browsing-tree-recent-starred): `Tree` (the file tree,
  default), `Recent`, or `Starred`. Switching roots re-scopes this pane but
  leaves whatever document you're reading untouched; each root's Tree
  expansion/selection, Browse Mode and Browse Scope are remembered
  independently for the rest of the session, so switching back restores it
  exactly as you left it. Files with uncommitted changes are marked `M`
  (modified) or `A` (new / untracked) — see [Git status markers](#git-status-markers)
  below.
- **Document Viewer** — a document, rendered as styled markdown (headings,
  tables, syntax-highlighted code fences, lists, blockquotes, rules, emphasis)
  and wrapped to the pane width. Read-only: no cursor, no selection, no copy —
  just scrolling. A session always has one Document Viewer (DV1) and may have
  a second (DV2) beside it — see [A second Document Viewer](#a-second-document-viewer).
- **Footer** — key hints for whatever pane has focus, plus transient messages
  (a rejected launch path, a refused action). A message clears on your next
  keypress. Hints truncate gracefully on a narrow terminal, always keeping
  `help` and `quit`.

Only one pane has focus at a time; most keys route to it. Every pane shows
the key that jumps focus to it, right-aligned on its own title row (e.g.
`Files` ... `[f]`, DV2's `[^o]`) — dropped if the pane is too narrow to fit
both. See [Focus](#focus) below.

## Keybindings

Bindings are provisional (cheap to retune) but consistent across panes: movement
is vim-style (`j`/`k`/arrows, plus `g`/`G` for top/bottom where it applies).

### Global (any non-modal pane)

| Key | Action |
|---|---|
| `?` | open the Help modal (full keybinding list for the focused pane) |
| `ctrl+n` | toggle the Navigator (hide/restore the whole left sidebar) |
| `r` | focus the Roots Nav (reveals the Navigator if hidden) |
| `f` | focus the Files Nav (reveals the Navigator if hidden) |
| `o` | focus DV1 |
| `ctrl+o` | focus DV2 (no-op if DV2 doesn't exist — a bare focus request never creates it) |
| `ctrl+x` | close DV2, wherever focus is (no-op if DV2 doesn't exist); focus moves to DV1 if DV2 held it |
| `R` | refresh the Active Root's file tree and every open Document Viewer's document from disk |
| `e` | edit the active document in `$EDITOR` |
| `/` | start find in the active document |
| `n` / `N` | next / previous find match (once a search has matches) |
| `space` | open the Picker (fuzzy find files or grep across files, see below) |
| `q` / `ctrl+c` | quit |
| `esc` | context-dependent cancel (dismiss find, close Help, close Picker) |

There is no pane-cycling key: every pane is reached by a direct focus key
(`r`/`f`/`o`/`ctrl+o`) instead — see [Focus](#focus).

`e`, `/`, `n` and `N` are reachable from any pane, not just a Document
Viewer's, but always act on DV1's document — unless focus is currently on
DV2, in which case they act on DV2's.

### Roots Nav

| Key | Action |
|---|---|
| `j`/`k` or `↓`/`↑` | move selection |
| `enter` | activate the selected root |
| `s` | star / unstar the selected root |

### Files Nav

| Key | Action |
|---|---|
| `j`/`k` or `↓`/`↑` | move selection |
| `l`/`→` | expand the selected folder (`Tree` mode) |
| `h`/`←` | collapse the selected folder (`Tree` mode) |
| `enter` | open the selected file into DV1 and move focus there |
| `ctrl+enter` (fallback `ctrl+l`) | open the selected file into DV2 instead, creating it if needed, and move focus there |
| `p` | peek at the selected file in DV1, without opening it or moving focus (press `p` again, or move the selection, to return) |
| `ctrl+p` | peek at the selected file in DV2 instead, creating it if needed, without moving focus |
| `s` | star / unstar the selected document |
| `tab` | cycle Browse Mode: `Tree` → `Recent` → `Starred` → `Tree` … |
| `shift+tab` | cycle the current mode's Browse Scope (no-op in `Tree`) |

See [Browsing: Tree, Recent, Starred](#browsing-tree-recent-starred).

### Document Viewer (DV1 or DV2, whichever has focus)

Each viewer acts on its own document only — scrolling, finding, starring,
editing and history in DV1 never touch DV2, and vice versa.

| Key | Action |
|---|---|
| `j`/`k` or `↓`/`↑` | scroll one line |
| `ctrl+u` / `ctrl+d` | half-page up / down |
| `pgup`/`ctrl+b` , `pgdown`/`ctrl+f` | page up / down |
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
| *(typing)* | updates the Results list live (once 3+ characters are typed) |
| `↑` / `↓` | move the Results selection; the Preview follows it |
| `tab` | switch mode: File ↔ Grep (the typed query carries over) |
| `shift+tab` | switch Scope: Active Root → All Roots → Starred (the typed query carries over) |
| `alt+h` | toggle hiding dotfiles from the Results, for the rest of this session |
| `alt+s` | star / unstar the selected Result |
| `enter` | activate the selected Result into DV1 |
| `ctrl+enter` (fallback `ctrl+l`) | activate the selected Result into DV2 instead, creating it if needed |
| `esc` | dismiss the Picker — nothing about the session changes |

## Focus

Exactly one pane has focus at a time, and every pane is reached by a direct
key from anywhere non-modal: `r` (Roots Nav), `f` (Files Nav), `o` (DV1),
`ctrl+o` (DV2 — a no-op if DV2 doesn't exist). `r`/`f` reveal the Navigator
first if it's hidden; `o`/`ctrl+o` don't, since DV1/DV2 are never inside it.
There's no Tab-cycling. Each pane's own focus key is shown right-aligned on
its title row (`Files` ... `[f]`, DV2's `[^o]`), dropped if the pane is too
narrow for both title and badge.

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
scroll position each, and each walks its own [Document History](#document-history).
DV2 is session-only — always absent at launch, never persisted.

`R` (refresh) re-reads the Active Root's tree and both viewers' documents
(whichever exist) in one action, regardless of which pane has focus. The
[staleness indicator](#staleness-indicator) also now spans both viewers, not
just one.

## Starring roots and documents

Both roots and individual documents can be starred, independently of each
other — un-starring a root never touches the star on any document under it,
and vice versa.

**Roots:** press `s` in the Roots Nav to star or unstar the selected root.
Starred roots are what a plain `noggin` (no argument) opens next time,
marked `★` in the Roots Nav. They're recorded, one absolute path per line, in
a plain text file at:

- `$XDG_DATA_HOME/noggin/roots` (or the platform equivalent, e.g.
  `~/.local/share/noggin/roots` on Linux, `~/Library/Application Support/noggin/roots`
  on macOS)

The file is deliberately hand-editable — `.gitignore`-style comments (`#`) and
blank lines are fine, so you can fix a moved root by editing the path directly.
A missing or unreadable file is treated as "no starred roots" rather than an
error, and stale paths (root no longer exists) are kept, not silently pruned —
they just show up broken in the Roots Nav.

**Documents:** press `s` to star or unstar a document from the Files Nav
(the selected file), from a focused Document Viewer (that viewer's own
document — DV1 and DV2 can have different starred documents open at once),
or `alt+s` in the Picker (the selected Result). A star shows as `★` on Files
Nav rows (in every Browse Mode except `Starred`, where membership already
says it), to the left of the document's title in a Document Viewer, and on
Picker rows in either mode. A starred document survives un-starring its
root, moving to a different root, or that root simply not being part of a
later session — it's tracked by absolute path, not by root. A **broken
starred document** (path no longer on disk) still shows up starred, marked
`(missing)`, but can't be opened; it's reachable from the Files Nav's
`Starred` mode and the Picker's `Starred` scope even when no open root
covers it (see [Browsing](#browsing-tree-recent-starred)).

Document stars, along with each document's last-viewed time (used by
[Recent](#browsing-tree-recent-starred)), live in a second plain file
alongside the roots file:

- `$XDG_DATA_HOME/noggin/documents` (same directory as `roots` above)

Each line is `<0 or 1>\t<RFC3339Nano timestamp or ->\t<absolute path>`
(tab-separated); a starred entry is never evicted, only its timestamp is
cleared, if the store grows past its internal cap.

Note: both files live under noggin's *data* directory, not its config
directory (`$XDG_CONFIG_HOME`) — a config-directory `roots` file from an
older noggin is not read or migrated.

## What gets shown

A file is only shown/opened if it looks like markdown-ish prose:

- extensions `.md`, `.markdown`, `.txt` (case-insensitive)
- extensionless conventional names `README`, `LICENSE`, `CHANGELOG`
  (case-insensitive)

Folders are always shown, even empty ones or ones containing nothing readable
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
`Tree` (the file tree, the default), `Recent`, and `Starred`. A status row
below the pane's title always names the current mode and scope, e.g.
`Tree · -`, `Recent · Active Root`, `Starred · All Starred` — costing the
list one row of height in every mode.

`shift+tab` cycles the current mode's Browse Scope (a no-op in `Tree`,
which has none):

- **Recent** — `Active Root` ↔ `All Roots`. Lists documents by when they were
  last Activated (opened into a viewer) or sent to `$EDITOR`; peeking doesn't
  count. `Active Root` shows at most 100 documents; `All Roots` merges every
  Root's list by timestamp with no further cap.
- **Starred** — `Active Root` → `All Starred` → `Orphans`. `All Starred` is
  every starred document in the store, including ones under no current
  Root; `Orphans` is exactly those documents — starred, but under no
  session Root at all.

`Recent` and `Starred` are flat lists, not trees: each row shows the file
name on the left and, right-aligned, the base name of its Root — or, for an
orphan, its left-truncated parent path. `Recent` rows carry the usual `★`
star marker; `Starred` rows omit it (mode already implies it) and instead
mark a broken starred document with `(missing)`. Any mode/scope combination
with nothing to show says so directly (e.g. "No starred documents outside a
session Root.").

The current mode is remembered per Root, and the scope is remembered per
(Root, mode) pair, for the rest of the session — returning to a Root restores
what you last had there. Every session starts every Root on `Tree`.

## Document History

Each Document Viewer (DV1 and DV2 independently) keeps its own back/forward
Document History of the documents opened into it, walked with `[`/`]` while that
viewer has focus. It's browser-style: activating a new document after going
back discards the forward entries, and the same document can appear more
than once but never twice in a row. Going back or forward restores that
document's own most recently used scroll position in that viewer (remembered
per document, not per history entry).

History can walk into documents under other Roots without ever changing the
Active Root. The `$EDITOR` handoff and `R` (refresh) do not add history
entries — both re-render the same document in place instead. History and its
scroll memory are session-only for both viewers, dropped entirely when DV2
is closed or noggin exits.

## Editing

Press `e` to hand off to your editor: `$EDITOR` if set, otherwise `vi`. `e` is
reachable from any pane, and always acts on DV1's document unless focus is
currently on DV2, in which case it acts on DV2's. noggin releases the
terminal for the editor and resumes when it exits, re-reading the file from
disk so your edits show up immediately — there's no live file-watching, this
is the one point where it re-reads. With no document open in the targeted
viewer, `e` just reports that instead of doing nothing.

## Help

Press `?` anywhere for a modal listing every binding relevant to the pane you're
in (a superset of the footer's hints — it also lists things like cancel/next
match/previous match for the Document Viewer). Press `?` or `esc` again to close
it; whatever had focus keeps it.

## Frontmatter

YAML frontmatter at the top of a document is stripped before rendering — it's
never shown as a table or raw block.

## Keeping things in sync

noggin never watches the filesystem — everything on screen is a snapshot from
when the root or document was last read.

Press `R` from any pane to refresh: it re-reads the Active Root's file tree
and both DV1's and DV2's documents (whichever exist) from disk in one pass.
Files Nav expansion and selection are preserved (selection moves to the
nearest surviving ancestor if its exact node is gone); each viewer's own
scroll position is preserved independently, clamped to the end if its
document shrank past it; an in-progress or committed find may be lost in
either viewer. If a viewer's document no longer exists or can't be read,
that viewer goes empty. Nothing outside the Active Root's tree and the two
viewers' documents is touched — other roots, other documents, and Picker
results are unaffected. Refresh also recomputes the git status markers below,
and doesn't add a [Document History](#document-history) entry.

### Git status markers

Files with uncommitted changes are marked in the Files Nav: `M` for modified,
`A` for new/untracked. Only those two states are shown — staged vs. unstaged
isn't distinguished, and deleted or renamed files aren't marked. Status is
resolved against the git repository containing the Active Root (its root may
sit above the Root itself), not the Root path directly, so a Root that is a
subdirectory of a repo still shows correct markers. A Root outside any git
repository, or a machine without `git` on `PATH`, shows no markers, silently.
Markers are computed when a root is opened and recomputed on every `R`
refresh — they are never kept live in between.

### Staleness indicator

Every few seconds, noggin quietly checks whether the Active Root's tree, DV1's
document, or DV2's document (if it exists) has changed on disk since it was
last read. If any has, a dedicated highlighted row appears just above the
Footer reading "something changed on disk; press R to refresh" — it occupies
no space at all while nothing is stale. It never says *what* changed, and
never refreshes or re-reads anything on your behalf — pressing `R` is the
only way to act on it. It's one combined signal across tree/DV1/DV2, not
per-source; it clears the next time you refresh and stays cleared until
another change is detected.

## Finding across files (the Picker)

Press `space` from any pane, including with the Navigator hidden, to open the
Picker: a modal fuzzy-finder over the files in scope. It always opens the same
way — File mode, Active Root scope, empty Prompt — and never remembers
anything between opens, in this session or the next.

The Picker has two modes, switched in place with `tab` without losing what
you've typed:

- **File** (the default) — fuzzy-matches your query against file *names* only,
  ranked by match strength, with the matched characters highlighted in each
  row. Each row also shows the file's Root and its path relative to that
  Root, so two same-named files in different Roots stay distinguishable.
- **Grep** — searches for your query as literal text (no regex) across file
  *contents*, case-insensitively unless your query itself has an uppercase
  letter. One row per matching line, showing the file, its line number and
  the line's text, with the matched text highlighted.

Either mode searches one of three Scopes, cycled in place with `shift+tab`:
**Active Root**, **All Roots**, or **Starred** (every starred document,
including orphans — see [Browsing](#browsing-tree-recent-starred)). The title
above the Prompt names both mode and scope, e.g. `File · Active Root` or
`Grep · Starred`. All Roots silently skips any Broken root.

A query shorter than 3 characters doesn't search at all ("type at least 3
characters"); a query that matches nothing says "no results"; a result set
larger than 200 says so ("first 200 results (truncated)") instead of implying
that's everything. The Preview pane shows the selected Result's file as raw,
unrendered text — the File Picker from the top, the Grep Picker scrolled to
and highlighting the matched line — and you can scroll it without affecting
anything else. `enter` activates the selected Result: it becomes the Active
Document (landing on the matched line for a Grep Result, with next/previous
`n`/`N` traversal working from it, same as a typed `/` search), without
changing the Active Root or touching the Files Nav tree's expansion or
selection. `esc` dismisses the Picker with no side effects at all — same
Active Document, same Active Root, same scroll positions.

Only [Readable files](#what-gets-shown) are ever searched or shown.

### Requires `ripgrep`

The Picker is powered by [`ripgrep`](https://github.com/BurntSushi/ripgrep)
(the `rg` binary on your `PATH`) for both listing file names and searching
text. Without it, noggin still launches and reads documents normally — only
the Picker is affected: opening it shows "Picker unavailable: ripgrep (rg)
not found on PATH" instead of a search.

Because `rg` supplies the file list, its own ignore rules apply, not
noggin's Files Nav ignore rules: `.gitignore`/`.ignore` files are respected
(consistent with the Files Nav). Dotfiles and dot-directories (e.g. `.github/`)
are included by default in both modes, matching what the Files Nav tree
already shows — `.git/` itself stays excluded unconditionally, in either
pane. `alt+h` toggles hiding them again, in both modes at once; the title
above the Prompt gains a ` · Hidden` suffix while the toggle is on. It's
session-only — relaunching noggin resets it to showing dotfiles.

## Scope

noggin does not create, rename, move or delete files or folders, and it isn't an
editor — writing happens by handing off to `$EDITOR`.
