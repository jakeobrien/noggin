# Retired decisions — C1 read-across-roots

Retired at close (2026-08-02) per `.charm/DECISIONS.md`'s retention sweep: each entry's
subject either no longer exists, was scoped to a now-complete task with no ongoing
constraint, or is part of a superseded replan-triage chain retired intact. Moved here, never
deleted — see the `charm-close` report for the one-line reason each was retired.

## D8: Pre-agreed cut order if C1 overruns

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: Twelve deliverables against a one-week appetite.
- Decision: Cut in this order, most expendable first: S12 (help modal), S9 (navigator toggle),
  S8 (find-in-document).
- Why: Deciding the sacrifice while calm is better than triaging it on day five. Everything
  above those three is what makes the dogfood week possible at all.

## D11: Plan shape — walking skeleton first, cut candidates last

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: The repo is greenfield, and the spec's most valuable deliverables (roots, browsing)
  all presuppose a rendered document and a frame to put it in.
- Decision: M1 stands up a walking skeleton (module, frame, Document Viewer) that opens a
  document from a bare path argument; the real S1 launch contract lands in M3.1 and replaces
  that ad-hoc argument handling. The pre-agreed cut candidates (S8, S9, S12 per D8) are
  grouped in the final milestone in reverse cut order.
- Why: Every later milestone needs something on screen to attach to, and a temporary path
  argument is a few lines that S1 subsumes. Putting the cut candidates last means an overrun
  is a stop, not a re-plan.

## D12: Sandbox Go toolchain repair validated — workaround downgraded, not dropped

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: `findings/go-toolchain-env.md` documented a broken `GOROOT` requiring every Go
  command to run with it unset, no C compiler, and no gopls/staticcheck/goimports. The user
  reported repairing the container and asked for validation.
- Decision: An independent fresh-context validator confirmed all of it fixed — `go1.26.5` via
  a real `/home/agent/.local/go`, working gcc/cgo, and gopls/staticcheck/goimports all present
  and runnable, with a full build/vet/gofmt/test/module-fetch loop passing untouched. Updated
  the finding, its index entry, `PLAN.md`'s build-environment note and M1.1 wording, and
  `PROGRESS.md`'s handoff to reflect this. Kept the Makefile's `GOROOT` neutralisation as a
  defensive default for host portability rather than ripping it out, since it's free insurance
  and M1.1 isn't built yet.
- Why: This is an environment fact change, not a spec change — no plan tasks, keys or
  acceptance criteria needed re-scoping beyond wording. Kept the finding file rather than
  deleting it: it's the historical record of a diagnosis that changed once already, and the
  recipe-side recommendations it made (drop `GOROOT` from ENV, checksum verification, etc.)
  remain open and unverified from inside the container either way.

## D17: M1 code-review nits (T3-T5) get one small cleanup task, not three

- Status: decided
- Cycle: C1
- Date: 2026-08-01
- Context: The M1.1-M1.4 checkpoint review filed three related nits against already-`done`
  M1 code: test-only key-parsing helpers (`keyPress`, `modifiers`, `namedKeys`) plus a
  reachable `panic` live in `keymap.go`, a non-`_test.go` file, so they compile into the
  shipped binary (T3, medium); tests reach two levels into `bubbles/viewport.Model` internals
  instead of through noggin's own surface (T4, low); `pane.title()`'s abbreviations drift from
  `GLOSSARY.md`'s terms with no comment explaining why (T5, low).
- Decision: One new task, `M1.5`, covering all three — relocate the helpers to
  `keymap_test.go`, add small `viewer` accessors so tests stop reaching into the third-party
  model, and comment the glossary divergence. Scheduled ahead of `M2.1` since it is cheap
  (effort 2) and touches code already fresh in mind from the review.
- Why: T3 is medium (shipped-binary bloat plus a reachable panic in production code) so it
  gets a decision entry per the escalation matrix; T4 and T5 are low but sit in the same
  files and are one PR's worth of work, so bundling avoids three near-identical plan entries
  for one afternoon of cleanup.

## D19: M2.2 Files Nav — shared movement bindings, dedicated expand/collapse, expansion state layered over the immutable tree

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: M2.2 needed key bindings for the Files Nav pane and a place to hold "which folders
  are open / which row is selected", over `tree.Node` which is immutable data (M2.1 handoff).
  It also needed a way for the running binary to actually show a tree, ahead of M3.1's root
  resolution.
