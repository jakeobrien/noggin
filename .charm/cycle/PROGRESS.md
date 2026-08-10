# Progress — C4 recall

<!-- Per-cycle. Cursor and statuses ONLY — no work summaries (git history carries those).
     The Handoff section is the single exception: overwrite it each session, never append.
     Statuses: todo | in-progress | done | blocked | deferred.
     Blocked rows reference the DECISIONS.md / TRIAGE.md item they wait on — the set of
     blocked rows is the cycle's blocking-issue list.

     Reviewed / Validated columns: date (yyyy-mm-dd) the task was last covered by code
     review / validate, or `-` if it never has been. Set per task key, by the role that
     performed the check — or the parent that dispatched it as a subagent — for exactly the
     tasks that check actually covered, never assumed for a whole milestone. This is what
     makes reviewed/validated state accurate across replans that insert tasks earlier,
     partial mid-milestone checks, and skipped end-of-milestone gates — a single cycle-wide
     checkpoint can't represent any of those. Clear a cell back to `-` when the task's code
     changes after that date (fix-ups, reopened work) — it needs re-covering. -->

## Cursor

Current task: M2.5

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.4 done — star indicator on the Document Viewer's title row. `model.paneTitle` (model.go) now
reads `m.docRecord(v.doc.path).Starred` for the Active Document (the same accessor
`toggleActiveDocStar` already uses) and, when set, prefixes the label with the same `"★ "`
glyph the Files Nav/Roots Nav rows use — `"Document: ★ a.md"` vs plain `"Document: a.md"`. Only
the Active-Document branch of `paneTitle` gets it; the Peek branch (`"Peek: " + v.peekLabel()`)
returns earlier and is untouched — S3 only speaks to the Active Document's own star, and Peek
names a different, non-Active document by design (S7/M7.1). No new row-layout ripple this time:
`paneTitle` returns a single string rendered through `renderBox`'s existing `Width`/wrap path, so
no fixed-column test elsewhere needed adjusting.
Evidence: `TestDocumentViewerTitleShowsStarIndicatorWithNavigatorHidden` (M2.4's acceptance
clause verbatim — Navigator hidden via `toggleNavigator`, DV1 and DV2 star and un-star
independently, `View().Content` checked for `"★ "` immediately left of each label) and
`TestDocumentViewerTitleStarIndicatorDegradesAtNarrowWidth` (resize to 24 cols, every rendered
line's `lipgloss.Width` ≤ terminal width, label still present) — both in new
`viewer_star_render_test.go`. `make check` green; `make check-pty` not re-run this session (no
code this task touched is in its path; TRIAGE T4's pre-existing M3.3/M3.4 failures are unrelated
Roots-store harness drift, not a Document Viewer concern).
Next: M2.5, star/un-star the Picker's selected Result (needs M1.4, done) — *pre-agreed cut 1,
drop first if the cycle spills*. Picker rendering/actions live in `picker.go`/`picker_view.go`;
`toggleDocStar`/`toggleActiveDocStar` (model.go) are the two existing callers of
`docRecord`/`withDocRecord` to pattern-match against, but the Picker's own selected-Result
accessor (not yet located this session) will be the new call site, and per S2 un-starring must
leave the Result row and cursor position unchanged — only its marker clears, no row removal here
(that's M4.3's Files Nav `Starred`-mode behavior, not the Picker's).

Gotchas surfaced while planning:
- `store.go` today is Roots-specific (`loadStore`/`saveStore`/`toggleStore`, plain one-path-
  per-line format). M1.2's document store needs a two-field record — give it its own file and
  names rather than overloading these.
- `frame_test.go`'s `TestEveryBindingIsCovered` will fail on any new binding (M4.1's two cycle
  keys, M6.1's two history keys) until the coverage list is updated.
- M5.1 changes the meaning of Picker scope: `liveRootsInScope` (`scope.go:26`) is Root-set-
  based, but `Starred` is a file set. Expect to refactor rather than add a branch.
- Gate is `make check`; `make check-pty` is the separate pty layer.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | -        | -         |                                          |
| M2.2 | done   | -        | -         |                                          |
| M2.3 | done   | -        | -         |                                          |
| M2.4 | done   | -        | -         |                                          |
| M2.5 | todo   | -        | -         | pre-agreed cut 1                         |
| M2.6 | todo   | -        | -         |                                          |
| M3.1 | todo   | -        | -         |                                          |
| M3.2 | todo   | -        | -         |                                          |
| M4.1 | todo   | -        | -         |                                          |
| M4.2 | todo   | -        | -         |                                          |
| M4.3 | todo   | -        | -         | `Orphans` scope is pre-agreed cut 3      |
| M4.4 | todo   | -        | -         |                                          |
| M4.5 | todo   | -        | -         |                                          |
| M4.6 | todo   | -        | -         |                                          |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
