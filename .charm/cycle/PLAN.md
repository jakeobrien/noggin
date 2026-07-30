# Plan — C1 read-across-roots

Greenfield: the repo has no code yet, so M1 stands up the walking skeleton. Milestone order
puts the pre-agreed cut candidates (S8, S9, S12 — D8) last, in reverse cut order, so an
overrun trims the tail instead of forcing a re-plan (D11).

Build environment: `GOROOT` is misset in this sandbox; use `GOROOT=/usr/lib/go` (go 1.26).
See `findings/go-toolchain-env.md`.

## M1 — Walking skeleton and Document Viewer

Delivers: S7 (plus the frame every later milestone hangs off)

- **M1.1** — Scaffold the Go module and a Bubble Tea program that starts and exits cleanly —
  done when: `go build ./...` and `go vet ./...` pass and the binary runs and quits on `q`
  and `ctrl+c` restoring the terminal — effort: 2
- **M1.2** — Lay out the frame: Navigator column (two stacked pane regions) beside the
  Document Viewer, Footer line beneath, keyboard focus moving between panes, reflowing on
  terminal resize; every binding resolved through one keymap definition (D2) — done when:
  the three regions render at their proportions at two terminal sizes, focus moves between
  panes and back, and `rg` finds no key literal outside the keymap file — effort: 3
- **M1.3** — Render the active document: Glamour-styled markdown wrapped to pane width,
  frontmatter pre-stripped (D5), viewport scrolling by line, half-page, page, top and bottom,
  and an empty-state message with no document open — done when: a fixture document containing
  frontmatter, headings, a table, a fenced code block, nested and task lists and a blockquote
  renders styled with no frontmatter visible, all five scroll actions move the viewport, and
  launching with no path shows the empty-state text — effort: 3 (needs: M1.2)

## M2 — Browsing the active root

Delivers: S5, S6

- **M2.1** — Readable-file policy and tree walk in one adjustable place: markdown-ish
  allowlist, `.gitignore` respected, `.git` excluded, dotfiles and hidden directories shown —
  done when: a unit test over a fixture tree returns exactly the expected file set — a
  gitignored `.md` excluded, `.git/` absent, `.charm/notes.md` present, `main.go` absent,
  `README` present — and the allowlist lives in one declaration — effort: 3
- **M2.2** — Files Nav pane: the active root as a tree opening with top-level entries
  collapsed, selection movement, expand/collapse, folders with no readable content still
  listed — done when: on a fixture tree the pane opens showing only top-level entries, a
  folder expands and collapses in place, selection moves across the expanded set, and an
  empty folder is still listed — effort: 3 (needs: M2.1)
- **M2.3** — Activating a file makes it the active document and moves focus to the viewer —
  done when: selecting a fixture file and pressing the activate key renders that file in the
  viewer with focus in the viewer, and scroll keys then act on it — effort: 1
  (needs: M1.3, M2.2)

## M3 — Roots

Delivers: S1, S2, S3, S4

- **M3.1** — Resolve the session's roots from the launch argument per S1, every rejection
  producing a message and a session that still starts — done when: a table-driven test covers
  bare launch, dir argument, file argument (cwd becomes the root, file becomes active), dir
  that is already a starred root (no duplicate, becomes active), non-existent path,
  subdirectory of a starred root (covering root becomes active), parent of a starred root,
  and more than one argument — effort: 3 (needs: M1.1)
- **M3.2** — Starred-roots store: one plain hand-editable file under the user's config
  directory (D7), tolerant load, star toggle written through — done when: a star survives a
  process restart, a missing file and a garbage file each yield an empty root list with no
  error, and a stored path that no longer exists is still loaded — effort: 2
- **M3.3** — Roots Nav pane: ad-hoc root first then starred roots, each entry showing base
  name plus left-truncated parent path, active highlighted, starred marked, missing paths
  marked broken and not activatable; move / activate / toggle-star actions — done when two
  same-named roots are visually distinguishable by their parent paths, the star mark toggles
  and persists, and activating a broken root is refused with a footer message — effort: 3
  (needs: M3.1, M3.2)
- **M3.4** — Activating a root re-roots the Files Nav, resetting expansion and selection,
  while the open document stays rendered and unchanged — done when: with a document open from
  root A, activating root B shows B's tree collapsed at top level and the viewer still shows
  A's document at its previous scroll position — effort: 2 (needs: M2.2, M3.3)

## M4 — Handoff and orientation

Delivers: S10, S11

- **M4.1** — Edit handoff: release the terminal to `$EDITOR` (fallback `vi`), resume on exit
  and re-render from disk (D3) — done when: with `EDITOR` set to a script that appends a line,
  the edit key returns to noggin showing the appended line, and with no document open the
  footer reports that instead — effort: 2 (needs: M1.3)
- **M4.2** — Footer: key hints for the focused pane, always including quit, truncating
  gracefully to advertise the help affordance; transient messages displace the hints and clear
  on the next keypress — done when: hints change with focus between Roots Nav, Files Nav and
  viewer, a narrow terminal shows a truncated hint set still naming the help affordance, and a
  rejected action's message clears on the next keypress — effort: 2 (needs: M1.2)

## M5 — Reading refinements

Delivers: S8, S9, S12. Cut candidates (D8) in reverse cut order — M5.3 goes first if the
cycle overruns, then M5.2, then M5.1.

- **M5.1** — Find in document: `/`-style incremental search, next/previous traversal that
  wraps, case-insensitive unless the query has uppercase, dismissal leaving the view put, and
  the matched line scrolled on screen and indicated — done when: a phrase in the last screen
  of a long fixture document is brought on screen and its line indicated, traversal wraps past
  the last match, an uppercase query stops matching lowercase, and dismissing leaves the
  viewport where the last match was — effort: 3 (needs: M1.3)
- **M5.2** — Navigator toggle: one key hides and restores the whole sidebar, the viewer
  re-wrapping to the width it gains — done when: the key hides the sidebar, document text
  re-wraps to full width, and the key restores the previous layout with the sidebar's selection
  intact — effort: 1 (needs: M1.2)
- **M5.3** — Help modal listing the current context's keys, reachable from the footer
  affordance and dismissible with one key — done when: the modal opens over any pane, lists
  the bindings that pane's keymap actually defines, and one key closes it returning focus
  where it was — effort: 2 (needs: M4.2)
