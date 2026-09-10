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

<!-- North star dilated 2026-08-14 (C4a). The original — "reduce friction to accessing and
     recording information" across note sources only — was a deliberately narrow hedge
     against the value not materialising. See DECISIONS.md. -->

## Cycles

| Key | Name               | Status  | Outcome / intent | Archive |
|-----|--------------------|---------|------------------|---------|
| C1  | read-across-roots  | done    | Delivered all 12 spec'd deliverables (S1–S12) across 26 plan tasks; none of D8's pre-agreed cuts were needed | `archive/C1-2026-08-read-across-roots/` |
| C2  | find               | done    | Fuzzy filename find and literal grep with preview, scoped to Active Root or All Roots, ripgrep-backed | `archive/C2-2026-08-find/` |
| C2a | polish             | done    | Install target, manual refresh, per-Root tree memory, dotfile Picker toggle, two-line Roots Nav, new bindings (r/f/R/p/space), Peek, per-file git markers, and a footer-width staleness-indicator fix | `archive/C2a-2026-08-polish/` |
| C3  | compare            | done    | Second Document Viewer (DV2), ctrl+ targeting from Files Nav/Picker, four-way direct focus keys replacing Tab-cycling, ctrl+x close, dual-viewer refresh/staleness, session-only DV2 state | `archive/C3-2026-08-compare/` |
| C4  | recall             | done    | Delivered all 10 spec'd deliverables (S1–S10) across 46 plan tasks; none of D63's four pre-agreed cuts were needed | `archive/C4-2026-08-recall/` |
| C4a | wikipedia-spike    | done    | GO, with a known limitation (D112): conversion quality clears the reading bar on 4/4 samples, one (`E (mathematical constant)`) with a residual maths-rendering gap; spike code kept as reference (D97), not integrated. Gates the Wikipedia-collection, dictionary-collection and C6 entries below | `archive/C4a-2026-09-wikipedia-spike/` |
| C5  | organize           | withdrawn | Un-planned 2026-09-07 before any spec: priorities moved to Sources first. Its goals (tags, outline, file management) live in Unscheduled. Key retired, never reused | - |
| C6  | sources            | done    | Delivered all 13 spec'd deliverables (S1–S13 + S9a) across M1–M10 (~90 tasks incl. replan-added hygiene/bugfix work); none of the 3 pre-agreed cuts (Recent, Root names, background tint) were needed | `archive/C6-2026-09-sources/` |
| C7  | <unnamed>          | planned | Not yet scoped — run `/brainstorm` to name and spec it. Strong candidates waiting in Unscheduled below: the dictionary Query Source (completes the Sources story just built) and the style pass (flagged "required before shipping"; S10/S11 are provisional pending it) | - |

<!-- C4a is a suffixed interstitial key per D42: C5 was already referenced by intent in
     DECISIONS.md, knowledge/ and archived cycle docs ("Tags are C5"), so renumbering was
     not available. Ledger order is independent of key order. C6 runs before the withdrawn
     C5 for the same reason: D94 had already scheduled the collections question as C6. -->

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
- Tags, outline and file management — non-hierarchical structure and document shape (the
  withdrawn C5 organize; D10 on tag notation is still open)
- A dictionary / thesaurus **Query Source** (from C4a; unblocked per D112; C6 built the
  Source model specifically to admit it, and proved the fit — S1–S7 restate cleanly for it
  with only the name changed, per C6's own acceptance test)
- A style pass: dark/light mode, palette, style sheet — required before shipping; C6's
  background-focus and de-emphasis choices are provisional until then (from C6)
- Reading saved LLM-chat transcripts as an ordinary collection, plus a handoff key to a chat
  TUI (from C4a; chat itself is deliberately *not* noggin's job)
- Select mode + copy in the Document Viewer (from C1); **wiki-link and URL following** (from
  C4), now also covering **following links inside Wikipedia articles** (from C6; today they
  render as plain text) — carved out because all three need a notion of "which link is
  current", i.e. the document cursor C1/D4 refused. Revisit as one piece: whatever earns the
  cursor should pay for all of them at once.
- Non-English Wikipedia editions (from C6; C6 is English-only by spec)
- Offline reading of starred Wikipedia articles — persisting article content instead of only
  a pointer (from C6; C6 deliberately never stores content)
- Launching noggin directly into a Wikipedia article via a CLI URL/title argument (from C6)
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
