# Vision — C2a polish

<!-- Per-cycle. The raw, high-level goals for THIS cycle. Not formalized, possibly un-vetted;
     captured from the human during brainstorming. May contain contradictions and open ends —
     SPEC.md is where they get resolved. Free-form. Intent that reaches beyond this cycle
     belongs in ROADMAP.md, not here. The cycle key (C<n>) comes from the roadmap ledger. -->

## Goals

An interstitial cycle, inserted before `compare`: high-value tweaks that make noggin more
usable *while it is still being built out*. The user is dogfooding a half-built tool; this
cycle pays down the friction that dogfooding exposed.

Raw wish list as pitched:

- **deploy** — build `noggin` to a PATH-accessible location.
- **picker default** — Picker Results (all modes) should show hidden files; better yet, a
  visibility toggle.
- **refresh** — file-tree / document refresh. Not a full watcher: event-driven (window/app
  focus, manual, maybe pane focus).
- **Roots Nav layout** — two lines per entry: base name bold on line 1, path (same truncation
  treatment) unbold on line 2.
- **commands**
  - `r` — focus Roots Nav
  - `f` — focus Files Nav
  - `R` — manual refresh
  - `p` — in Files Nav, render the selected document *without* moving Focus to the Document
    Viewer
  - remap Picker from `ctrl+p` to `<space>`
- **git annotations**
  - per file (M / A?)
  - per line (+ / - / ~?)
- **toggleable markdown render mode**
  - *read mode*: current word wrapping; current margin highlighting for Find results (or none
    if tricky); no linewise git annotations
  - *info/results mode*: no word wrapping (to enable frictionless highlighting and a gutter);
    exact-word Find highlighting; linewise git annotations

## Resolved in interview (round 1)

- **Refresh** scope: Active Root's file tree + Active Document only. Not other Roots, not the
  Picker. Pain: reselecting the Active Document after an out-of-process change; relaunching
  the app when a file is added. Tree expansion + selection preserved; Document scroll position
  preserved; Find state may be lost. Deleted selection: walk up to the nearest surviving node.
- Manual-only refresh is acceptable — focus/blur triggers dropped. An interval-based
  *staleness indicator* (no auto-refresh, ~5–10s, shows that tree or document changed) is
  wanted if cheap.
- **Hidden files**: pain is not being able to Find/Grep across dotfiles that are visible in the
  tree. Minimum: align Picker defaults to show dotfiles. Ideal for this cycle: a session-only,
  Picker-only dotfile toggle. No change to Files Nav. `.gitignore` behaviour unchanged.
  `.git/` stays excluded.
- **`<space>`** for the Picker is deliberate leader-key emulation, not pager semantics.
  `ctrl+p` is dropped, not kept as an alias.
- **`p`** is an addition, not a correction: `enter` keeps its focus-jumping behaviour. `p` is
  accepted as a cheap step toward the deferred Follow mode.
- **`r`/`f`** never hide the Navigator; they focus it, revealing it first if hidden. Active from
  Roots Nav, Files Nav and Document Viewer; inert while the Picker is open.
- **Deploy**: `make` target that builds and installs to `/usr/local/bin`. No versioning.

## Non-goals / boundaries

- Not the render-mode toggle, not the per-line git gutter, not exact-word Find highlighting,
  not horizontal scrolling. Cut on the strength of `findings/git-status-gutter-approach.md`,
  and deliberately *not* parked on the roadmap: re-raised only if still felt after this cycle.
- Not any form of automatic refresh. S10 tells you something changed; it never acts.
- Not an editor, as always.

## Notes

- Pitched as "short". Whether the whole list fits one cycle is the first thing to settle.
- The render-mode idea is the user's own answer to a C1 dead end: exact-span Find highlighting
  was abandoned because Glamour's ANSI escapes drift the offsets
  (`knowledge/viewport-search-ansi-offsets.md`). A no-wrap, lightly-styled mode is plausibly
  the enabler — that connection is the interesting part of the pitch.
- No git tooling exists in the project at all today; `.gitignore` parsing is hand-rolled and
  `.git/` is merely excluded from the tree walk.
- No focus/blur reporting is wired up today; bubbletea can report terminal focus.
- `findings/git-status-gutter-approach.md` reframes the render-mode idea: the dual-mode trick
  is real and cheaper than a renderer fork, but (a) Find highlighting is *not* coupled to wrap
  mode — it already works in rendered-line space, so "exact-word highlighting" is not a mode-2
  benefit; (b) an unwrapped mode needs horizontal scrolling noggin does not have, against D4;
  (c) a per-*file* indicator is mapping-free and cheap, unlike the per-line gutter.
- Existing Unscheduled entries this pitch collides with: hidden-file/gitignored visibility
  toggles + `nogginignore` (from C1); user-configurable keybindings (from C1); a
  rendered-markdown Preview pane (from C2).
