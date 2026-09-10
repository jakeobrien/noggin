# Retired decisions — C6 sources

<!-- Retired at /charm-close (2026-09-10). Moved out of the durable DECISIONS.md, never
     deleted, per the doc-map's retention rule: subject gone, task-scoped with no ongoing
     constraint, or a superseded chain whose subject is gone. See the close report for
     one-line reasoning per entry. -->

## D94: Wikipedia is modelled as a kind of Root, not as a new tab/channel layer — but the model itself is deferred

- Status: decided — deferred model question resolved by C6/D114
- Cycle: C4a
- Date: 2026-08-14
- Context: The pitch was a top-row tab bar selected by number keys (`1` Notes, `2` Wikipedia).
  Against it: C1–C4 already provide almost the entire Wikipedia UI — Browse Modes Recent /
  Starred for a non-tree source (C4), the Picker's debounced async search (C2), the Document
  Viewer, DV2 side-by-side (C3), Document History (C4), Find, and a path-keyed Document Store
  whose key could widen to a URI so starred articles join starred notes in one collection.
  For it: a Wikipedia root looks identical to a Notes root yet diverges in ~6 places (no Tree,
  no `e`, no Grep, no git markers, no staleness, `R` means refetch), i.e. a permanent
  conditional-affordance tax that visible tabs would avoid. The dividing line was found to be
  "does it yield documents read in the Document Viewer?" — which chat fails and every other
  candidate source passes; with chat cut (D93), nothing left needs a channel layer.
- Decision: no tab/channel navigation paradigm. Wikipedia is a Root-like collection. The
  deeper question — whether a typed Root is just a different content source with conditional
  UX, or needs a more distinct layer with a defined shared/divergent boundary — is
  **deliberately left open** and scheduled as `C6 collections`.
- Why: the paradigm was speculative chrome for a two-item list, and building it would have
  answered the collections-model question in code before the question was ripe. The `tab` /
  `channel` terminology is dropped with it (`tab` is also already bound in the Files Nav and
  Picker, and C3 removed tab-cycling).

## D96: C4a ships no feature — it is a spike whose deliverable is a verdict

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: Four strategies were weighed for de-risking Wikipedia: a standalone TUI prototype,
  a temporary noggin "mode", roughing it in as a real Root while accepting destabilisation,
  or something thinner. The only genuinely uncertain thing is **conversion fidelity** — does a
  Wikipedia article survive REST HTML → cleanup → markdown and read like a document rather
  than soup. That risk has zero dependency on noggin's model, panes, picker or root typing.
- Decision: a throwaway command-line spike that writes `.md` files, judged by reading them in
  noggin itself via a scratch root. No integration, no search (exact title only), nothing
  reachable from the TUI. The cycle's deliverable is a go / no-go / retune-exhausted verdict
  plus findings; a **no-go is a successful outcome**.
- Why: writing files gets the real Glamour pipeline at real pane widths for free — a better
  fidelity test than a hand-written viewer, with none of the code. Search is known-solvable
  and proves nothing. In-place integration would have risked a half-integrated Wikipedia root
  trusted by nobody and deleted by no one, in a codebase four clean cycles and a pty harness
  deep. The conversion logic is model-agnostic and survives into whatever C6 decides.

## D97: The C4a spike code stays as-is, unmodified — kept as reference, not integrated

- Status: decided
- Cycle: C4a
- Date: 2026-09-03
- Context: SPEC OQ1, resolved at `/charm-close` with D112's GO verdict in hand. The named
  options were narrower than the situation warranted: keep the conversion and drop the
  command-line wrapper, or delete both and rebuild from the findings record.
- Decision: neither extreme. Keep the whole `cmd/wikispike/` package exactly as it shipped
  — conversion logic and CLI wrapper both — with no refactor. Do not split a reusable
  library interface out of it yet.
- Why: C6/collections (D94) has not yet decided what a Collection is or how Wikipedia would
  fit noggin's model; extracting a library shape now means guessing at an interface C6 might
  reject. The package is already isolated from noggin (D100 — retired with this cycle's
  narrower fix-up decisions, but the isolation property itself still holds in the code) and
  costs nothing left in place. Revisit extraction when a future cycle actually specs a real
  Wikipedia feature; until then this is exactly the model-agnostic keeper D96 anticipated,
  with `knowledge/wikipedia-conversion-recipe.md` (plus its maths/image-placeholder
  companions) as the reference for whoever picks it up.

