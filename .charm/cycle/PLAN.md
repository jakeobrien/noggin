# Plan — C2a polish

<!-- Per-cycle. The spec decomposed into milestones and tasks. Technically aware but
     high-level: state WHAT and the done-condition, not step-by-step HOW — trust executor
     judgment. Every task MUST have an acceptance criterion, machine-checkable where possible,
     and an effort estimate: a relative 1-5 gut-scoping (1 trivial, e.g. fixing a few typos;
     5 very high, e.g. large or complicated tasks) — not a rigorous prediction, just how the
     task looks at scoping time. Durable keys: milestones M1, M2...; tasks M1.1, M1.2...
     Never renumber; append or suffix (M1.2a) when inserting. Keys are scoped to this cycle —
     durable docs referencing them qualify with the cycle key (C3/M1.2). Note dependencies
     where ordering matters. -->

## M1 — Install target

Delivers: S1

- **M1.1** — `make install` places `bin/noggin` at `$(PREFIX)/bin`, `PREFIX ?= /usr/local`,
  without building — done when: `make build && make install` puts a working `noggin` on
  PATH, and `make install PREFIX=<custom-dir>` places it there instead, without requiring
  `install` itself to invoke `build` — effort: 1.

## M2 — Navigation keybindings

Delivers: S6 (focus + Picker-open portion; `R` and `p` land with their own features in M5/M7)

- **M2.1** — `r` focuses the Roots Nav and `f` focuses the Files Nav, from the Roots Nav, Files
  Nav or Document Viewer, revealing the Navigator first if it is hidden and never hiding it;
  both are inert while the Picker is open — done when: pressing `r`/`f` from any of those three
  panes moves Focus and reveals a hidden Navigator, and pressing them with the Picker open does
  nothing — effort: 2.
- **M2.2** — The Picker opens with `<space>`; `ctrl+p` is retired outright, not kept as an
  alias — done when: `<space>` opens the Picker from Roots Nav, Files Nav and Document Viewer,
  and `ctrl+p` no longer does anything anywhere — effort: 1.

## M3 — Tree state memory across Roots

Delivers: S3

- **M3.1** — Files Nav expansion state and selection are remembered per Root and restored when
  that Root becomes active again, correct even when the tree shape underneath has not changed
  (retires the C1 "accepted friction" Unscheduled entry) — done when: expanding/collapsing nodes
  and selecting one in Root A, switching to Root B and back to Root A, shows the same expansion
  and selection as left — effort: 3.

## M4 — Per-file git status markers

Delivers: S8

- **M4.1** — Files whose git status is modified or new are marked `M`/`A` respectively in the
  Files Nav on initial load, resolved relative to the containing repository's root; a Root
  outside any git repository shows no markers and raises no error; only these two states are
  distinguished (no staged/unstaged, deleted or renamed marking; dirty submodules and
  worktree/bare-`.git` setups are not handled) — done when: a Root inside a git repo shows
  `M`/`A` markers matching its actual working-tree status at launch, and a Root outside any
  repository shows none, silently — effort: 3.

## M5 — Manual refresh

Delivers: S2 (needs: M3.1, M4.1)

- **M5.1** — `R`, from any pane, re-reads the Active Root's file tree and the Active Document
  from disk on demand: Files Nav expansion and selection are preserved (selection moves to the
  nearest surviving ancestor if its node is gone), the Document Viewer's scroll position is
  preserved (clamped to the end if the document shrank past it), Find state may be lost, and
  S8's markers are recomputed in the same pass; if the Active Document no longer exists or is
  unreadable the Viewer becomes empty; nothing outside the Active Root's tree and the Active
  Document is touched — done when: creating, editing and deleting files outside noggin under
  the Active Root, then pressing `R`, shows every change in both panes without losing tree
  expansion, selection (or its nearest surviving ancestor) or reading position, with markers
  up to date; deleting the Active Document and refreshing leaves the Viewer empty — effort: 3.

## M6 — Dotfiles in the Pickers

Delivers: S4

- **M6.1** — Both Picker modes include dotfiles in their Results by default, matching what the
  Files Nav tree already shows; `.git/` stays excluded unconditionally; `.gitignore` handling
  is unchanged everywhere — done when: a dotfile visible in the Files Nav tree is findable by
  both File Picker and Grep Picker with no extra keystroke — effort: 1.
- **M6.2** — A session-scoped `alt+h` toggle hides and shows dotfiles, shared by both Picker
  modes, surviving the Picker being closed and reopened within the session but reset on
  relaunch — done when: `alt+h` hides dotfiles from both modes' Results, pressing it again
  restores them, and the hidden/shown state is unchanged after closing and reopening the
  Picker in the same session — effort: 2. (needs: M6.1)

## M7 — Peek

Delivers: S7

- **M7.1** — `p` in the Files Nav renders the selected document in the Document Viewer without
  making it the Active Document and without moving Focus, with an inverted header identifying
  the Peeked document; a Peeked document cannot be scrolled or searched with Find; `e` is a
  silent no-op while Peeking; `p` on the Active Document is a no-op — done when: pressing `p`
  on a non-Active file shows it with the Peek header while Focus stays in the Files Nav,
  attempting to scroll it, Find it, or Edit it does nothing, and `p` on the Active Document
  does nothing — effort: 3.
- **M7.2** — Peek is cleared by any selection change, any Focus change, opening the Picker, a
  second `p`, or a Refresh, restoring the Active Document at the scroll position it had before
  Peeking (or the empty Viewer if there was no Active Document) — done when: each of those five
  triggers, performed while Peeking, leaves the Viewer exactly as it was before Peek started —
  effort: 2. (needs: M7.1)

## M8 — Two-line Roots Nav rows

Delivers: S5

- **M8.1** — Each Roots Nav entry renders as two lines — base name in bold on line 1, path
  unbold on line 2 with the existing left-truncation treatment; active (`●`)/starred (`★`)
  markers and the `(missing)` broken-root suffix stay on line 1; both lines of the selected
  entry carry the selection highlight; movement steps one entry (two lines) at a time — done
  when: every Roots Nav entry shows name and path on separate lines, the selected entry's both
  lines are highlighted, and `j`/`k` move exactly one entry per keypress — effort: 2.

## M9 — Footer and Help Modal accuracy

Delivers: S9 (needs: M2.1, M2.2, M5.1, M6.2, M7.1)

- **M9.1** — Every binding added, changed or retired by S4/S6/S7 is reflected correctly in the
  Footer hints and the Help Modal for every pane it applies to — done when: `?` in each pane
  lists exactly that pane's current bindings and the Footer's contextual hints match, with no
  remaining reference to `ctrl+p` anywhere — effort: 2.

## M10 — Staleness indicator

Delivers: S10 — *pre-agreed cut: planned last, dropped without ceremony if the cycle tightens*
(needs: M5.1, for the "press `R`" message to mean anything)

- **M10.1** — A periodic (~5–10s) check reports in the UI that the Active Document or the
  Active Root's tree has changed on disk, without refreshing or re-walking anything, saying
  only that something changed and pointing at `R` — never what changed; must not perceptibly
  cost anything on a large Root — done when: an out-of-process change to the Active Document
  or under the Active Root's tree surfaces the indicator within one check interval, with no
  refresh, no tree walk, and no description of the change — effort: 2.