- Decision: (1) **Selection movement reuses the shared `LineUp`/`LineDown` bindings** (up/k,
  down/j) — focus gating routes the same physical key to the viewer's scroll or the Files
  pane's move, so no duplicate up/down bindings enter the keymap. (2) **Expand/collapse are
  the pane's own bindings** (`Expand` = right/l, `Collapse` = left/h) — a folder opens and
  closes in place; a file or already-open/closed folder is a no-op. Activation (open a file)
  is deliberately left to M2.3. (3) **Expansion/selection state lives in a `filesNav` struct**
  (`cmd/noggin/files.go`), keyed by absolute `Path`, value-semantic (expand/collapse clone the
  map) to match the type's immutable-update convention; `visibleRows` flattens the tree to its
  open part and the root node itself is not a row. (4) **`main.go` grows temporary dir-argument
  scaffolding** — a directory launches into the Files Nav, a file still opens the viewer — the
  same shape as the existing bare-path document scaffolding, and flagged in-comment as replaced
  wholesale by M3.1/M3.4.
- Why: all reversible and contained (medium) so logged, not asked. Reusing the movement keys
  goes with the grain of the existing focus-routing (D1). The dir scaffolding is the minimum to
  make M2.2 demonstrable end-to-end (pty) without pre-empting M3's root contract.

## D20: M3.1 resolveRoots is a pure decision over injected starred paths; rejection message parked on model until M4.2

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: M3.1 resolves the session's roots from the launch argument (S1), but its acceptance
  cases ("already a starred root", "subdirectory of a starred root", "parent of a starred root")
  presuppose a starred-roots concept that M3.2's persistent store has not built yet. The handoff
  also flagged that no footer-message mechanism exists (M4.2) for the rejections S1 mandates.
- Decision: (1) **`resolveRoots(args, starred []string, cwd) resolution` is a pure function** —
  the starred roots arrive as an injected parameter, not read from a store, so M3.1 needs no
  persistence and the table-driven test injects them directly; `main.go` passes `nil` until
  M3.2 wires the store in. (2) **It stats only the single argument** and returns the active root
  as a *path* — the side-effecting walk (`tree.Walk`) and document read stay in
  `model.applyResolution`, which downgrades their failures to messages too, never an exit (T7).
  (3) **Sub/parent detection is pure path arithmetic** (`filepath.Rel` on cleaned absolute
  paths), so a starred root need not exist on disk — matching S2's "missing paths are kept".
  (4) **The rejection message is parked on `model.message`** with display deferred to M4.2's
  footer; a `root{path, starred}` type and `model.{roots, activeRoot}` fields are introduced
  minimally for M3.2/M3.3 to build on.
- Why: medium and reversible, so logged not asked. The injected-starred seam is the smallest
  cut that lets M3.1 land and be fully tested before the store exists (go with the grain of the
  milestone order); parking the message honours "silence is not an error" without pre-empting
  M4.2's footer ownership.

## D24: C1's replan (2026-08-01) routes the M1-M3 review/validate backlog into three new tasks rather than a fourth cleanup-only milestone

- Status: decided
- Cycle: C1
- Date: 2026-08-01
- Context: TRIAGE.md had accumulated 19 open items (T9-T27) from the M1-M3 code reviews and
  the M3 validation pass — mostly low severity, three medium (T16, T17, T18). None invalidate
  the spec or plan shape; all are in-area fixes/hardening for work already built.
