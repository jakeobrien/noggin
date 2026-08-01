# Triage — C1 read-across-roots

<!-- Per-cycle triage INBOX for actionable items surfaced out-of-band: code-review and
     validate findings, bugs / in-scope issues discovered during build that were not
     anticipated in the plan, work deliberately skipped per the escalation matrix (high,
     non-blocking, work-shaped), and the consequences of plan-invalidating discoveries.
     NOT a backlog and NOT a decision log (judgment questions go to DECISIONS.md).

     Only charm-replan parses items in full; every other role reads this inbox at summary
     level (count, severities, items referenced by blocked tasks).

     Lifecycle: consume-and-clear. Items are added `open`, then triaged by charm-replan per
     the escalation matrix — promoted to a PLAN.md task (with an acceptance criterion),
     deferred beyond this cycle into ROADMAP.md Unscheduled, recorded as a DECISIONS.md entry
     (wontfix), or dropped with a reason. Once triaged, the item is removed here (the outcome
     lives in PLAN / ROADMAP / DECISIONS). This doc stays small by construction; if it is
     growing, that means triage is overdue. It must be empty before a cycle can close.

     Keep one line per item. Ids (T1, T2…) are transient scratch labels, not durable keys —
     they are only for referring to an item until it is triaged, and may be reused after
     the inbox is cleared. Cite like findings: filepath:line. Severity is the reporter's
     estimate, not ground truth — triage filters false positives and wontfix.

     Sources: review | validate | build | docs | user. Status: open | triaged. -->

| Id | Source | Severity | Citation | Item | Status |
|----|--------|----------|----------|------|--------|
| T28 | review | medium | cmd/noggin/roots.go:64-72 | M3.5/T20's `resolveLaunch` drops *any* launch argument on a `resolveCwd` failure (`cwdErr != nil && len(args) > 0`), not just a relative one — an absolute-path argument needs no cwd at all (see `absPath`'s `filepath.IsAbs` guard) but is still discarded, surfacing an undocumented rejection ('path argument ignored') outside S1's enumerated rejection list; the doc comment on `resolveLaunch` itself only claims the relative case. No test covers an absolute-path argument combined with a cwd failure. Recommend gating the drop on the argument actually being relative, and adding a regression test for the absolute-path+cwdErr case. | open |
| T29 | review | medium | cmd/noggin/find.go:159-198 | `findMatches`' smart-case path lower-cases `haystack`/`needle` via `strings.ToLower` then reuses those byte offsets against the original `stripped` string; `ToLower` can change a rune's byte length (verified: U+212A Kelvin sign 3→1 bytes, U+0130 dotted-I 2→1 bytes), so a match after such a rune can be misattributed to the wrong line/column — the exact byte-offset-drift class of bug the file's own doc comment (find.go:13-26) says it exists to avoid. Recommend a regression test for a case-folding-sensitive rune before a match, and recomputing offsets against the folded string consistently (or avoiding the fold entirely for offset purposes). | open |
| T30 | review | medium | cmd/noggin/find.go:24,43,59,69,88,89,121,152,153,191,214,215; cmd/noggin/navigator_test.go:67 | 13 comments contain the literal 6-character escape text `\u2014` instead of a real UTF-8 em-dash (confirmed via `od -c`), unlike every sibling file touched this milestone (help.go, model.go, keymap.go, footer.go, layout.go, viewer.go, help_test.go all use the actual em-dash character). Recommend re-encoding all 13 occurrences to the real em-dash to match codebase prose style. | open |
| T31 | review | medium | cmd/noggin/model.go:340-349 | Exclusive-input-mode dispatch for modals (`findEditing`, then `helpOpen`, each hand-checked before the main key switch) is duplicated ad hoc per modal with no shared abstraction; a third modal repeats the same hand-ordered `if` before Quit, which is easy to get wrong silently. Recommend a single `m.inputCapture()` (or small interface) returning the active exclusive handler, if any. | open |
| T32 | review | low | cmd/noggin/keymap.go:71-75,83-87 | `Cancel` and `Help` field doc comments still describe M5.3's help modal in future tense ("since S12's help modal (M5.3) will need...", "even before S12's modal exists to open on it"), but M5.3 shipped in this same review scope — comments are now stale/misleading. Recommend updating both to state current reality. | open |
| T33 | review | low | cmd/noggin/find.go:27; cmd/noggin/viewer.go:30 | `viewer.find` field is typed `find` (field name identical to its own type name) — an avoidable naming collision/shadow risk. Recommend renaming the field (e.g. `search`) or the type. | open |
| T34 | review | low | cmd/noggin/viewer.go:64-72 | A `tea.WindowSizeMsg` resize arriving while a find query is mid-edit silently resets `find` to `newFind()` (query and editing state both lost) with no footer message, unlike every other lose-state path in this codebase (activate/edit/find-with-nothing-to-act-on), which messages explicitly per SPEC's "confusion is not allowed" principle. No test covers this path. Recommend a regression test asserting the intended behavior, and consider a footer message if silent drop is not intended. | open |
| T35 | review | low | cmd/noggin/model.go:366-375 | `FindNext`/`FindPrevious` (`n`/`N`) are dispatched above the per-pane focus switch, so they scroll/highlight the Document Viewer even when Roots Nav or Files Nav has focus; SPEC S8 describes find as scoped "over the open document" and the shared keymap.go:57-63 doc comment says `Find` is "global like Edit" but doesn't explicitly restate that for `FindNext`/`FindPrevious`. Likely intentional (mirrors `Find`'s global reach) but not explicitly confirmed in a comment. Recommend a one-line comment confirming intent, or gating on viewer focus if the narrower reading of S8 was intended. | open |
| T36 | review | low | go.mod:9-19 | M5.1 adds `github.com/charmbracelet/x/ansi` as a direct dependency and pulls in `github.com/atotto/clipboard` indirectly via `bubbles/v2/textinput` (clipboard itself unused). Not a scope violation, but flagging for the security-review dependency-graph awareness already noted in PROGRESS.md. | open |
