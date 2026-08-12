# Retired decisions — C4 recall

Retired at close (2026-08-13) per `.charm/DECISIONS.md`'s retention sweep: each entry's
subject either no longer exists, was scoped to a now-complete task with no ongoing
constraint, or is fully self-contained in the shipped code/tests it explains. Moved here,
never deleted — see the `charm-close` report for the one-line reason each was retired.

## D7: The starred-roots store lives in the config directory, hand-editable

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: It is arguably application data, not user configuration.
- Decision: Treat it as config for now, in a plain hand-editable format.
- Why: While iterating, being able to fix a moved or broken root in an editor is worth more
  than taxonomic correctness — it is also the only repair mechanism, since in-session root
  management is out of scope. Move it to a data location once it stops needing hand edits.

## D63: C4's pre-agreed cut list

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: C4 grew during brainstorming (new store, three star surfaces, three browse modes
  with per-mode scopes, per-viewer history, a new Picker scope, pane badges). C1 pre-agreed its
  cuts (D8) and needed none of them; the exercise was still worth it.
- Decision: if the cycle spills, cut in this order without re-litigation: (1) starring from the
  Picker, (2) per-document scroll memory in history, (3) the `Orphans` browse scope, (4)
  `Starred` scope on the Grep Picker. Recorded in SPEC.md's Pre-agreed cuts section.
- Why: each is a refinement of a deliverable that still stands without it — none removes a
  cycle-acceptance outcome outright. Ordered by how little is lost.

## D65: S10's "four" Footer focus hints are three in the code

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: S10 says the four focus-key hints (`r`/`f`/`o`/`ctrl+o`) are removed from the Footer.
  Only `FocusRoots`/`FocusFiles`/`FocusDV1` are actually listed there
  (`cmd/noggin/footer.go:36,38,44`); `FocusDV2` was deliberately never in the Footer (C3, see
  `footer.go:18`).
- Decision: M7.2 removes the three that exist. No spec amendment — the outcome S10 describes
  (no focus-key hints in the Footer) is unchanged.
- Why: a spec miscount with an identical end state is not worth a brainstorm amend cycle; the
  discrepancy is recorded here so a validator does not read three removals as incomplete work.

## D70: A stale Files Nav star marker (TRIAGE T7) is accepted work, not a design change — M2.7

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: code review found that `toggleActiveDocStar` (DV1/DV2, M2.2) and `toggleStarPicker`
  (Picker, M2.5) both mutate `m.docs` directly but never refresh `m.files.starred` the way
  `toggleDocStar` (Files Nav, M2.1) does — so a document already listed in the Files Nav shows
  a stale S3 marker until the next unrelated refresh. This contradicts S3's "star is visible"
  guarantee for two of its three star-mutating surfaces.
- Decision: accepted as a new task, M2.7 — refresh (or stop caching and derive live)
  `m.files.starred` from every mutator of the star field, not only the Files-Nav-sourced one.
  Not a Pre-agreed cut candidate; it directly undermines a delivered surface (S3), not an
  optional extra.
- Why: high severity (correctness bug in a delivered deliverable) but non-blocking (M3+ does
  not depend on it) — per the escalation matrix, accept and log rather than break flow, since
  the fix is small (effort 2) and well-scoped.

## D71: The pty harness's stale config-dir store path (TRIAGE T4) is accepted work — M1.1a

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: `make check-pty`'s M3.3/M3.4 checks (Roots pane listing/marking/persistence, from
  C3) fail because `test/pty/checks.py`'s `config_env` still builds the starred-Roots store
  path under `$XDG_CONFIG_HOME/noggin/roots` — the path M1.1's clean break moved off of, onto
  `dataDir()`/`storePath()`. The harness was never updated to match, so every M3.3/M3.4 process
  starts with an empty store and can't find the file it expects to persist. `make check` (the
  actual gate) is unaffected and green; this is `make check-pty` only.
- Decision: accepted as a new task, M1.1a — update `config_env` to resolve the store path the
  same way the shipped code does.
