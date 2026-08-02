# Plan — C1 read-across-roots

Greenfield: the repo has no code yet, so M1 stands up the walking skeleton. Milestone order
puts the pre-agreed cut candidates (S8, S9, S12 — D8) last, in reverse cut order, so an
overrun trims the tail instead of forcing a re-plan (D11).

Build environment: as of 2026-07-31 the sandbox's `GOROOT` resolves to a real, installed
go1.26.5 toolchain and `go`/`gcc`/`gopls`/`staticcheck`/`goimports` all work as inherited — the
GOROOT-unset workaround this note previously mandated is no longer required in this
container. M1.1's Makefile still neutralises `GOROOT` defensively (cheap insurance for a host
where it isn't set correctly), not because it is currently broken here. Checks remain `go vet`
+ `gofmt -l` + `go test`; staticcheck/goimports/gopls are now available too but not mandated.
See `findings/go-toolchain-env.md`.

Library baseline: Charm v2 (`charm.land/{bubbletea,bubbles,lipgloss,glamour}/v2`) as of M1.4.
M1.1-M1.3 were built against v1 and are migrated in place by that task; everything from M2.1
onward is v2-only. See D14 and `findings/bubbletea-v2-migration.md`.

## M1 — Walking skeleton and Document Viewer

Delivers: S7 (plus the frame every later milestone hangs off)

- **M1.1** — Scaffold the Go module and a Bubble Tea program that starts and exits cleanly —
  done when: `go build ./...`, `go vet ./...` and `gofmt -l .` are clean via a checked-in
  make/just target that neutralises `GOROOT` for host portability, and the binary runs and
  quits on `q` and `ctrl+c` restoring the terminal — effort: 2
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
- **M1.4** — Migrate the skeleton to Charm v2 (`charm.land/{bubbletea,bubbles,lipgloss,glamour}/v2`),
  one atomic swap (D14) — done when: no `github.com/charmbracelet/{bubbletea,bubbles,lipgloss,glamour}`
  import remains, `make check` is clean, and M1.1's, M1.2's and M1.3's acceptance criteria all
  still hold on v2 — including the frame at two terminal sizes, no key literal outside the
  keymap file, the fixture document rendered with frontmatter hidden and all five scroll
  actions working — with the pty harness re-verified against v2's startup queries and
  `expandTabs` either justified or dropped against lipgloss v2's own tab handling — effort: 3
  (needs: M1.3)
- **M1.5** — M1 code-review cleanup (T3-T5, D17): relocate the test-only key-parsing helpers
  out of production code, stop tests reaching into `bubbles/viewport` internals, and comment
  the intentional pane-title/glossary naming mismatch — done when: `keyPress`, `modifiers`,
  `namedKeys` and their panic live only in `keymap_test.go` (not compiled into the shipped
  binary), `viewer_test.go` asserts against noggin's own accessors (e.g. `offset()`,
  `atBottom()`, `lineCount()`) rather than `bubbles/viewport.Model` fields, and `pane.title()`
  carries a comment explaining its deliberate divergence from `GLOSSARY.md` — effort: 2
  (needs: M1.4)

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
- **M2.4** — internal/tree hardening (review findings T11-T14, replan C1/2026-08-01): align
  `policy.go`'s doc comment with the glossary term "Readable file" (T12), resolve `tree.go`'s
  dead non-directory branch in `Walk` — drop it or cover it with a test and correct its comment
  (T11), decide and implement case handling for the conventional extensionless names per D23
  (T13), and add a `.gitignore` escape-sequence test case (T14) — done when: `Walk`'s file-root
  branch is either removed (letting `os.ReadDir` fail naturally on a non-directory root) or
  covered by a test with its comment corrected; `isReadable` matches `README`/`LICENSE`/
  `CHANGELOG` case-insensitively with a locking test; `policy.go`'s doc comment says "Readable
  file"; and `TestIgnoreMatch` gains a case asserting `\!important.md` matches the literal name
  `!important.md` rather than being treated as a negation line — effort: 2 (needs: M2.1)

## M3 — Roots

Delivers: S1, S2, S3, S4

