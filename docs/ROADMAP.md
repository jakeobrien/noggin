# Roadmap

## North star

noggin is **a browser of document collections** — a thoughtspace for navigating,
synthesizing and remembering knowledge drawn from several collections at once. Search,
navigation and consumption; **never editing**, and never authoring. Light text input at
most. It works *with* your editor and your other tools, not instead of them.

Collections are admitted for a specific value they carry, not accumulated on whim:

| Collection | Value it carries |
|---|---|
| Notes | personally-relevant information; the **synthesis target** |
| Wikipedia | authoritative, highly trustable reference |
| Dictionary / thesaurus | the universal glossary — the medium everything else is written in |

The value compounds through two cross-collection features: **starring** and
**link-following** — connections and collections of your own that span sources.

Good looks like: `noggin .` and every relevant collection is instantly in reach, including
the cwd as an ad-hoc root; switching or comparing across collections costs a keystroke, not
a context reload. Explicitly not an editor — the editor boundary is a deliberate, clean
handoff, and the same boundary applies to anything that *creates* documents rather than
serving them (LLM chat included: hand off to a chat TUI, read its saved transcripts back as
an ordinary collection).

Why not the incumbents: Obsidian is mouse-first, its vaults are heavy to switch and cannot
sit side-by-side, and it has no ad-hoc source; Glow has no search, no side-by-side, no
persisted sources; nvim renders markdown poorly and is unfocused for this task.


## Roadmap Items

### Headline features

- Tags
- Document outline
- File management (add, move, copy, delete, duplicate)
- A dictionary / thesaurus **Query Source**
- **wiki-link and URL following** (now also covering **following links inside Wikipedia articles**
  render as plain text) — carved out because all three need a notion of "which link is
  current", i.e. the document cursor C1/D4 refused. Revisit as one piece: whatever earns the
  cursor should pay for all of them at once. (maybe a mode: (Link mode) that can jump from link to link)
- Backlinks; daily/zettelkasten note creation; outline re-ordering (from C1)
- Browse-by-results: send Picker Results to the Files Nav 
- Select mode + copy in the Document Viewer

### Improvements

- Regex (or fuzzy) Grep Picker queries
- A rendered-markdown Preview pane instead of raw text
- Reveal the activated Result in the Files Nav tree — switch Root, expand ancestors, select
- "Follow" mode for Files Nav (peeks files as they are selected)
- Key shortcuts for opening Picker in different modes

### Config

- A `nogginignore` file plus hidden-file / gitignored-file visibility toggles 
- User-configurable keybindings (from C1; keymap is build-time-only for now)

### Tech debt

- Rename `go.mod`'s module path from the bare `noggin` to a resolvable path before
  `go install`/external import ever needs it
- Per-viewer / per-source staleness signal granularity, instead of one combined signal 
  (deliberately kept a single unspecific signal spanning tree + both viewers)
- A shared exclusive-input-mode dispatch abstraction across modals (currently hand-duplicated
  per modal in model.go) — the Picker became the third modal in C2 and still didn't justify
  it (C2/D35: the three are less alike than the count suggests); only worth it if a fourth
  modal is ever added, none planned yet (from C1 code review T31, reaffirmed C2/D35)
- A shared save/load helper for the Roots and document stores (`store.go`/`docstore.go` both
  hand-roll `os.MkdirAll`-then-`os.WriteFile`) — only worth it if a third store format shows up
  (from C4 code review, TRIAGE T3)