- Why: medium severity, non-blocking (the pty layer is a separate, already-known-partial gate
  per PROGRESS's gotchas), but caused by this cycle's own M1.1 change and cheap to fix
  (effort 2) — accept and log rather than leave a known-broken check sitting unfixed.

## D72: The Picker gets a visible star marker (TRIAGE T8) — S3 amended, built as M2.11

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: S2 says un-starring from the Picker "leaves the Result in place with its star
  marker cleared", and M2.5's own acceptance criterion talks about confirming "the marker is
  gone" — but the Picker's Results list rendered no star marker at all
  (`pickerFileRow`/`pickerGrepRow` took no starred/docs input); `TestPickerStarTogglesSelected-
  FileResult` only checked internal `docRecord`/`picker.results` state. S3, the deliverable
  that enumerates marker *surfaces*, never listed the Picker among them — only Files Nav rows
  and the Document Viewer title. Left open pending a product-surface call rather than assumed.
- Decision: option (a) — add a visible star marker to Picker rows, taken by direct user
  instruction this session. SPEC.md's S3 is amended in place (no key change; S3 already
  covered "star is visible" as a deliverable, this closes a surface it should have named from
  the start) to add "a star marker on Picker rows (File and Grep) for a starred Result," and
  its cycle-level acceptance line gains a matching bullet. Built immediately as **M2.11**
  (`cmd/noggin/picker_view.go`): both `pickerFileRow` and `pickerGrepRow` gain a `starred bool`
  parameter and render the same "\u2605 "/"  " two-column field Files Nav/Roots Nav already use,
  prepended ahead of the row's existing content and computed live from `m.docs` on every
  render (`pickerResultsContent` now takes a `docs map[string]docRecord`) — unlike the Files
  Nav's own cached `m.files.starred` (M2.7/D70), there is nothing here to go stale, since the
  Picker re-renders straight from the store every frame. Tests:
  `TestPickerFileRowShowsStarMarkerForAStarredResult`,
  `TestPickerGrepRowShowsStarMarkerForAStarredResult` (unit, marker present/absent, rest of
  the row byte-identical), and `TestPickerStarTogglingUpdatesTheRenderedMarkerLive`
  (integration: toggling `PickerStar` shows/clears the marker in the very next `View()`, no
  refresh needed). `make check` green.
- Why: high severity (spec/build mismatch on a delivered surface), decided directly rather
  than left open once the user weighed in — the escalation matrix's "high, non-blocking"
  routing to an open DECISIONS entry was the right call while unresolved, but a product-surface
  question is exactly what a direct user answer settles outright.

## D74: Picker Activate must also record `lastViewed` (TRIAGE T12) — accepted as M3.3

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: code review (TRIAGE T12) found that `activatePickerInto` and
  `activateGrepPickerInto` call `openDocumentInto` directly rather than through
  `activateInto`, so neither ever reaches `recordViewed` — a document opened via the File or
  Grep Picker never gets a `lastViewed` timestamp. GLOSSARY's Activate entry and the keymap's
  own doc comment already use "Activate" for both the Files Nav's and the Picker's commit
  action, landing either way in DV1 or DV2 — the same action S4 names as a recording trigger.
  D73 (M3.1) deliberately chose explicit call sites over centralizing in `openDocumentInto`,
  but only enumerated `activateInto`/`finishEdit`; it did not consider that the Picker's own
  activate functions bypass `activateInto` entirely, so the gap is a missed call site under
  D73's own chosen mechanism, not a reason to revisit it.
- Decision: accepted as a new task, M3.3 — add explicit `recordViewed` calls at
  `activatePickerInto`'s and `activateGrepPickerInto`'s own DV1/DV2 landing points, extending
  D73's explicit-call-site list to four rather than reversing D73's rejection of
  centralizing inside `openDocumentInto` (launch-time opens via `applyResolution` still do
  not record, unchanged from D73).
- Why: high severity — this undercuts a delivered deliverable (S4) and will silently starve
  M4.2's `Recent` mode of every Picker-opened document once built — but non-blocking for the
  cycle's current cursor, and the fix is small and well-scoped (effort 2). Per the escalation
  matrix: accept and log rather than break flow. M4.2's PLAN entry now also depends on M3.3
  so `Recent` mode isn't built and tested against a known-incomplete data source.

## D75: `recordViewed`'s `m.files.starred` recompute-on-every-write stays as is (TRIAGE T16)

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: code review (TRIAGE T16) noted `recordViewed` recomputes and reassigns
  `m.files.starred` via `filesStarredMarks()` on every recency-only write, even though only
  `LastViewed` changed — coupling a write path that has nothing to do with stars to Files Nav
  state, for currently-harmless cost bounded by `docStoreCap` (1000 records, D68).
