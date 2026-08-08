# Spec — C2a polish

<!-- Per-cycle. Functional, non-technical, concrete. This is the contract of intent: what the
     deliverables ARE, not how they are built. Keep it product-level so upfront errors don't
     cascade into implementation. Durable keys (S1, S2...) — never renumber. Mid-cycle
     changes go through /brainstorm amend: supersede or suffix keys, renew the agreed date;
     charm-replan then reconciles the plan. Keys are scoped to this cycle — durable docs
     referencing them qualify with the cycle key (C2a/S2). -->

Status: agreed 2026-08-05

An interstitial cycle inserted before `C3 compare`: friction that dogfooding a half-built
noggin exposed. Every deliverable is independently useful and independently cuttable.

## Deliverables

- **S1 — Install target.** `make install` places a `noggin` binary in a PATH-accessible
  location. Destination is `$(PREFIX)/bin` with `PREFIX ?= /usr/local`, so it is overridable
  for testing and for users without write access to `/usr/local`. Requires an explicit
  `make build` first — `install` does not build.

- **S2 — Manual refresh.** A `R` binding re-reads the Active Root's file tree and the Active
  Document from disk, on demand, from any pane. Nothing else is refreshed (other Roots, Broken
  Root re-checks and Picker results are out of scope). Across a refresh:
  - Files Nav expansion state and selection are preserved.
  - The Document Viewer's scroll position is preserved; if the document shrank past the current
    offset, the offset clamps to the end.
  - Find state may be lost.
  - If the selected tree node no longer exists, selection moves to the nearest surviving
    ancestor-ward node.
  - If the Active Document no longer exists or is unreadable, the Document Viewer becomes empty.
  - Per-file git status markers (S8) are recomputed in the same pass.

- **S3 — Tree state memory across Root switches.** Files Nav expansion state and selection are
  remembered per Root, so switching away and back restores the tree as it was. (Retires the
  C1 "accepted friction" Unscheduled entry; included because it is the same mechanism S2 needs.)

- **S4 — Dotfiles in the Pickers.** Both Picker modes include dotfiles in their Results by
  default, matching what the Files Nav tree already shows. A session-only `alt+h` toggle hides
  and shows them, shared by both modes, reset on relaunch. `.git/` remains excluded
  unconditionally. `.gitignore` handling is unchanged in every pane.

- **S5 — Two-line Roots Nav rows.** Each Roots Nav entry occupies two lines: the base name in
  bold on line 1, the path on line 2, unbold, with the existing left-truncation treatment. The
  active (`●`) and starred (`★`) markers and the `(missing)` broken-root suffix stay on line 1.
  Both lines of the selected entry carry the selection highlight. Movement steps one entry —
  two lines — at a time. The Navigator's height split is not revisited; fewer visible roots is
  accepted.

- **S6 — Keybindings.** `r` focuses the Roots Nav; `f` focuses the Files Nav; both reveal the
  Navigator first if it is hidden, and neither ever hides it. Both are active from the Roots
  Nav, Files Nav and Document Viewer, and inert while the Picker is open. `R` is manual refresh
  (S2). `p` is Peek (S7), in the Files Nav only. The Picker opens with `<space>` instead of
  `ctrl+p`; `ctrl+p` is retired, not kept as an alias.

- **S7 — Peek.** In the Files Nav, `p` renders the selected document in the Document Viewer
  *without* changing the Active Document and without moving Focus. Deliberately minimal — an
  experiment that is allowed to feel unfinished rather than accrue premature machinery:
  - The Document Viewer shows an inverted header identifying the Peeked document.
  - A Peeked document cannot be scrolled and cannot be searched with Find.
  - `e` (Edit) is a silent no-op while Peeking.
  - Peek is cleared by: any selection change, any Focus change (including `l`/`tab`/`f`),
    opening the Picker, a second `p`, and a refresh. Clearing restores the Active Document at
    the scroll position it had.
  - With no Active Document, clearing reverts to the empty Viewer.
  - `p` on the Active Document is a no-op. `<enter>` is the only promotion path, keeping its
    current behaviour.

