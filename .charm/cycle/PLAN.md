# Plan — C3 compare

<!-- Per-cycle. The spec decomposed into milestones and tasks. Technically aware but
     high-level: state WHAT and the done-condition, not step-by-step HOW — trust executor
     judgment. Every task MUST have an acceptance criterion, machine-checkable where possible,
     and an effort estimate: a relative 1-5 gut-scoping (1 trivial, e.g. fixing a few typos;
     5 very high, e.g. large or complicated tasks) — not a rigorous prediction, just how the
     task looks at scoping time. Durable keys: milestones M1, M2...; tasks M1.1, M1.2...
     Never renumber; append or suffix (M1.2a) when inserting. Keys are scoped to this cycle —
     durable docs referencing them qualify with the cycle key (C3/M1.2). Note dependencies
     where ordering matters. -->

## M1 — Dual viewers, targeting and focus

Delivers: S1, S2, S3, S5, S6

- **M1.1** — Replace the model's single `viewer` with a `dv1`/`dv2` pair, add `dv1Pane`/
  `dv2Pane` to the pane enum, and introduce a focused-viewer accessor so every call site that
  currently assumes one `m.viewer` (scroll, Find, Edit, Peek, refresh) reads/writes through
  it instead — behavior-preserving for DV1 alone, DV2 not yet reachable. Delete
  `paneOrder`/`shiftFocus`/Tab-Shift+Tab bindings as part of the same change (S5 removes
  them; keeping them alive mid-refactor only to delete them next task is wasted motion) —
  done when: the full existing test suite passes with DV1 behaving identically to today, no
  `m.viewer` field remains, and Tab/Shift+Tab are no longer bound — effort: 4
- **M1.2** — Split the viewer region 50/50 in `layout.go`/`model.frame()` when DV2 exists;
  DV1 takes the full region when it doesn't (D54: no minimum-width guard, each half always
  gets its exact share) — done when: a layout/frame test with DV2 open shows two equal-width
  regions summing to the viewer width, and one with DV2 absent shows DV1 at the full width —
  effort: 3 (needs: M1.1)
- **M1.3** — Add direct focus-jump keys `r`/`f`/`d`/`shift+d` (Roots Nav/Files Nav/DV1/DV2),
  reachable from any non-modal Focus; `shift+d` is a no-op if DV2 doesn't exist — done when:
  a test drives each of the four keys from every other non-modal pane and lands Focus
  correctly, and confirms `shift+d` is inert with DV2 absent — effort: 2 (needs: M1.1)
- **M1.3a** — Rebind S5's focus-jump keys per D55 (supersedes M1.3's chords, keeps the
  behavior): FocusDV1 `d`→`o`, FocusDV2 `shift+d`/`D`→`ctrl+o` — update `keymap.go`'s
  bindings and doc comments and every M1.3 test/hint reference (`focus_test.go`,
  `frame_test.go`'s `allBindings()`, footer/help hint tests) to the new chords — done when:
  `make check`/`make check-pty` pass with `o`/`ctrl+o` driving every existing M1.3 test
  unchanged in behavior, and no stray reference to the old `d`/`shift+d`/`D` bindings remains
  — effort: 2 (needs: M1.3)
- **M1.4** — Files Nav's targeting modifier: `shift+enter` (bound together with the `L`
  fallback per D53) loads the selected file into DV2, creating it first if absent and moving
  Focus to DV2; `shift+p` peeks it into DV2, creating it first if absent, without moving
  Focus. Plain `enter`/`p` are unaffected and keep targeting DV1 — done when: a test from
  Files Nav confirms both modified chords create-and-target DV2 with the right Focus
  behavior, and plain `enter`/`p` still target DV1 unchanged — effort: 4 (needs: M1.2, M1.3)
- **M1.4a** — Rebind S3's targeting modifier per D55 (supersedes M1.4's chords, keeps the
  behavior): Files Nav's ActivateDV2 from `shift+enter`+`L` to `ctrl+enter`+`ctrl+l`, PeekDV2
  from `shift+p`/`P` to `ctrl+p` — update `keymap.go` bindings/doc comments and the M1.4 test
  suite (`target_test.go`, hint/help tests, `frame_test.go`'s `allBindings()`) to the new
  chords — done when: `make check`/`make check-pty` pass with `ctrl+enter`/`ctrl+l`/`ctrl+p`
  driving every existing M1.4 test unchanged in behavior, and no stray reference to
  `shift+enter`/`L`/`shift+p`/`P` remains — effort: 2 (needs: M1.4)
- **M1.5** — Picker's targeting modifier: `ctrl+enter` (+ `ctrl+l` fallback, per D55 — both
  non-printable, so `handlePickerInput`'s always-typing Prompt never swallows either as query
  text, closing TRIAGE T3 without a Picker-specific carve-out) on Activate loads the selected
  Result into DV2, creating it first if absent, and moves Focus to DV2; plain `enter` is
  unaffected — done when: a test activates a Picker Result with `ctrl+enter` and separately
  with `ctrl+l`, confirms DV2 is created/loaded/focused for both, confirms plain `enter`
  still targets DV1, and confirms typing a literal `l`/`L` into the Prompt still works —
  effort: 3 (needs: M1.4, M1.4a)
