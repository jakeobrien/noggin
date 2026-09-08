# Triage — C6 sources

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
| T1 | review | low | internal/wikipedia/fetch.go:56; internal/wikipedia/fetch_test.go:234,263 | M2.8 changed `NewClient` but no test calls `NewClient()` directly — both zero-value/concurrency tests deliberately bypass it ("not NewClient()") and no caller in the repo invokes it yet; a future regression in `NewClient`/`ensureInit` would go uncaught. Add a test exercising `NewClient()` itself (e.g. `Fetch` against an httptest server, or asserting non-nil `httpClient`/`cache`). | open |
| T2 | review | low | cmd/noggin/docstore.go:157 (escapeDocField); cmd/noggin/docstore_test.go:276,343 | M1.7's and M1.8's tab/newline round-trip tests (`TestDocStoreArticleTitleTabAndNewlineRoundTrip`, `TestDocStoreKeyTabAndNewlineRoundTrip`) never cover a literal backslash in a title or key — the one character the escape scheme special-cases and reverses first (per the doc comment). Add a backslash case to lock in escape-order correctness rather than leaving it asserted only in prose. | open |
| T18 | build | low | test/pty/checks.py (11 failures incl. `M1.2 all three regions present`, `M7.1 the Picker opens in File mode...`, `C2a/M2.1 pty: r moves the footer's focus name to Roots`) | `make check-pty` has 11 pre-existing failures reproduced identically with M3.1's changes stashed out (verified via `git stash`), so they pre-date M3.1 and are not caused by it; `make check` (go test) is clean. Likely pty-harness/sandbox timing flakiness (assertions read a screen captured after a fixed `s.pump`/`s.send` delay) rather than a product regression, but not root-caused here. Re-run `make check-pty` on a clean checkout and, if still failing, investigate before trusting `check-pty` as a milestone gate. | open |
| T20 | review | low | .charm/GLOSSARY.md:103 (`Result`) | `Result`'s glossary definition ("a file (File Picker) or a file-and-line (Grep Picker)") predates M3.2/M3.3's Wikipedia Result and is now also the thing M3.4's Picker star acts on; two Wikipedia milestones have built on an un-updated glossary entry. Fold in a third case ("an article title+description (Wikipedia Picker)") before M3.5 adds more Wikipedia-Result behaviour on top of it. | open |
| T21 | review | low | cmd/noggin/picker.go:366-369 (`toggleHiddenFiles`) | `toggleHiddenFiles` unconditionally calls `queuePickerSearch` even when `p.mode == wikipediaMode`, where `hideDotfiles` has no bearing on Wikipedia results/rendering (no file path, no dotfile concept). Pressing alt+h while Wikipedia mode is open triggers a real, debounced (300ms) network search with no visible effect on the Results list, and this interaction is untested (`TestPickerAltHTogglesDotfilesInBothModesAndSurvivesReopen` only covers File/Grep). Recommend short-circuiting the requery when `p.mode == wikipediaMode` (still flip `m.hideDotfiles` for later File/Grep use, just skip `queuePickerSearch`), or adding a test that documents the current behaviour as intentional. | open |
| T22 | review | low | internal/wikipedia/search.go (`(*Client).Search`); cmd/noggin/picker.go:697 | No test calls the real `(*wikipedia.Client).Search` method directly — only the package-level `Search` function (`search_test.go`) and `cmd/noggin`'s `fakeWikipediaSearcher` (which never touches the real type) are exercised. Same gap-shape as already-open T1 (`NewClient` never called directly): a future regression in the one-line delegation (e.g. accidentally reading/writing a `Client` field, or drifting from the package function) would go uncaught. Add a test exercising `(*Client).Search` itself (e.g. against an `httptest` server via the existing overridable-base pattern `searchTitles`'s own tests use). | open |

<!-- empty as of 2026-09-08 replan: T15 -> PLAN M2.8, T17 -> PLAN M1.8; T1/T2 added by fresh-session code review (M1.8/M2.8) same day; T21/T22 added by fresh-session code review (M3.1-M3.5) 2026-09-08 -->
