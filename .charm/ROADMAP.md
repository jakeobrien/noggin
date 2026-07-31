# Roadmap

## North star

Reduce friction to accessing and recording information. Several sources of notes and docs
are live at once — personal notes, shared work notes, and the local docs of whatever
projects are in flight. noggin makes all of them reachable in one keyboard-driven session:
find it fast, read it well, compare across sources, hand off to `$EDITOR` when writing is
the point.

Good looks like: `noggin .` and every relevant source is instantly in reach, including the
cwd as an ad-hoc root; switching or comparing across sources costs a keystroke, not a
context reload. Explicitly not an editor — the editor boundary is a deliberate, clean
handoff.

Why not the incumbents: Obsidian is mouse-first, its vaults are heavy to switch and cannot
sit side-by-side, and it has no ad-hoc source; Glow has no search, no side-by-side, no
persisted sources; nvim renders markdown poorly and is unfocused for this task.

## Cycles

| Key | Name             | Status  | Outcome / intent                                                            | Archive |
|-----|------------------|---------|-----------------------------------------------------------------------------|---------|
| C1  | read-across-roots | active  | `noggin .` — saved + ad-hoc roots, switch between them, browse a tree, read rendered markdown, hand off to the editor | -       |
| C2  | find              | planned | Telescope-grade fuzzy filename find and grep with preview, scoped to a root or all roots | -       |
| C3  | compare           | planned | More than one document on screen at once, with keyboard targeting           | -       |
| C4  | recall            | planned | The session remembers: document history, starred documents, recent, wiki-links | -       |
| C5  | organize          | planned | Non-hierarchical structure and document shape: tags, outline, file management | -       |

## Unscheduled

- Viewing source code and other non-markdown text files, with code-block styling (from C1)
- A `nogginignore` file plus hidden-file / gitignored-file visibility toggles (from C1)
- Editing in a tmux pane beside a live noggin, instead of suspend-and-resume (from C1)
- Remembering tree expansion and selection across root switches (from C1; accepted friction)
- Multiple path arguments at launch; adding and removing roots in-session (from C1)
- User-configurable keybindings (from C1; keymap is build-time-only for now)
- Modifier-multiplied navigation acceleration, 5×/15×/50× (from C1; half-page/page may suffice)
- Move the starred-roots store from config to a data location once it stops needing hand edits (from C1)
- Persisted (cross-session) document history (from C1 vision triage)
- Browse-by-results: send finder results to the Files Nav (from C1)
- Select mode + copy in the Document Viewer (from C1)
- "Follow" mode: auto-preview on selection change (from C1)
- Search over documents opened this session; repeat-last-search (from C1)
- Backlinks; daily/zettelkasten note creation; simple in-app editing; outline re-ordering (from C1)
- Audio input; encryption; git sync; wikipedia pane; agent chat pane (from C1)
- Rename `go.mod`'s module path from the bare `noggin` to a resolvable path before
  `go install`/external import ever needs it (from C1/T6, code review)