- Decision: split the backlog by locus and risk rather than one grab-bag task: **M2.4**
  (internal/tree correctness + test hardening: T11-T14, plus D23's case-sensitivity call),
  **M3.5** (roots/messaging correctness fixes where behavior diverges from what's communicated
  or from S1/S2's invariants: T17, T18, T20), and **M3.6** (the remaining standards nits, a
  real-but-mechanical duplication refactor, and one validate-sourced test-hardening item: T9,
  T10, T16, T19, T21-T25, T27) — mirroring D17's M1.5 precedent of one bundled cleanup task per
  batch rather than nineteen. T15 and T26 are folded into M4.2's existing acceptance criteria
  instead of new tasks, since both are silent-failure gaps in territory M4.2 already owns
  (footer messaging) and neither needs its own milestone slot. Total added effort: 7
  (M2.4: 2, M3.5: 2, M3.6: 3) plus M4.2's bump from 2 to 3.
- Why: medium and reversible — it grows the plan by 8 effort points against a cycle that
  already has pre-agreed cut candidates (D8) to absorb an overrun, so logged not asked. Keeping
  correctness fixes (M3.5) separate from pure standards/duplication cleanup (M3.6) and from the
  tree-specific batch (M2.4) keeps each task's acceptance criterion checkable in one sitting,
  rather than one criterion trying to cover unrelated files and risk levels.

## D25: M5.2 Navigator toggle binds to ctrl+n

- Status: decided
- Cycle: C1
- Date: 2026-08-02
- Context: S9/M5.2 names the action ("one key hides and restores the whole sidebar") but not
  the key. Every single-character mnemonic candidate was already taken (`n`/`N` by
  FindNext/FindPrevious, `s` by Star, `b`/`h`/`l` by scroll and Files Nav navigation); `ctrl+b`
  and `ctrl+f` are already PageUp/PageDown aliases.
- Decision: `ctrl+n` ("Navigator"), advertised in the footer as "sidebar" for every pane
  alongside Edit, the other global action already shown everywhere (footer.go's
  `paneHints`).
- Why: low risk and cheap to revisit — D2 confines every key literal to keymap.go, so
  retuning this one binding later is a one-file, one-line change; no plan or test depends on
  the specific key, only on the binding named `keys.Navigator` existing.

## D27: C1's replan (2026-08-02) routes the M5 review backlog into two new tasks, one deferral and one drop

- Status: decided
- Cycle: C1
- Date: 2026-08-02
- Context: TRIAGE.md had accumulated 9 open items (T28, carried over from the M3.5 review;
  T29-T36, from the M5.1-M5.3 fresh-context two-pass review) — three medium correctness/UX
  bugs (T28, T29, T34 by final judgment; T30 and T31 were reporter-rated medium but assessed
  lower here), the rest low standards nits or an informational dependency note. None
  invalidate the spec or plan shape.
- Decision: split by nature, mirroring D17/D24's precedent of one bundled cleanup task per
  batch rather than nine near-identical plan entries. **M5.4** (correctness/UX fixes, effort
  2): T28 (`resolveLaunch` drops an absolute-path launch argument on a `resolveCwd` failure
  it doesn't need — outside S1's enumerated rejections), T29 (`findMatches`' smart-case path
  can misattribute a match's line when a case fold changes a rune's byte length — the exact
  bug class `find.go`'s own doc comment says it exists to avoid), T34 (a resize mid-find-edit
  silently drops the query with no footer message, unlike every other lose-state path in the
  codebase). **M5.5** (standards cleanup, effort 1): T30 (13 literal `\u2014` escape-text
  occurrences instead of real em-dashes), T32 (stale future-tense help-modal comments now
  that M5.3 shipped), T33 (`viewer.find` field shares its name with its own type), T35 (an
  intent comment for FindNext/Previous's global dispatch). **Deferred** to `ROADMAP.md`
  Unscheduled: T31 (no shared exclusive-input-mode abstraction across modals) — real
  observation, but only bites if a third modal is ever added, and none is planned in C1.
  **Dropped**: T36 (M5.1's new indirect dependency) — already covered by `PROGRESS.md`'s
  existing dependency-graph-for-security-review note; folded the specific detail into that
  note rather than opening a separate item.
- Why: medium and reversible — it grows the plan by 3 effort points against a cycle that
  already has cut candidates (D8) to absorb overruns, though M5 (the last milestone, all cut
  candidates already built) has no further room to cut into, so logged not asked. T30 and
  T31 were downgraded from the reporter's medium rating on inspection: T30 is purely a
  comment-encoding inconsistency with no functional effect (cheap enough to fold into the
  standards batch rather than a dedicated task), and T31 has no live second occurrence to
  fix against — factoring an abstraction now would be speculative generalization for a
  modal count of two. T28, T29 and T34 earned their task because each is a genuine,
  testable behavior gap against what S1/S8/the confusion principle promise.

## D28: T28's fix distinguishes a directory argument from a file argument, not just absolute from relative

- Status: decided
- Cycle: C1
- Date: 2026-08-02
- Context: M5.4 built T28 (`resolveLaunch` dropping an absolute-path launch argument on a
  `resolveCwd` failure it does not need). D27's framing — "only a relative one still gets
  dropped" — holds for a directory argument (`resolveDir` never touches `cwd` once resolved)
  but not for a file argument: `resolveFile` always makes `cwd` the ad-hoc root (S1), even
  when the file argument itself is absolute, so dropping it on a `cwd` failure is still
  correct there. A literal absolute-vs-relative check would have reintroduced T20's bare-"`.`"
  bug for an absolute file argument.
- Decision: `resolveLaunch` gained a `needsCwd(args, cwd)` helper that stats the single
  argument (duplicating the stat `resolveRoots` does moments later) to tell directory from
  file before deciding: zero or >1 arguments never need `cwd`; a relative argument always
  does (T20); an absolute argument needs it only if it turns out to be a file. Locked in by
  `TestResolveLaunchKeepsAbsoluteDirArgOnUnresolvableCwd` (survives) and
  `TestResolveLaunchStillDropsAbsoluteFileArgOnUnresolvableCwd` (still drops) in
  `roots_test.go`.
- Why: low-risk, technical-correctness refinement of an already-accepted plan task, not a
  scope or spec change — resolved by judgment per the escalation matrix rather than a
  replan round-trip. The duplicate stat is a deliberate, cheap trade against threading
  `cwdErr` into `resolveRoots` itself, which would complicate a function whose doc comment
  already commits it to one pure, S1-shaped decision.

## D29: C1's replan (2026-08-03) routes the M5.4/M5.5-diff review backlog into two new tasks and one direct documentation fix

- Status: decided
- Cycle: C1
- Date: 2026-08-03
- Context: TRIAGE.md held 7 open items (T37-T43) from the 2026-08-02 fresh-context two-pass
  review of M5.4/M5.5's own diff — one medium (T37), the rest low. None invalidate the spec
  or plan shape.
- Decision: mirroring D24/D27's precedent of one bundled task per batch by nature rather than
  seven near-identical entries. **M5.6** (correctness, effort 2): T37 (`resolveLaunch`'s
  `needsCwd` and `resolveRoots` each stat the same launch argument independently — a narrow
  TOCTOU window that could reintroduce T20's bare-`.` bug; fix is to thread one stat through
  both rather than add a race test) and T43 (a *committed* find match, not just one still
  being edited, is silently dropped by any width-changing resize or Navigator toggle — the
  same "confusion is not allowed" gap T34 closed for the mid-edit case, left open for the
  post-Enter case; pre-existing since M5.1, not a regression). **M5.7** (standards/
  test-hardening, effort 1): T38 (`needsCwd`'s unused `cwd` parameter), T39 (`findMatches`'
  always-declare/sometimes-overwrite span shape), T40 (`resolveLaunch` test gap for
  `cwdErr != nil` at zero/many args), T41 (T29's regression test doesn't cover a match that
  itself spans a fold-length-changing rune, only one preceding it). **T42** (PLAN.md's M5.4
  bullet text stale against D28's more precise fix) is not a plan task at all — fixed directly
  by amending that bullet's prose in place, since the divergence is in the plan's own
  documentation, not in code.
- Why: medium and reversible — 3 effort points against a cycle with no further cut room
  (D27), so logged not asked. T37 earned medium severity's DECISIONS.md citation because it is
  a genuine, if narrow, correctness gap against a stated invariant; T43 is low severity but
  bundled with it because both are small, same-shaped "finish what T34/T28 started" fixes to
  the roots/find code T37 also touches. T38-T41 are cosmetic or test-coverage-only, no
  behavior change, same class as M5.5's batch. T42 skipped PLAN.md task machinery entirely:
  amending stale prose that already correctly cites the real fix (D28) is a documentation
  correction, not work requiring an acceptance criterion and effort estimate.

## D30: C1's replan (2026-08-03) routes the M5.6 validate/review follow-up (T44/T45) and one review test-gap (T46) into one new task; T44 dropped as superseded

- Status: decided
- Cycle: C1
- Date: 2026-08-03
- Context: TRIAGE.md held three open items from M5.6's validate and its fresh-context
  review: T44 (validate) flagged that M5.6/T43's own acceptance text names "a resize **or**
  Navigator toggle" but only the resize path has a dedicated test, and asked replan to decide
  whether a toggleNavigator-driven test is worth adding. T45 (review, medium) went further and
  proved by a temporary test that toggling the Navigator after committing a find sets
  `m.message = "window resized; search cleared"` even though no resize occurred — `resize`'s
  dismissed/cleared messages hardcode "window resized" regardless of whether the real trigger
  was a `tea.WindowSizeMsg` or `toggleNavigator`, so the footer is now factually wrong on
  exactly the path T44 asked about. T46 (review, low) found T37's stat-reuse refactor added an
  absolute-not-found-arg-plus-unresolvable-cwd branch in `resolveLaunch` with no test of its
  own, traced by hand to behave correctly.
- Decision: **T44 dropped as superseded** — T45's investigation answers T44's coverage
  question directly (the shared code path does fire, and does so incorrectly), so there is
  nothing left for a standalone toggleNavigator-driven test to prove once T45 lands; a
  correct-message test on that same path is T45's own acceptance criterion, not a second
  task. **New task M5.8** (`resize`'s message must not claim a resize happened when the
  actual cause was a Navigator toggle, T45; plus `resolveLaunch`'s missing
  absolute-not-found + unresolvable-cwd test case, T46) — bundled per the D17/D24/D27/D29
  precedent of one task per batch. Effort 2.
- Why: medium and reversible — 2 effort points against a cycle already past its cut room
  (D27, D29), so logged not asked. T45 earns the DECISIONS.md citation because a wrong footer
  message is a direct violation of the spec's own "confusion is not allowed" principle
  (SPEC.md's Principles) — worse than the silence it would otherwise be logging over. T46 is
  bundled alongside it rather than given its own task because it is a one-test, no-behavior-
  change addition in an adjacent file already touched by this cycle's roots.go correctness
  work (M5.6/M5.7), the same reasoning D29 used to bundle T37 and T43 together.