- **M3.1** — Resolve the session's roots from the launch argument per S1, every rejection
  producing a message and a session that still starts — **replaces**, rather than
  supplements, `main.go`'s pre-S1 bare-path scaffolding and its `os.Exit(1)`-on-bad-path
  behaviour (T7): a bad path must become a rejection message plus a session that still
  starts, per S1's "silence is not an error" principle, not a process exit — done when: a
  table-driven test covers bare launch, dir argument, file argument (cwd becomes the root,
  file becomes active), dir that is already a starred root (no duplicate, becomes active),
  non-existent path, subdirectory of a starred root (covering root becomes active), parent of
  a starred root, and more than one argument — effort: 3 (needs: M1.1)
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
- **M3.5** — S1/roots-store message and edge-case fixes (T17, T18, T20; replan C1/2026-08-01):
  three places where actual behavior diverges from what's communicated or from the roots
  invariants — done when: the too-many-arguments rejection message no longer implies
  `args[0]` was honored (T17); toggling a star with an unresolvable config directory either
  declines to flip the display or surfaces a message, matching the explicit-write-error branch
  (T18); and an `os.Getwd` failure at launch falls back to an absolute path (or surfaces a
  message) instead of feeding a bare `.` into root resolution (T20) — effort: 2
  (needs: M3.1, M3.2)
- **M3.6** — M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21, T22, T23, T24, T25, T27;
  replan C1/2026-08-01): the accumulated low/medium standards nits from the M1-M3 reviews plus
  one validate follow-up — done when: `keymap.go`'s header comment gains the "production file"
  qualifier (T9); `viewer.setOffset` is either a value-receiver `withOffset` or carries a
  comment explaining the deliberate exception (T10); `rootsNav`'s and `filesNav`'s
  near-duplicate `move`/`view` windowing logic is extracted into one shared helper (T16);
  `newRootsNav`'s unused `storePath` parameter is dropped or `withStorePath` is threaded
  through it (T19); a test exercises `activateRoot` clearing a stale `m.message` on success
  (T21); a table case covers `resolveFile`'s cwd-is-itself-a-starred-root dedup path (T22); the
  stray "(T: rare...)" annotation in `rootsnav.go` is replaced with plain prose (T23);
  `firstIndex` is renamed or inlined (T24); the four-file-duplicated `lipgloss.Color("245")` is
  centralized (T25); and M3's two ad-hoc validation scripts (starred-roots restart persistence,
  cross-root document-unchanged) become permanent checks in `test/pty` (T27) — effort: 3
  (needs: M2.1, M2.2, M2.3, M3.1, M3.2, M3.3, M3.4)

## M4 — Handoff and orientation

Delivers: S10, S11

- **M4.1** — Edit handoff: release the terminal to `$EDITOR` (fallback `vi`), resume on exit
  and re-render from disk (D3) — done when: with `EDITOR` set to a script that appends a line,
  the edit key returns to noggin showing the appended line, and with no document open the
  footer reports that instead — effort: 2 (needs: M1.3)
- **M4.2** — Footer: key hints for the focused pane, always including quit, truncating
  gracefully to advertise the help affordance; transient messages displace the hints and clear
  on the next keypress; also closes two carried-forward silent-failure gaps (T15, T26; replan
  C1/2026-08-01) — done when: hints change with focus between Roots Nav, Files Nav and viewer,
  a narrow terminal shows a truncated hint set still naming the help affordance, a rejected
  action's message clears on the next keypress, `activate()`'s file-read error surfaces a
  footer message instead of failing silently (T15), and an empty Roots Nav (no starred, no
  ad-hoc root) shows guidance satisfying SPEC.md's cycle-level "a first run with neither
  explains what to do next" (T26) — effort: 3 (needs: M1.2)

## M5 — Reading refinements

Delivers: S8, S9, S12. Cut candidates (D8) in reverse cut order — M5.3 goes first if the
cycle overruns, then M5.2, then M5.1.

- **M5.1** — Find in document: `/`-style incremental search, next/previous traversal that
  wraps, case-insensitive unless the query has uppercase, dismissal leaving the view put, and
  the matched line scrolled on screen and indicated — done when: a phrase in the last screen
  of a long fixture document is brought on screen and its line indicated, traversal wraps past
  the last match, an uppercase query stops matching lowercase, and dismissing leaves the
  viewport where the last match was — effort: 2 (re-scoped from 3 by D14: bubbles v2's
  `viewport` provides `SetHighlights`, `HighlightNext`/`Previous` and `EnsureVisible`, so go
  with the grain and use them rather than mapping offsets by hand) (needs: M1.3, M1.4)
