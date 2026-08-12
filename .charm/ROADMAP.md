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

| Key | Name               | Status  | Outcome / intent | Archive |
|-----|--------------------|---------|------------------|---------|
| C1  | read-across-roots  | done    | Delivered all 12 spec'd deliverables (S1–S12) across 26 plan tasks; none of D8's pre-agreed cuts were needed | `archive/C1-2026-08-read-across-roots/` |
| C2  | find               | done    | Fuzzy filename find and literal grep with preview, scoped to Active Root or All Roots, ripgrep-backed | `archive/C2-2026-08-find/` |
| C2a | polish             | done    | Install target, manual refresh, per-Root tree memory, dotfile Picker toggle, two-line Roots Nav, new bindings (r/f/R/p/space), Peek, per-file git markers, and a footer-width staleness-indicator fix | `archive/C2a-2026-08-polish/` |
| C3  | compare            | done    | Second Document Viewer (DV2), ctrl+ targeting from Files Nav/Picker, four-way direct focus keys replacing Tab-cycling, ctrl+x close, dual-viewer refresh/staleness, session-only DV2 state | `archive/C3-2026-08-compare/` |
| C4  | recall             | done    | Delivered all 10 spec'd deliverables (S1–S10) across 46 plan tasks; none of D63's four pre-agreed cuts were needed | `archive/C4-2026-08-recall/` |
| C5  | organize           | planned | Non-hierarchical structure and document shape: tags, outline, file management | - |

## Unscheduled

- Viewing source code and other non-markdown text files, with code-block styling (from C1)
- A `nogginignore` file plus hidden-file / gitignored-file visibility toggles (from C1)
- Editing in a tmux pane beside a live noggin, instead of suspend-and-resume (from C1; C3's
  dual viewers make concurrent editing meaningful, still not delivered)
- Multiple path arguments at launch; adding and removing roots in-session (from C1)
- User-configurable keybindings (from C1; keymap is build-time-only for now)
- Browse-by-results: send Picker Results to the Files Nav (from C1)
- Regex (or fuzzy) Grep Picker queries; C2 is literal-only (from C2, D29)
- A rendered-markdown Preview pane instead of raw text (from C2, D30/D33)
- Reveal the activated Result in the Files Nav tree — switch Root, expand ancestors, select
  (from C2, D31)
- File Picker matching against path components, not just the file name (from C2, D31)
- Picker memory: last mode, last Scope, last query (from C2, D31; see repeat-last-search below)
- Select mode + copy in the Document Viewer (from C1); **wiki-link and URL following** (from
  C4) — carved out of C4 because both need a notion of "which link is current", i.e. the
  document cursor C1/D4 refused. Revisit as one piece: whatever earns the cursor should pay for
  all of them at once.
- "Follow" mode: auto-preview on selection change (from C1); any targeting for it is also
  unscheduled (C3)
- Search over documents opened this session; repeat-last-search (from C1)
- Backlinks; daily/zettelkasten note creation; outline re-ordering (from C1)
- Rename `go.mod`'s module path from the bare `noggin` to a resolvable path before
  `go install`/external import ever needs it (from C1, code review)
- Per-viewer / per-source staleness signal granularity, instead of one combined signal (from
  C3; C3 deliberately kept a single unspecific signal spanning tree + both viewers)
- A shared exclusive-input-mode dispatch abstraction across modals (currently hand-duplicated
  per modal in model.go) — the Picker became the third modal in C2 and still didn't justify
  it (C2/D35: the three are less alike than the count suggests); only worth it if a fourth
  modal is ever added, none planned yet (from C1 code review T31, reaffirmed C2/D35)
- A shared save/load helper for the Roots and document stores (`store.go`/`docstore.go` both
  hand-roll `os.MkdirAll`-then-`os.WriteFile`) — only worth it if a third store format shows up
  (from C4 code review, TRIAGE T3)