- **M1.6** — Verify/finish routing so each viewer's Find, Edit and scroll act only on its own
  Active Document once it holds Focus — full DV1/DV2 independence (S2, S6) — done when: a
  test loads distinct documents with distinct Find state into DV1 and DV2, drives Find/Edit/
  scroll against whichever holds Focus, and confirms the other viewer's document, Find state
  and scroll position are untouched, in both directions — effort: 3 (needs: M1.1, M1.4)

- **M1.6a** — Extend `resize()`'s find-dropped message (`"…; search cleared"`/`"…; search
  dismissed"`) to DV2 symmetrically, so a resize that drops DV2's committed or in-progress
  Find is surfaced the same way DV1's already is, and decide how two simultaneous drops
  combine into the single `m.message` slot — done when: a resize test with distinct Find
  state in both DV1 and DV2 shows a message naming both viewers' drops (or a documented
  combined wording), matching S2's independence guarantee at the messaging layer too —
  effort: 2 (needs: M1.6) (TRIAGE T5)
- **M1.7** — Extend `test/pty/harness.py`'s `KEY` table and `checks.py` with entries for the
  four C3 chords (`ctrl+enter`, `ctrl+l`, `ctrl+p`, `ctrl+o`) and at least one pty check
  driving `ctrl+enter` to create-and-focus DV2 and `ctrl+o` to focus it — done when:
  `make check-pty` exercises all four chords and fails if either regresses — effort: 2
  (needs: M1.3a, M1.4a, M1.5) (TRIAGE T6)

## M2 — Close DV2

Delivers: S4

- **M2.1** — Bind `ctrl+x` globally (any pane's Focus, not while Picker or Help Modal owns
  the keyboard) to close DV2: no-op if DV2 doesn't exist; Focus moves to DV1 if DV2 held it,
  otherwise Focus is unchanged — done when: a test closes DV2 from every non-modal Focus and
  confirms the Focus-landing rule, confirms it is a no-op with DV2 absent, and confirms it is
  inert while the Picker or Help Modal is open — effort: 2 (needs: M1.1, M1.3)
- **M2.1a** — Extend `test/pty/checks.py`/`harness.py`'s KEY table with real-binary coverage
  for `ctrl+x` (M1.7 covered the four targeting/focus chords that existed at the time but
  postdates CloseDV2) — done when: `make check-pty` drives `ctrl+x` against the real binary,
  confirms DV2 closes and Focus lands on DV1 when it held it, and fails if the binding
  regresses — effort: 1 (needs: M2.1, M1.7) (TRIAGE T10)
- **M2.1b** — `closeDV2` prunes DV2's entry from the staleness baseline immediately (not just
  on the next refresh), so a closed DV2's former document can no longer flip the combined
  staleness signal while it is no longer visible — done when: a test edits DV2's former
  document on disk immediately after closing it (no intervening Refresh) and confirms
  `m.stale` stays false — effort: 1 (needs: M2.1, M5.2) (TRIAGE T12)

## M3 — Refresh both viewers

Delivers: S7

- **M3.1** — Extend `refresh()` to re-read DV1's and DV2's Active Documents (whichever
  exist), not just one, alongside the existing tree re-walk, in the same single action,
  regardless of which pane holds Focus — done when: a test with both DV1 and DV2 open,
  documents changed on disk, confirms one Refresh updates the tree and both viewers'
  content, preserving each viewer's scroll position and the message-priority behavior the
  single-viewer case already has — effort: 3 (needs: M1.1)
- **M3.1a** — Fix `refresh()`'s message-combining switch (model.go:651-658) to guard against
  overwriting an existing read-failure message, the same way `refreshViewer`'s own branch
  and the pre-M3.1 single-viewer code already do (TRIAGE T11 — confirmed by charm-validate:
  a document that fails to re-read while it had a committed Find currently loses its
  read-failure message to the find-dropped message, contradicting both `refresh()`'s own
  doc comment and M3.1's acceptance clause); extract the combining logic shared with
  `resize()`'s near-identical switch into one helper both call (TRIAGE T13), so the two
  call sites can't drift independently — done when: a regression test with a document that
  fails to re-read on Refresh AND had a committed Find (single-viewer and dual-viewer cases)
  confirms the read-failure message wins over the find-dropped message, and `refresh()`/
  `resize()` both call the same combining helper — effort: 3 (needs: M1.6a, M3.1)

## M4 — DV2 is session-only

Delivers: S8

- **M4.1** — Confirm DV2's existence, Active Document and Peek/Find state are never written
  to any persisted store and never restored at launch — done when: a test opens DV2,
  restarts a fresh model against the same on-disk store, and confirms DV2 is absent while
  whatever DV1/Roots state is normally persisted is unaffected — effort: 1 (needs: M1.1)

## M5 — Staleness indicator: its own row

Delivers: S9

- **M5.1** — Relocate the staleness indicator from the Footer's appended hint suffix to its
  own dedicated row above the Footer (highlight color, inverse text), occupying no space
  while nothing is stale, same wording and "press R to refresh" resolution as today — done
  when: a layout/render test shows the row absent (zero height) when not stale and present
  with the existing message when stale, at a width too narrow for the old Footer suffix to
  ever have fit, and the old suffix-append code path is removed — effort: 3
- **M5.2** — Extend the staleness baseline/check to track DV1's and DV2's Active Documents
  (whichever exist) instead of one, still surfaced as the single combined signal — done
  when: a test makes only DV2's document stale (DV1 and the tree untouched) and confirms the
  row shows, and vice versa for DV1 — effort: 2 (needs: M1.1, M5.1)