- **M5.2** — Navigator toggle: one key hides and restores the whole sidebar, the viewer
  re-wrapping to the width it gains — done when: the key hides the sidebar, document text
  re-wraps to full width, and the key restores the previous layout with the sidebar's selection
  intact — effort: 1 (needs: M1.2)
- **M5.3** — Help modal listing the current context's keys, reachable from the footer
  affordance and dismissible with one key — done when: the modal opens over any pane, lists
  the bindings that pane's keymap actually defines, and one key closes it returning focus
  where it was — effort: 2 (needs: M4.2)
- **M5.4** — M5 review correctness/UX fixes (T28, T29, T34; replan C1/2026-08-02): three
  places where behavior diverges from what S1/S8 promise or fails silently against the
  "confusion is not allowed" principle — done when: `resolveLaunch` no longer drops an
  absolute-path *directory* launch argument on a `resolveCwd` failure it doesn't need — an
  absolute *file* argument still needs `cwd` and still gets dropped, per D28's more precise
  rule than this bullet's own original framing — with a regression test covering an
  absolute-path directory argument plus a failing cwd (T28); `findMatches`' smart-case path
  no longer misattributes a match to the wrong line when a rune's case fold changes its byte
  length, with a regression test using a length-changing fold (e.g. Kelvin sign U+212A)
  preceding a match (T29); and a `tea.WindowSizeMsg` arriving mid-edit of a find query
  surfaces a footer message when it resets the in-progress search, with a test covering the
  path (T34) — effort: 2 (needs: M5.1, M3.1, M3.5)
- **M5.5** — M5 review standards cleanup (T30, T32, T33, T35; replan C1/2026-08-02): four
  low-risk nits from the M5 fresh-context review — done when: all 13 literal `\u2014`
  escape-text occurrences in `find.go`/`navigator_test.go` are re-encoded to the real
  em-dash character (T30); `keymap.go`'s `Cancel`/`Help` doc comments no longer describe the
  help modal in future tense (T32); `viewer.find`'s field is renamed so it no longer shares
  its name with its own type (T33); and a one-line comment on `FindNext`/`FindPrevious`'s
  dispatch confirms they are intentionally global like `Find`, not scoped to viewer focus
  (T35) — effort: 1 (needs: M5.1, M5.2, M5.3)
- **M5.6** — Roots/find correctness fixes (T37, T43; replan C1/2026-08-03): two places where
  the code's stated invariant is technically racy or incomplete, both cheap to close — done
  when: `needsCwd`'s stat result is threaded through to `resolveRoots` (or `resolveLaunch`
  performs one stat and passes it to both), removing the independent second stat T37 flags as
  a TOCTOU window, with the existing table-driven tests still passing; and a resize or
  Navigator toggle arriving after a find match is committed (post-Enter, indicated via
  `applyCurrentMatch`) either preserves the match/indication or surfaces a footer message
  instead of silently discarding it (T43), extending T34's `findEditing()` resize guard to
  also check `hasFindMatches()`, with a test covering the committed-find resize path —
  effort: 2 (needs: M3.1, M3.5, M5.1, M5.4)
- **M5.7** — Review standards/test-hardening cleanup (T38, T39, T40, T41; replan
  C1/2026-08-03): four low-risk nits from the 2026-08-02 M5.4/M5.5 diff review — done when:
  `needsCwd`'s unused `cwd` parameter is dropped, or a comment documents why it's kept for
  signature symmetry with `resolveRoots` (T38); `findMatches`' always-declare/
  sometimes-overwrite `start, end := hStart, hEnd` shape is replaced with a small
  `origSpan`-style helper, or a comment justifies the current shape (T39); `resolveLaunch`
  gains test cases for a `cwdErr != nil` with zero launch arguments and with more than one,
  locking in `needsCwd`'s documented claims at the `resolveLaunch` level, not just
  `needsCwd`'s own (T40); and `findMatches`' fold-length-change regression test gains a case
  where the match itself spans a fold-length-changing rune (e.g. Kelvin sign U+212A), not
  just a fold preceding one (T41) — effort: 1 (needs: M5.4, M5.5)