## D121: C6 milestones are sequenced by the spec's pre-agreed cut order

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: SPEC "Pre-agreed cuts" ranks Recent-for-articles, Root names, then background tint
  as the sacrificial order on overrun. Recent and Root names also have dependants (S9/S10 fall
  back to base names; Recent rows share M4's fetch path).
- Decision: M6 (Recent), M7 (Root names) and M9 (tint) sit after the core Wikipedia slice
  (M1–M5) in cut order; M8's location line and Picker rows depend on M7's label helper but are
  written to degrade to the folder base name if M7 is cut. Product docs (M10) close the cycle.
- Why: an overrun then costs exactly the work the spec already agreed to lose, without a
  replan to untangle half-built dependants.

## D123: M2.1's byte-identical regeneration claim is unverifiable by a later fresh-context check

- Status: open
- Cycle: C6
- Date: 2026-09-08
- Context: M2.1's done-when ("regenerating one C4a sample through wikispike before and after
  the move yields identical markdown, recorded in the task's commit message") was satisfied at
  build time (commit `1d70bbd`: "byte-identical markdown, 16003 bytes, diff empty") but the
  "before" state (pre-move code) is not checked out anywhere afterward. A fresh-context
  validate pass re-ran `wikispike` against live Wikipedia post-move and got a matching byte
  count, which corroborates but does not reproduce the original before/after diff.
- Decision: **Open.** Not specific to M2.1 — any future task whose done-when is an "X and Y
  produce identical output" comparison should consider committing the diff artifact (or a
  checksum of both outputs) alongside the commit message, so later validation doesn't rely on
  re-fetching live content that happens to still match. No action taken on M2.1 itself, which
  stands as validated on the evidence available.
- Why: A process-quality gap in how equivalence claims are made checkable, not a defect in
  M2.1's actual result.

## D125: M2 review findings T4–T6 (medium) accepted as new hardening tasks, not a reopen of M2.1–M2.3

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: Replan triage of TRIAGE T4 (zero-value `Client` panics on `Fetch`), T5 (untested
  concurrency claim) and T6 (parameter shadowing in `fetchAt`/`fetchCanonicalHTML`/
  `canonicalFromURL`) — all medium severity, all against already-`done`/`Reviewed`/
  `Validated` M2.1–M2.3.
- Decision: Accepted as new tasks M2.4 (T4, T5) and M2.5 (T6, plus low-severity T7, T8)
  rather than reopening M2.1–M2.3's status — M2.3's own done-when criteria and tests still
  hold; these are additive hardening, not a correction of a failed acceptance criterion.
- Why: Keeps the Status table's Reviewed/Validated dates meaningful (they attest to the
  done-when criteria, not to every review nit) while still ensuring the findings turn into
  tracked work before M4.2 builds concurrent real-fetch usage on top of `Client`.

## D131: syncPreview's File-mode Starred-scope article placeholder stays a placeholder — Activate now fetches, but a Picker selection move must not (D119)

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: M3.5's own handoff flagged both of `activatePickerInto`'s and `syncPreview`'s
  `f.article` guards as "M4.2's job" to replace with real behaviour. This task replaces
  Activate's refusal with a real fetch, but D119 (cross-cutting: "every network call is
  user-initiated — search, Activate, Peek, Refresh") does not list a Results-list selection
  move among the triggers allowed to reach the network, and `syncPreview` runs on exactly that
  — every cursor move over the Picker's Results list, not a user's explicit request to open
  anything.
- Decision: `syncPreview`'s article placeholder is reworded ("no preview — Wikipedia articles
  are not previewed from here") but not replaced with a live fetched preview; only
  `activatePickerInto`'s refusal was replaced.
- Why: Fetching on every selection move would violate D119 and would also fetch far more often
  than necessary (every row the cursor passes over while browsing Starred, not just the one the
  user commits to) — the same "no Preview" contract S2 already gives wikipediaMode's own
  Results, extended consistently to File mode's Starred-scope article rows.
- Revisit: none expected.