- Decision: wontfix for this cycle. No PLAN task; the coupling stands as implemented.
- Why: low severity, speculative — the reviewer's own note frames it as "worth revisiting if
  this path gets hot," not a present correctness or performance problem, and any document
  store bounded at 1000 records makes a full recompute cheap regardless of trigger. Revisit
  only if a real cost is observed, not preemptively; logged so a future session doesn't
  reintroduce the same review finding without knowing it was already weighed.

## D76: `starMarker()`'s doc comment is corrected, not made true by refactor (TRIAGE T17) — M2.12

- Status: decided
- Cycle: C4
- Date: 2026-08-11
- Context: code review (TRIAGE T17) found M2.11's `starMarker()` doc comment claims the
  two-column marker field is "shared by every star-marking surface (files.go, rootsnav.go,
  model.go's Document Viewer title, and now Picker rows)", but only Picker rows actually call
  it — files.go and rootsnav.go still hand-roll the identical literal independently. Two fixes
  were open: refactor files.go/rootsnav.go to call `starMarker`, or reword the comment.
- Decision: reword the comment only, as a new task, M2.12. Not a refactor: files.go's and
  rootsnav.go's row renderers belong to M2.3 and the C2a/C3 Roots Nav work, both already
  `done`/Reviewed/Validated — touching them to chase a comment's accuracy would clear those
  cells back to `-` (doc-map's own rule) for a purely cosmetic gain, since the duplication
  itself causes no divergence today (both call sites render the exact same literal).
- Why: low risk × low-medium value — the finding is a doc-comment overclaim, not a functional
  bug; the cheapest true fix is the one that doesn't reopen already-verified work. Decided and
  logged rather than left open, since there is no real judgment call left once the reopening
  cost is weighed.

## D77: Picker-path `recordViewed` failure coverage extended, not centralized (TRIAGE T18) — M3.5

- Status: decided
- Cycle: C4
- Date: 2026-08-11
- Context: code review, re-confirmed by validate (TRIAGE T18), found M3.4's regression tests
  (`TestActivateStoreWriteFailureSurfacesMessage`, `TestFinishEditStoreWriteFailureSurfacesMessage`)
  cover only `activateInto`/`finishEdit`, not the two `recordViewed` call sites M3.3 added at
  `activatePickerInto`/`activateGrepPickerInto`. No functional bug today (neither Picker path
  sets a success message a store-write failure could clobber), but GLOSSARY's Activate entry
  and D74 both treat Picker commit as "Activate" too, so this is narrower than M3.4's stated
  goal and T13's original spirit.
- Decision: accepted as a new task, M3.5 — add a File Picker and a Grep Picker variant of the
  existing store-write-failure test, mirroring the two M3.4 already has, rather than reworking
  either M3.4's tests or `recordViewed` itself.
- Why: low-medium severity, non-blocking, cheap (effort 1) — regression insurance against a
  future success-message addition on the Picker path, not a present bug. Accept and log rather
  than leave a known coverage gap sitting unfixed, per the escalation matrix.

## D78: Grep-mode same-frame star-toggle render test added for M2.11 parity (TRIAGE T19) — M2.13

- Status: decided
- Cycle: C4
- Date: 2026-08-11
- Context: validate (TRIAGE T19) found M2.11's same-frame live-render toggle test
  (`TestPickerStarTogglingUpdatesTheRenderedMarkerLive`) exercises only File Picker mode; no
  equivalent test drives a real `m.View()` render through a `PickerStar` toggle in Grep mode.
  The underlying read (`docs[...].Starred`, computed fresh every render) is mode-agnostic, so a
  real bug is unlikely, but M2.11's acceptance wording explicitly says "both Picker modes" for
  the toggle clause.
- Decision: accepted as a new task, M2.13 — add a Grep-mode variant of the same-frame toggle
  test, mirroring the existing File Picker one.
- Why: low severity, cheap (effort 1), and closes M2.11's acceptance criterion literally rather
  than leaving it short of its own stated wording. Accept and log rather than drop, since the
  gap is real even though a bug is unlikely.

## D81: Files Nav mode/scope cycle keys get the M2.8 discoverability treatment — M4.8

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: build-flagged (TRIAGE T20) — `keys.FilesMode`/`keys.FilesScope` (M4.1) are absent
  from `filesPane`'s Footer hints and the Help Modal listing, unlike every other Files-Nav-only
  binding added alongside its own task (Peek, ctrl+ DV2-targeting) and unlike M2.8's own
  precedent for `keys.Star`.
- Decision: accepted as a new task, M4.8 — add both bindings to `filesPane`'s `paneHints` and
  the Help Modal listing, mirroring M2.8's shape exactly.
- Why: medium severity, cheap (effort 2), and there is already a same-cycle precedent for
  exactly this gap (M2.8) — no judgment call left once that precedent is applied. Accept and
  log rather than leave a second discoverability gap sitting unfixed.

## D83: Review-flagged test-coverage gaps bundled into one task — M4.10

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review flagged four coverage gaps with no accompanying functional bug: no
  test exercises `sortByRecency`'s or `starredRows`' path tie-break with two records sharing a
  timestamp/name (TRIAGE T26, medium); the `view()` height==1 branch has no dedicated test
  unlike every sibling branch (TRIAGE T25, medium); the orphan bucket's `recentPerRootCap` cap
  is untested past 100 orphan documents specifically (TRIAGE T29, low); and the Recent-mode
  render test doesn't assert the star marker itself, unlike Starred-mode's (TRIAGE T31, low).
- Decision: accepted as a single new task, M4.10, covering all four — each is a small,
  independent test addition with no shared implementation risk, so bundling keeps the plan
  from growing four tiny tasks for what is really one "close the coverage gaps review found"
  unit of work.
- Why: medium+low severity, no bug suspected in any of the four (reviewers explicitly called
  the underlying logic correct by inspection), but each is cheap and closes a real gap in
  logic specifically added to guard against nondeterminism or cap violations — worth accepting
  rather than leaving as latent regression risk with no test to catch it.

## D84: `toggleDocStar`'s inline flat-mode branch stays inline, not extracted for symmetry (TRIAGE T30)

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review (TRIAGE T30, low) noted `activateFlatRowInto`/`togglePeekFlatRowInto`
  are named helpers with their own doc comments, while `toggleDocStar`'s equivalent flat-mode
  branch is left inline — a minor structural inconsistency across an otherwise-symmetric trio.
- Decision: wontfix for this cycle. No PLAN task; the inline branch stands as implemented.
- Why: low severity, cosmetic-only — `toggleDocStar` and its M4.3 branch are already
  `done`/Reviewed; extracting a helper for symmetry alone would clear that Reviewed cell back
  to `-` for zero functional gain, the same reopening-cost reasoning D76 already applied to a
  comparable cosmetic finding. Revisit only if a future task needs to touch this branch for a
  real reason.

## D85: `clampSelected`'s double-scan on mode/scope/root transitions stays as is (TRIAGE T35)

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review (TRIAGE T35, low) noted `clampSelected` recomputes `rowCount()`/
  `flatRows()` on every mode/scope/root transition, and the following render recomputes the
  same thing again — compounding the codebase's existing "always recompute, never cache"
  design (D79) with an extra full doc-map scan per transition.
- Decision: wontfix for this cycle. No PLAN task; the double-scan stands as implemented.
- Why: low severity, speculative — the reviewer's own note frames it as low-risk, bounded by
  the 100-per-Root cap, with no observed cost. Identical shape to D75 (`m.files.starred`'s
  full recompute on every write): a bound already keeps the scan cheap regardless of trigger
  frequency. Revisit only if a real cost is observed, not preemptively.

## D86: M4.8's undertested Help Modal / footer hint additions get dedicated tests (TRIAGE T32, T33) — M4.12

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review found two coverage gaps in M4.8's discoverability work. T32 (medium):
  `TestHelpModalListsThePaneOwnBindings`'s `filesPane` case only substring-checks for
  `"mode"`/`"scope"` text, which `keys.PickerMode`/`keys.PickerScope` (already listed
  unconditionally for `filesPane`) render identically — the test cannot actually catch a
  regression that drops `keys.FilesMode`/`keys.FilesScope` from `filesPane`'s `helpBindings`.
  T33 (medium): no footer test asserts `keys.FilesMode`/`keys.FilesScope` actually appear in
  `filesPane`'s hint text at all — M4.8's footer touches only bumped a truncation width to
  keep unrelated strings fitting, a genuinely untested addition.
- Decision: accepted as a single new task, M4.12, covering both — same shape as D83's
  bundling of independent, low-shared-risk test additions into one plan task.
- Why: medium severity (a real hole in regression coverage for a delivered discoverability
  feature, M2.8/M4.8's precedent), cheap (effort 2), and the codebase already has an
  identity-based assertion pattern (`TestHelpBindingsListPickersOwnBindingsFromEveryPane`) and
  a footer-hint-isolation pattern (`TestPaneHintsNamePeekOnlyInFilesPane`) to mirror — no
  design judgment left, just applying existing precedent. Accept and log rather than leave a
  known regression-coverage gap unfixed.

## D87: `flatEmptyText`'s stale default-branch/comment mismatch and a tie-break test's ambiguous comment are fixed together (TRIAGE T34, T36) — M4.13

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: two low-severity review findings, both small doc/behavior clarity fixes with no
  shared implementation risk. T34: `flatEmptyText`'s doc comment (post-M4.9) calls its
  `default` case an unreachable defensive fallback, but that fallback still returns
  `emptyFilesText` ("no Root open") rather than `emptyTreeText` ("open but empty", the message
  M4.9 introduced and a Tree-mode caller would actually want there) — stale given M4.9's split.
  T36: `TestRecencyAndStarredNameTiesBreakByPath`'s tie-break assertion is only correct because
  subdirectory literals `"a"`/`"b"` sort lexicographically, not because of any documented
  guarantee — ambiguous for a future reader.
- Decision: accepted as a single new task, M4.13, covering both — correct `flatEmptyText`'s
  returned value or its comment (whichever stays true to M4.9's actual split) and add the
  one-line clarifying comment to the tie-break test.
- Why: unlike D76's comment-only precedent, both findings sit inside code from the same
  session's own work (M4.9, M4.10) rather than reopening a separately-`done`,
  already-Reviewed-and-Validated milestone — the reopening-cost reasoning that made D76 a
  comment-only fix and D84 a full wontfix doesn't apply here at the same weight. Cheap
  (effort 1) and closes a real (if low-severity) latent-bug risk in T34's case. Accept and log.

## D88: M4.7's acceptance wording gets its two missing regression cases (TRIAGE T37) — M4.14

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review found M4.7's acceptance criterion names four re-clamp call sites
  needing coverage (`cycleMode`, `cycleScope`, `setRoot`'s flat-mode memory restore, and
  un-starring the selected row out of Starred), but only two have dedicated regression tests
  in `flat_selection_clamp_test.go` (`cycleMode` and the un-star path). `cycleScope` shrinking
  the row set and `setRoot`'s memory-restore branch have no direct assertion on `f.selected`
  after the shrink. All four funnel through the same `clampSelected()` (D82), so a real bug is
  unlikely, but this is a literal gap against the task's own "done when" enumeration.
- Decision: accepted as a new task, M4.14 — add the two missing cases, mirroring the existing
  ones' shape.
- Why: low severity per the reviewer (shared implementation, low bug risk) but the task's own
  stated acceptance criterion is not literally met without them, and the fix is cheap
  (effort 2) — accept and log rather than leave a named acceptance-criterion clause
  unactually-tested.

## D89: M5.1 delivers the Grep Picker's own Starred scope too, despite PLAN.md naming it a pre-agreed cut

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: SPEC.md's pre-agreed-cuts list (#4) and PLAN.md's M5.1 note both name the Grep
  Picker's own `Starred` scope a candidate to drop first if the cycle spills. Implementing
  M5.1's shared `pickerScope` axis (one field read by both modes, S8's own "switching mode
  never rewrites the scope") and its candidate-set resolution (`starredPickerFiles`,
  `grepPickerFiles`) made Grep-mode search under Starred scope work for free — no separate
  code path, no meaningful extra risk or effort over File-mode alone. M5.1's own acceptance
  wording already asks for it literally ("a Grep and a File search under Starred return
  matches..."). This mirrors two earlier findings this cycle (PROGRESS.md's M2.5 "pre-agreed
  cut 1, not actually cut" and TRIAGE T40 on M4.3's Orphans/cut 3): a milestone noted as a
  contingency cut sometimes turns out cheap enough to just deliver.
- Decision: ship both modes' Starred scope in M5.1, not just File's. PLAN.md/SPEC.md's cut
  #4 wording is now stale in the same way cut #1's and #3's already were; leaving a
  PROGRESS.md note rather than editing SPEC.md/PLAN.md after the fact, consistent with how
  M2.5's own not-actually-cut case was recorded.
- Why: the acceptance criterion is the arbiter (per the build process), and it already
  demanded both modes; the cut existed only as spill insurance, and the cycle has not spilled.
  Deliberately crippling Grep-mode's search after already building the shared mechanism that
  makes it work would be *extra* engineering effort (a special-case exclusion), not less.

## D92: M4.13's tie-break-test comment stays at four lines, not trimmed to "one-line" (TRIAGE T39)

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review (TRIAGE T39, low) found M4.13's own acceptance wording
  (`PLAN.md` M4.13, D87) literally asked for "a one-line comment" on
  `TestRecencyAndStarredNameTiesBreakByPath`; what landed is four lines. The comment is
  accurate and matches the project's own norm of thorough narrative comments elsewhere in the
  same file; no functional issue.
- Decision: wontfix. No PLAN.md wording amendment, no code change — the delivered comment
  stands.
- Why: low severity, cosmetic literal-wording mismatch with no behavioral consequence.
  Trimming an already-accurate, already-`done`/Reviewed comment to satisfy a word count would
  cost more than it's worth, and rewriting M4.13's own acceptance wording after the fact to
  match what shipped sets a precedent (retroactively loosening acceptance criteria) worse than
  just noting the mismatch here, per the same standard this cycle already applies to itself
  (T37/M4.14's own literal-gap fixes were about missing coverage, not comment length).

## D93: `Orphans` scope (M4.3) was never actually cut, mirroring D89's M5.1 precedent (TRIAGE T40)

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: validate (TRIAGE T40) found `Orphans` scope — SPEC.md's pre-agreed cut #3, and
  PLAN.md's M4.3 note `*Orphans is pre-agreed cut 3.*` — is in fact fully implemented,
  reachable via `keys.FilesScope`, and covered by dedicated tests
  (`TestStarredRowsOrphansScopeOnlyShowsOrphans`, the six-combo empty-state test, browse-memory
  round-trip). This is the third such case this cycle: M2.5 (Picker starring, PROGRESS's own
  "pre-agreed cut 1, not actually cut" note) and M5.1 (Grep Picker's `Starred` scope, D89) both
  turned out cheap enough to deliver once the underlying mechanism was built for the
  non-cut sibling.
- Decision: leave SPEC.md/PLAN.md's cut-#3 wording as written (stale, like cut #1's and #4's
  already are) rather than editing it after the fact; record the actual outcome as a
  PROGRESS.md note on M4.3's row, consistent with how M2.5's and M5.1's own not-actually-cut
  cases were recorded. No PLAN task — there is nothing left to build.
- Why: the acceptance criterion is the arbiter and M4.3's own wording already named `Orphans`'
  scope test explicitly; the cut existed only as spill insurance the cycle never needed. Same
  reasoning as D89, applied to the third instance of the same pattern.

## D94: Dead `grepFiles` is deleted, not restored to a live call site (TRIAGE T42) — M5.2

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review (TRIAGE T42, medium) found M5.1's split of `grepFiles` into
  `grepPickerFiles` left `grepFiles` itself dead in production — `pickerSearchCmd` now
  resolves `files` itself (`starredPickerFiles` or `listFiles`) and calls `grepPickerFiles`
  directly, so `grepFiles` has no remaining production caller, kept alive only by
  `files_search_test.go`'s pre-existing tests. Its own doc comment claiming `pickerSearchCmd`
  is a caller is now factually wrong. Two options were open: delete `grepFiles` and repoint its
  tests at `grepPickerFiles(listFiles(...), query)` directly, or correct the doc comment and
  wire it back into `pickerSearchCmd`'s non-Starred branch as a thin roots-based wrapper.
- Decision: delete `grepFiles`, as a new task, M5.2. No future caller is named anywhere in
  SPEC.md/PLAN.md/ROADMAP.md that would need a roots-based wrapper distinct from
  `grepPickerFiles(listFiles(...), query)` — keeping dead code alive against a hypothetical
  future caller is the same shape of speculative cost D75/D85 already declined to pay for a
  recompute, applied here to a whole function instead.
- Why: medium severity (a doc-comment claim that is factually wrong about the codebase's own
  call graph is exactly the kind of drift replan exists to catch), cheap (effort 2, no
  behavior change since the function is already unreachable from production) — accept and log
  rather than leave a live doc-comment lie sitting in the tree.
