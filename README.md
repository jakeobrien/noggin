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
|---------------|                                 |
| Roots Nav     |                                 |
|---------------|                                 |
| Files Nav     |                                 |
|               |                                 |
|---------------------------------------------------|
| Footer                                             |
-------------------------------------------------------
```

- **Roots Nav** — the list of roots in this session: the ad-hoc root (if any)
  first, then your starred roots. Broken roots (path no longer exists) are marked
  and can't be activated.
- **Files Nav** — the file tree of the *active* root. Switching roots re-scopes
  this pane but leaves whatever document you're reading untouched.
- **Document Viewer** — the active document, rendered as styled markdown
  (headings, tables, syntax-highlighted code fences, lists, blockquotes, rules,
  emphasis) and wrapped to the pane width. Read-only: no cursor, no selection, no
  copy — just scrolling.
- **Footer** — key hints for whatever pane has focus, plus transient messages
  (a rejected launch path, a refused action). A message clears on your next
  keypress. Hints truncate gracefully on a narrow terminal, always keeping
  `help` and `quit`.

Only one pane has focus at a time; most keys route to it.

## Keybindings

Bindings are provisional (cheap to retune) but consistent across panes: movement
is vim-style (`j`/`k`/arrows, plus `g`/`G` for top/bottom where it applies).

### Global (any pane)

| Key | Action |
|---|---|
| `tab` / `shift+tab` | next / previous pane |
| `?` | open the Help modal (full keybinding list for the focused pane) |
| `ctrl+n` | toggle the Navigator (hide/restore the whole left sidebar) |
| `e` | edit the active document in `$EDITOR` |
| `/` | start find in the active document |
| `n` / `N` | next / previous find match (once a search has matches) |
| `q` / `ctrl+c` | quit |
| `esc` | context-dependent cancel (dismiss find, close Help) |

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
| `l`/`→` | expand the selected folder |
| `h`/`←` | collapse the selected folder |
| `enter` | open the selected file in the Document Viewer and move focus there |

### Document Viewer

| Key | Action |
|---|---|
| `j`/`k` or `↓`/`↑` | scroll one line |
| `ctrl+u` / `ctrl+d` | half-page up / down |
| `pgup`/`ctrl+b` , `pgdown`/`ctrl+f` | page up / down |
| `g`/`home` , `G`/`end` | jump to top / bottom |

### While finding (`/`)

| Key | Action |
|---|---|
| *(typing)* | updates matches live |
| `enter` | commit the search — matches stay live for `n`/`N` |
| `esc` | dismiss the search entirely |

## Starring roots

Press `s` in the Roots Nav to star or unstar the selected root. Starred roots
are what a plain `noggin` (no argument) opens next time. They're recorded, one
absolute path per line, in a plain text file at:

- `$XDG_CONFIG_HOME/noggin/roots` (or the platform equivalent of
  `os.UserConfigDir()`, e.g. `~/.config/noggin/roots` on Linux)

The file is deliberately hand-editable — `.gitignore`-style comments (`#`) and
blank lines are fine, so you can fix a moved root by editing the path directly.
A missing or unreadable file is treated as "no starred roots" rather than an
error, and stale paths (root no longer exists) are kept, not silently pruned —
they just show up broken in the Roots Nav.

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

## Editing

Press `e` on the active document to hand off to your editor: `$EDITOR` if set,
otherwise `vi`. noggin releases the terminal for the editor and resumes when it
exits, re-reading the file from disk so your edits show up immediately — there's
no live file-watching, this is the one point where it re-reads. With no document
open, `e` just reports that instead of doing nothing.

## Help

Press `?` anywhere for a modal listing every binding relevant to the pane you're
in (a superset of the footer's hints — it also lists things like cancel/next
match/previous match for the Document Viewer). Press `?` or `esc` again to close
it; whatever had focus keeps it.

## Frontmatter

YAML frontmatter at the top of a document is stripped before rendering — it's
never shown as a table or raw block.

## Scope

noggin does not create, rename, move or delete files or folders, and it isn't an
editor — writing happens by handing off to `$EDITOR`. Fuzzy find/grep across
files, multiple documents on screen at once, and document history/recall are not
part of this release.
