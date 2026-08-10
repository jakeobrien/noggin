# Triage — C4 recall

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
| T1 | review | low | cmd/noggin/store_test.go:133 | `TestLoadStarredRootsNoConfigDir`'s name still says "NoConfigDir" though M1.1's clean break renamed the concept to data dir everywhere else (its own doc comment and body already say "data directory"/`XDG_DATA_HOME`); rename to `TestLoadStarredRootsNoDataDir` for consistency. | open |
| T2 | review | low | cmd/noggin/docstore.go:123 | `evictDocStore(records map[string]docRecord, cap int)` names its parameter `cap`, shadowing Go's predeclared `cap()` builtin inside the function body; rename to e.g. `limit` or `maxRecords`. | open |
| T3 | review | low | cmd/noggin/docstore.go:166 (cf. cmd/noggin/store.go:110) | `saveDocStore` duplicates `saveStore`'s `os.MkdirAll`-then-`os.WriteFile` shape near-verbatim; low priority now with only two stores, but worth a shared helper if a third store format is ever added. | open |
| T4 | build | medium | test/pty/checks.py:81-91,460-483 | `make check-pty`'s M3.3/M3.4 checks (Roots pane listing/marking/persistence) fail: `config_env` still builds the starred-roots store path under `$XDG_CONFIG_HOME/noggin/roots`, the location M1.1's clean break moved off of this cycle (now `dataDir()`/`storePath()`); the harness never got the matching update, so it starts every M3.3/M3.4 process with an empty store and then can't find the persisted file it expects. Confirmed pre-existing (same failures on the commit before M2.1/M2.2/M2.3, not caused by this cycle's work); `make check` (the actual gate) is unaffected and green. | open |
| T5 | build | low | cmd/noggin/footer.go:36 (cf. model.go:1092,1121) | M2.2's Document Viewer star toggle (`keys.Star`, bound in `Update`'s dv1Pane/dv2Pane case) and M2.1's Files Nav one are missing from `paneHints`/`helpBindings` for their own panes (filesPane's list has no `keys.Star` either) — both actions are real but undiscoverable from the Footer or Help Modal outside the Roots Nav. Pre-existing since M2.1/M2.2, only noticed while wiring M2.5's own Picker star discoverability (D69). | open |
| T6 | build | low | cmd/noggin/help.go:23-32 (D69) | M2.5's Picker-only star toggle (`keys.PickerStar`, alt+s) is absent from `helpBindings`'s three per-pane lists, unlike its Picker-only siblings (`PickerUp`/`PickerDown`/`PickerMode`/`PickerScope`/`ToggleHidden`) which are added everywhere as a Help-Modal-unreachable-while-Picker-is-open backstop — `PickerStar`'s own description ("star") collides with a pinned test (`TestHelpModalListsThePaneOwnBindings`) asserting plain `keys.Star`'s absence from the Files Nav/viewer Help Modal (see T5). The Footer's `pickerHintText` is its only discoverability surface today. Could be resolved together with T5 (e.g. giving `PickerStar` a more specific Help description, or revisiting what that pinned test actually pins). | open |
| T7 | review | high | cmd/noggin/model.go:553 (`toggleActiveDocStar`), cmd/noggin/picker.go:620 (`toggleStarPicker`) | Starring/un-starring a document from a focused Document Viewer (M2.2) or from the Picker (M2.5) updates `m.docs` but never calls `m.files.withStarred(m.filesStarredMarks())`, unlike `toggleDocStar` which does — so a document already visible in the Files Nav keeps a stale S3 marker until the next openFolder/refresh/Files-Nav star toggle. Reproduced: starring the Active Document from DV1 while the same file is a row in the Files Nav leaves that row unmarked. Fix: refresh `m.files.starred` (or derive it live from `m.docs` at render time instead of caching) from all three mutators of the star field, not just the Files-Nav-sourced one. | open |
| T8 | review | high | cmd/noggin/picker_view.go (`pickerFileRow`/`pickerGrepRow`/`pickerResultsContent`), cmd/noggin/picker.go:620 | S2 states un-starring from the Picker "leaves the Result in place with its star marker cleared", and M2.5's own PLAN acceptance criterion calls for confirming "the marker is gone" — but the Picker's Results list renders no star marker at all (`pickerFileRow`/`pickerGrepRow` take no starred/docs input), so nothing is ever shown to clear. `TestPickerStarTogglesSelectedFileResult` only asserts internal `docRecord`/`picker.results` state, silently substituting "internal Starred bool" for the spec's "visible marker". Note S3 itself never lists the Picker as a marker surface (only Files Nav rows and the Document Viewer title), so this may be a wording carry-over rather than a build gap — needs a call: add the marker to Picker rows, or amend S2/M2.5's language to stop describing a UI element that was deliberately left out. | open |
| T9 | review | low | cmd/noggin/picker.go:620 | `toggleStarPicker` breaks the `toggle{Target}Star` naming convention its siblings establish this cycle (`toggleDocStar`, `toggleActiveDocStar`); rename to `togglePickerStar` for consistency. | open |