## D132: TRIAGE T18's "pre-existing pty flakiness" diagnosis was wrong — `make check-pty`'s 11 failures are a real regression from M1.3/M1.6's own renames

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: T18 (filed during M3.1's build, severity low/build) reported 11 `make check-pty`
  failures, verified via `git stash` to reproduce identically with M3.1's changes removed, and
  concluded they "pre-date M3.1" and are "likely pty-harness/sandbox timing flakiness ... not
  root-caused." Replan re-ran `make check-pty` and read every failing assertion in
  `test/pty/checks.py` rather than accepting the filed diagnosis at face value (T18 itself
  asked for exactly this: "investigate before trusting check-pty as a milestone gate").
- Decision: The 11 failures are not flaky and not unexplained. They split into two
  deterministic groups, both caused by this cycle's own deliberate renames landing without a
  matching update to `test/pty/checks.py` (a cross-cycle regression suite, unlike the
  per-cycle model/render tests that were updated alongside M1.3/M1.6 themselves):
  (1) nine assertions (`M1.2` x2-at-two-sizes, `M3.2`, `M3.3`, `M3.4`, `M5.2`,
  `C2a/M2.1`) hard-code the pre-M1.6 "Roots" pane label / footer focus name, now "Sources";
  (2) two assertions (`M7.1`, `C2a/M6.2`) hard-code the File Picker's pre-M1.3 "Active Root"
  scope label, now universally "Active Source" for File mode (Grep alone keeps "Active Root",
  per M1.3's own acceptance criterion). `git stash`ing only M3.1's diff does not remove M1.3's
  or M1.6's own commits, so the "pre-dates M3.1" observation was correct but the conclusion
  drawn from it (flakiness) did not follow — the failures pre-date M3.1 because M1.3/M1.6
  pre-date M3.1 in build order, not because they are timing-dependent. Routed to a new PLAN
  task, M1.9, rather than re-filing T18 as-is or dropping it.
- Why: Leaving "flakiness, not root-caused" on record would have let the rest of the cycle
  keep landing on top of a silently-broken regression gate — most concretely, M9.1's own
  acceptance criterion names `make check-pty` directly, and a gate already carrying 11
  unexplained failures cannot distinguish M9.1's own tint regressions from this pre-existing
  noise. A five-minute rerun-and-read was cheap enough to do at replan time rather than defer
  the correction again.
- Revisit: none expected once M1.9 lands; if `make check-pty` still fails afterward, that is a
  new, distinct TRIAGE item, not a reopening of this one.

## D135: T26's location-line overflow bug is fixed by gating on boxHeight, not by making the line itself shrink further (C6/M8.2a)

- Status: decided
- Cycle: C6
- Date: 2026-09-09
- Context: Code review (TRIAGE T26, high) found `docContentHeight`/`peekContentHeight`
  (location.go:79-93) floor at `max(boxHeight-1, 0)` instead of suppressing the location line
  outright, so at a pathologically short terminal (`boxHeight` 0) `composeViewerContent` still
  appends the location row, overflowing the pane's height budget and corrupting the frame
  (DV1's bottom border disappears, panes misalign) — reproduced live via `sized(t, m, 80, 4)`.
- Decision: Follow `computeLayout`'s own `staleRow` precedent (`staleH := min(staleRowHeight,
  remaining)`, never overflowing) — gate location-line visibility on `boxHeight >= 1` rather
  than only flooring the reduced height at 0. Folded into the same task as T28 (dedup
  `docContentHeight`/`peekContentHeight`, same lines) since both touch the identical code.
- Why: The project already has an established, tested pattern for graceful degradation at
  this exact failure class; inventing a second (e.g. truncating the location line's own text)
  would add a new degradation shape for no benefit over reusing the existing one.

## D136: T29's SGR-7 reverse-video misdetection is fixed by positional SGR parsing, accepted as a real gap even though nothing currently triggers it (C6/M9.2)

- Status: decided
- Cycle: C6
- Date: 2026-09-09
- Context: Code review (TRIAGE T29, medium) found `tintPaneBackground`'s SGR-7 detector
  (style.go:120-137) matches raw `;`-split parameter tokens by bare string equality (`"7"`,
  `"27"`), so a numeric parameter of exactly 7 or 27 belonging to an extended-color spec
  (`48;5;7m`, `38;5;27m`, or any `38;2;`/`48;2;` truecolor component equal to 7 or 27) is
  misread as reverse-video set/clear rather than a colour value, silently dropping the S11
  tint for that run. Not hit by any current palette constant or test/pty check.
- Decision: Accept as a real, worth-fixing gap rather than a wontfix — parse SGR params
  positionally (skip the parameter(s) following a `38`/`48` token) instead of testing every
  token independently. Promoted to its own task (M9.2) rather than folded into M9.1, since
  M9.1 is already `done` and reviewed/validated.
- Why: Cheap to fix now while the code is fresh, and the failure mode (silent tint loss for a
  future glamour/style colour pick) would be hard to notice later without this note; deferring
  it would mean re-discovering the same root cause from scratch.

## D137: A build commit falsely stamped Reviewed dates on 7 unrelated tasks; corrected in PROGRESS.md, not treated as a PLAN task

- Status: decided
- Cycle: C6
- Date: 2026-09-10
- Context: TRIAGE T35 (review finding, filed high). The M1.12/M9.3 build commit (`d1ced0f`,
  2026-09-10) condensed PROGRESS.md's Handoff history and, in the same edit, stamped
  `Reviewed = 2026-09-10` on 7 unrelated already-`done` tasks (M1.11, M2.11, M7.4, M8.2a,
  M8.4, M9.2, M8.5) that had previously read `-`. No review of any of the 7 is recorded
  anywhere (commit message, this log, or `findings/`). The follow-up `review` commit
  (`5607d65`) legitimately reviewed and stamped only M1.12 and M9.3 — it did not touch or
  re-verify the other 7. Confirmed by reading both commits' diffs directly; the finding was
  accurate, not a false positive.
- Decision: Reset the 7 tasks' Reviewed cells in PROGRESS.md's Status table back to `-`
  rather than opening a PLAN.md task — this is a correction to the tracking doc's own
  accuracy, not code work with an acceptance criterion. All 7 (plus M1.12, M9.3, whose
  Validated cells were already `-`) now correctly show as awaiting a real, independent
  review/validate pass before `/charm-close`.
- Why: Left uncorrected, a future review/validate/close pass could have skipped these 7
  tasks on the strength of a stamp nobody actually earned — exactly the failure mode
  PROGRESS.md's own column rule ("stamped... for exactly the tasks that check actually
  covered, never assumed for a whole milestone") exists to prevent. The fix is mechanical
  and carries no ambiguity, so correcting it directly (rather than parking it as a task) is
  cheaper and clearer than the alternative.