- **S8 — Per-file git status markers.** Files whose git status is modified or new are marked
  with `M` and `A` respectively. Only those two states: staged/unstaged is not distinguished,
  and deleted and renamed are not marked. Roots that are not inside a git repository show no
  markers, silently. Status is resolved relative to the containing repository's root, not the
  Root. Dirty submodules and worktree/bare-`.git` setups are explicitly not handled. Markers
  are computed on load and on refresh (S2) — never on a background watch.

- **S9 — Footer and Help Modal reflect reality.** Every new, changed and retired binding from
  S4/S6/S7 is correct in the Footer hints and the Help Modal for each pane. The cycle is not
  done while `?` still advertises `ctrl+p`.

- **S10 — Staleness indicator.** *Pre-agreed cut: planned last, dropped without ceremony if the
  cycle tightens.* A periodic check (~5–10s) reports in the UI that the Active Document or the
  Active Root's tree has changed on disk, without refreshing anything. It only ever says
  "something changed, press `R`" — never what changed. It must not re-walk the tree to do this,
  and must not perceptibly cost anything on a large Root.

## Out of scope

Deliberately dropped, not deferred — no roadmap entry, to be re-litigated only if the tweaks
above leave the need still felt (user's call, this cycle):

- A toggleable markdown render mode (read vs info/results mode).
- A per-line git gutter (adds/modifies/deletes in the Document Viewer).
- Exact-word / inline-span Find highlighting; the current whole-line indication stands.
- Horizontal scrolling in the Document Viewer.

Out for this cycle, existing intent unchanged:

- Any automatic refresh: file watching, and focus/blur-triggered refresh (S10 indicates, never
  acts).
- Refreshing anything beyond the Active Root's tree and the Active Document.
- Staged-vs-unstaged, deleted, renamed, submodule and worktree git states.
- A gitignored-file visibility toggle, and `nogginignore` (remains Unscheduled, from C1).
- Any change to Files Nav dotfile or gitignore behaviour.
- Follow mode / auto-preview on selection change (remains Unscheduled, from C1); S7 is a
  cheaper manual stand-in.
- User-configurable keybindings (remains Unscheduled, from C1); S6 is build-time as usual.
- Build versioning / `--version`; renaming the bare `noggin` module path (remains Unscheduled).
- Per-root *document* state (C4 `recall`); S3 covers tree shape only.

## Acceptance (cycle-level)

- `make build && make install` puts a working `noggin` on PATH, and the same works with a
  custom `PREFIX` under a directory the tester owns.
- A file created, edited or deleted outside noggin becomes visible in the Files Nav and the
  Document Viewer after one `R` — without losing the expanded tree, the selection, or the
  reading position.
- Switching Root away and back leaves the tree as it was found.
- A dotfile visible in the Files Nav tree is findable by both Picker modes with no extra
  keystroke, and `alt+h` hides and restores dotfiles for the rest of the session.
- The Roots Nav shows base name and path on separate lines, and navigating it moves one entry
  per keypress.
- `r`, `f`, `R`, `p` and `<space>` all do what this spec says from every pane where they apply;
  `ctrl+p` no longer opens the Picker anywhere.
- `p` on a Files Nav file shows it with a Peek header and leaves Focus in the tree; moving the
  selection restores the previous document at its previous scroll position.
- Modified and new files carry `M` / `A` markers that are correct after an out-of-process
  `git add` / edit followed by `R`; a Root outside any repository shows no markers and no error.
- The Footer and Help Modal match the delivered bindings in every pane.
- `make check` and `make check-pty` pass.

## Open questions

Resolved at agreement:

- **Q-A — S8 markers appear in Files Nav rows only.** Not in File Picker or Grep Picker
  Results: the Pickers are transient and already dense with path text, and the tree is where
  "what have I touched" gets asked.
- **Q-B — deleted files stay unmarked.** S2's selection-walk-up and empty-Viewer behaviour is
  the deletion affordance; a marker for a file gone from the tree has nowhere to render.

Left to the plan:

- **S10's indicator surface** (Footer marker vs elsewhere) is deliberately unspecified.
