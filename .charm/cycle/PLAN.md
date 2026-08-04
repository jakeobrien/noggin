# Plan — C2 find

<!-- Per-cycle. The spec decomposed into milestones and tasks. Technically aware but
     high-level: state WHAT and the done-condition, not step-by-step HOW — trust executor
     judgment. Every task MUST have an acceptance criterion, machine-checkable where possible,
     and an effort estimate: a relative 1-5 gut-scoping (1 trivial, e.g. fixing a few typos;
     5 very high, e.g. large or complicated tasks) — not a rigorous prediction, just how the
     task looks at scoping time. Durable keys: milestones M1, M2...; tasks M1.1, M1.2...
     Never renumber; append or suffix (M1.2a) when inserting. Keys are scoped to this cycle —
     durable docs referencing them qualify with the cycle key (C3/M1.2). Note dependencies
     where ordering matters. -->

Shape: M1 builds the search layer headlessly, M2 the modal shell around it, M3/M4 make each
Picker mode work end to end, M5 makes them switchable in place. **M6 (S5 highlighting) is the
pre-agreed cut (D32)** — if the cycle runs long, skip it and go straight to M7.

Findings worth reading before M2: `findings/codebase-picker-integration.md` (modal dispatch,
compositing, find-state surface, and the two new-ground gaps: async `tea.Cmd` work and grep
line addressing — see D34).

## M1 — Search layer over ripgrep

Delivers: S3 (matching), S4 (searching), S6, S12 (cap), S13, S14 (detection)

Headless and unit-testable: no Bubble Tea, no UI. Everything here is a pure-ish function or a
subprocess call returning typed results/errors.

- **M1.1** — Detect `ripgrep` once at startup and carry its availability (and the reason it is
  unavailable) in session state, without affecting launch — done when: a run with `rg` absent
  from `PATH` still launches and opens a document, and a test asserts the availability flag and
  message for both the present and absent cases — effort: 1
- **M1.2** — Scoped, allowlist-filtered file listing via `rg --files`: Active Root or All Roots,
  Broken Roots silently skipped, results restricted to Readable files — done when: tests over a
  fixture tree show (a) only allowlist extensions/names returned, (b) `.gitignore`d and hidden
  files absent, (c) All Roots covers every live Root and omits a Broken one without erroring —
  effort: 3 (needs: M1.1)
- **M1.3** — Fuzzy-match and rank a query against **file names only**, capped, reporting whether
  the cap truncated the list — done when: tests show ranking puts a stronger name match above a
  weaker one, path components never influence a match, and a result set larger than the cap comes
  back capped with the truncation flag set — effort: 2 (needs: M1.2)
- **M1.4** — Literal (`--fixed-strings`) scoped grep with smart case, one result per matching
  line carrying file, line number and line text, capped with the same truncation flag —
  done when: tests show a query with regex metacharacters is matched literally, an all-lowercase
  query matches mixed case while a query with uppercase does not, only Readable files are
  searched, and the cap truncates — effort: 3 (needs: M1.2)
- **M1.4a** — Bound `grepFiles`' rg argument list (TRIAGE T1 / D43): batch the file list across
  multiple `rg` invocations (or otherwise cap argument size), merging results under the existing
  cap/truncation contract, so an All Roots search over a large tree can't silently pass ARG_MAX
  and fold a failed run into "no results" — done when: a test with a file count that would
  exceed a small injected argument-size limit still returns matches from files across every
  batch, and a genuine no-match still reports "no results" (not conflated with a failed run) —
  effort: 2 (needs: M1.4)
- **M1.4b** — Add the exact-cap/not-truncated boundary case to the grep cap test, matching the
  fuzzy-match test's existing coverage (TRIAGE T3 / D45) — done when:
  `TestGrepFilesCapsAndReportsTruncation`-shaped coverage includes a result set exactly at
  `pickerResultCap` asserting `truncated` is false, alongside the existing over-cap case —
  effort: 1 (needs: M1.4)
- **M1.4c** — Grep-batching test/safety hygiene (TRIAGE T18, T19): guard `grepArgMaxBytes`
  against a future parallel-test data race (a comment forbidding `t.Parallel()` on grep tests,
  or threading the limit as a parameter instead of a package var), and add the combined
  boundary case M1.4a/M1.4b each cover only in isolation — a result count landing exactly on
  `pickerResultCap` while spread across multiple `grepArgMaxBytes` batches — done when:
  `grepArgMaxBytes`'s test-only-mutation contract is documented or no longer a bare package
  var, and a test asserts `truncated` is false for a batch-spread result count exactly at
  `pickerResultCap` — effort: 2 (needs: M1.4a, M1.4b)

## M2 — The Picker modal shell

Delivers: S1, S2, S8 (open state), S12 (messages), S14 (unavailable message)

Mode-agnostic: the shell is driven by whichever search M1 exposes. At the end of M2 the Picker
opens, types, moves, previews nothing useful yet, and dismisses cleanly.

- **M2.1** — Picker state plus open/dismiss, reachable from every Focus including while the
  Navigator is hidden, always opening as File Picker / Active Root / empty Prompt — done when: a
  test opens the Picker from each Focus and with the Navigator hidden, and a frame comparison
  shows the pre-open frame is restored byte-for-byte after dismissal (same Active Document,
  Active Root, scroll positions) — effort: 2 (needs: M1.1)
- **M2.1a** — De-duplicate the Picker's open/dismiss reset-and-carry-preview-size sequence
  (TRIAGE T14 / D51): `openPicker` and `dismissPicker` (cmd/noggin/picker.go) repeat the
  identical "bump seq, snapshot preview size, replace with newPicker(), restore seq, restore
  preview size" sequence verbatim — the same one-path-forgot-the-carry-over shape D48/M3.2a
  already fixed once in one of the two paths. Extract a shared helper both callers use — done
  when: `openPicker` and `dismissPicker` share one reset helper, M2.1's and M3.2a's existing
  acceptance tests still pass unchanged, and no third reset path could add a preview-size
  carry-over without going through it — effort: 2 (needs: M2.1, M3.2a). Address before M4.x
  adds further Picker reset paths (grep preview / Find hand-off).
- **M2.2** — Modal layout and compositing: Prompt, Results list and Preview pane inside one
  modal composited over the live frame, clamped to the terminal — done when: the modal renders
  all three regions at several terminal sizes, never exceeds the terminal (test in the shape of
  `TestHelpOverlayNeverExceedsTheTerminal`), and the frame stays visible around its edges —
  effort: 3 (needs: M2.1)
- **M2.3** — Always-typing input dispatch: every printable key types into the Prompt; selection
  movement, Preview scrolling, Activate and Cancel are on non-typing keys; Footer and Help Modal
  cover the Picker's actions — done when: a test types a string containing characters that are
  bindings elsewhere (`j`, `g`, `/`) and finds them all in the Prompt with no side effects, and
  Footer/Help show the Picker's own actions while it is open — effort: 3 (needs: M2.2)
- **M2.3a** — Close the Help Modal gap left open by M2.3 (TRIAGE T5 / D46): list the Picker's
  own in-Picker actions (PickerUp/Down/Mode/Scope) in Help for the panes Help remains reachable
  from, so they aren't documented only in the footer's `pickerHintText` — done when: a test
  shows the Picker's own bindings appear in `helpBindings`' output (not just the opening key),
  and M2.3 is re-validated against its original acceptance criterion — effort: 2 (needs: M2.3)
- **M2.4** — Search dispatch as async work: the Prompt never blocks on a search, in-flight
  searches are debounced, and results arriving for a stale query are discarded — done when: a
  test drives fast successive queries and asserts the displayed results correspond to the final
  query only, and that a search command is issued fewer times than keystrokes typed — effort: 4
  (needs: M2.3; first async `tea.Cmd` work in the codebase, no in-repo precedent)
- **M2.5** — Status and empty states: the modal shows its mode and Scope, says so below the
  3-character threshold, says "no results", says the list is truncated at the cap, and reports
  Picker unavailability when `ripgrep` is missing — done when: tests assert the rendered text for
  each of the five states, including that no search is issued for a 1- or 2-character query —
  effort: 2 (needs: M2.4)

## M3 — File Picker end to end

Delivers: S3, S9 (file preview), S10 (activate)

- **M3.1** — Wire the File Picker to M1.3; each Result row shows the file's Root and its path
  relative to that Root — done when: a fixture with same-named files in two Roots renders two
  distinguishable rows, ordered by match strength — effort: 2 (needs: M1.3, M2.5)
- **M3.1a** — Disambiguate Root labels in File Picker rows when two Roots share a basename
  (TRIAGE T10 / D50): reuse `rootsNav.renderRow`'s left-truncated-parent disambiguation
  strategy in `pickerFileRow` instead of the bare `filepath.Base(f.root.path)` — done when: a
  fixture with two Roots sharing a basename but different parents renders distinguishable Root
  labels in File Picker rows, matching `rootsNav`'s own disambiguation behavior for the same
  case — effort: 2 (needs: M3.1)
- **M3.1b** — Fix `pickerRootLabel`'s double-counted gap budget and its test coverage (TRIAGE
  T13, T15, T16 / D52): the budget calculation subtracts the gap's width twice (once
  explicitly, once already embedded in the `rest` argument), under-computing the parent-path
  truncation budget by one gap width on every row and making M3.1a's disambiguation fix less
  effective than it should be at realistic terminal widths; the row test that should catch this
  derives its expected value from `pickerRootLabel` itself (tautological), and M3.1a's own
  disambiguation test runs at 140x30 rather than the project's standard 80x24, masking the bug
  — done when: the budget formula matches `name+gap+parent+gap+rest`'s actual rendered width
  (a test asserts `pickerFileRow`'s full rendered width equals the layout's content width), at
  least one test pins an independently-computed exact label string, and M3.1a's disambiguation
  test passes at the standard 80x24 fixture. Fold in TRIAGE T17 while touching this code:
  `pickerFileRow` builds `"  "+rel` twice (once as the `rest` arg, once concatenated after);
  compute it once into a local var and reuse both places — effort: 2 (needs: M3.1a)
- **M3.2** — Preview pane renders the selected Result's file as raw, unrendered text from the
  top, scrollable, following selection changes — done when: a test shows the preview content
  matches the file's raw bytes (no Glamour styling), starts at line 1, scrolls on the scroll
  keys, and re-targets when the selection moves — effort: 2 (needs: M3.1)
- **M3.2a** — Fix `dismissPicker`'s preview-size reset (TRIAGE T6 / D48): `dismissPicker` must
  preserve/restore the Preview viewport's Width/Height the same way `openPicker` carries them
  over, so opening the Picker a second time without an intervening `tea.WindowSizeMsg` doesn't
  render an empty 0x0 preview for the rest of the session — done when: a test drives
  open → dismiss → reopen with no resize in between and asserts the Preview pane still renders
  at its prior size — effort: 2 (needs: M3.2). Address before M4.2, which reuses this preview
  mechanism.
- **M3.2b** — Pin `syncPreview`'s error-content fallback string (TRIAGE T12): add a test
  asserting the exact `"preview unavailable: " + err.Error()` text, matching
  picker_view_test.go's existing convention of pinning user-visible strings — effort: 1
  (needs: M3.2)
- **M3.3** — Activate a File Result: it becomes the Active Document and nothing else moves —
  done when: after activating a Result belonging to a non-active Root, a test asserts the Active
  Document changed while the Active Root, the Files Nav tree's expansion and its selection are
  unchanged — effort: 2 (needs: M3.2)
- **M3.3a** — Add missing `activatePicker` test coverage (TRIAGE T7, T8): a read-error case
  (file removed/unreadable between listing and Activate, mirroring
  `TestActivateWithAReadErrorLeavesSessionRunning`'s shape) and a Grep-mode no-op case
  (Activate pressed with `mode = grepMode` asserts no state change) — done when: both cases have
  a passing test in picker_test.go — effort: 2 (needs: M3.3)
- **M3.3b** — Strengthen `TestPickerActivateOpensResultAsActiveDocumentOnly`'s expansion
  assertion (TRIAGE T9): assert `m.files.expanded` key/set equality, not just cardinality, so a
  same-count substitution would fail the test — effort: 1 (needs: M3.3)

## M4 — Grep Picker end to end

Delivers: S4, S9 (grep preview), S10 (matched line), S11

- **M4.1** — Grep mode wired to M1.4; each Result row shows file, line number and the matching
  line — done when: a fixture search renders one row per matching line across the Scope with the
  right line numbers, and long lines are truncated to the pane rather than wrapping the layout —
  effort: 2 (needs: M1.4, M2.5)
- **M4.2** — Preview opens scrolled to the matched line with that line highlighted — done when: a
  test on a match deep in a long file asserts the matched line is within the visible preview
  window and carries the highlight style — effort: 3 (needs: M4.1, M3.2)
- **M4.2a** — Grep row rendering hygiene (TRIAGE T21, T22 / D55): extract `pickerFileRow`'s and
  `pickerGrepRow`'s duplicated `.Width(width).MaxWidth(width).Render(line)` style tail into a
  shared `styleRow(line string, width int, selected bool) string` helper before M6 adds a
  third caller, and fix `TestPickerGrepResultsShowFileLineNumberAndText`'s tautological
  expected-string derivation (it currently builds its own expectations by calling
  `pickerGrepRow`) by adding at least one independently hand-computed row string plus an
  assertion that the match's line **text** is actually rendered — done when: `pickerFileRow`
  and `pickerGrepRow` both call one shared `styleRow` helper, and the grep row test has a
  hand-computed expected string and an assertion on rendered line text that a
  dropped-text bug would fail — effort: 2 (needs: M4.2). Address before M6, its anticipated
  third caller.
- **M4.3** — Activate a Grep Result: the file becomes the Active Document and the Document Viewer
  lands on the matched line — done when: a test activating a match deep in a long file asserts
  the Active Document changed and the viewport is scrolled such that the matched text is visible
  (see D34 for how the raw line is addressed in rendered content) — effort: 3 (needs: M4.2, M3.3)
- **M4.4** — Hand the grep query to the document's Find state as a Committed Find, with
  next/previous traversal over that query's other matches — done when: after activating a Grep
  Result in a file with several matches, a test asserts Find reports a committed query equal to
  the grep query and that FindNext/FindPrevious move between the document's matches — effort: 3
  (needs: M4.3)

## M5 — Mode and Scope switched in place

Delivers: S6, S7, S8

- **M5.1** — Switch mode (File ↔ Grep) inside the open Picker, carrying the typed Prompt over
  and re-running the search in the new mode — done when: a test types a query in File mode,
  switches to Grep, and asserts the Prompt is unchanged while the Results are grep results (and
  back again) — effort: 2 (needs: M4.1)
- **M5.2** — Switch Scope (Active Root ↔ All Roots) inside the open Picker in either mode,
  carrying the Prompt over — done when: a test with a match in a non-active Root finds it only
  after switching to All Roots, with the Prompt unchanged, and a Broken Root in the session
  produces no error row — effort: 2 (needs: M5.1)
- **M5.3** — No memory: mode, Scope and query are all discarded on dismissal, within a session
  and across sessions — done when: a test dismisses a Picker in Grep / All Roots with a query
  typed, reopens it, and finds File Picker / Active Root / empty Prompt; and no Picker state is
  written to the config or data store — effort: 1 (needs: M5.2)

## M6 — Match highlighting  <!-- pre-agreed cut, D32 -->

Delivers: S5

**This milestone is the cycle's pre-agreed cut (D32).** If time is short, mark both tasks
`deferred` and go to M7; nothing downstream depends on it.

- **M6.1** — Highlight matched characters within the file name in File Picker rows — done when: a
  test asserts the styled runs in a rendered row fall exactly on the fuzzy-matched positions of
  the file name and nowhere else — effort: 3 (needs: M5.3)
- **M6.2** — Highlight the matched text within the line in Grep Picker rows — done when: a test
  asserts the styled run covers the query occurrence in the rendered row, including a row whose
  line was truncated — effort: 3 (needs: M6.1)

## M7 — Hardening and documentation

Delivers: S14 (documented), cycle-level acceptance

- **M7.1** — Extend the pty check suite to cover the Picker against the real binary: open, type,
  results, preview, activate, dismiss, and the mode/Scope indicators — done when: `make
  check-pty` passes with new Picker checks, including one that asserts the frame after dismissal
  matches the frame before opening — effort: 2 (needs: M5.3)
- **M7.2** — Document the `ripgrep` requirement and Picker usage for users, and confirm the
  cycle-level acceptance walkthroughs by hand — done when: README (and `docs/` if used) states
  the requirement, its ignore-rule consequence and the keys, every command/example in it is
  verified against the binary, and each SPEC.md cycle-level acceptance bullet is confirmed or
  raised in TRIAGE.md — effort: 2 (needs: M7.1)
- **M7.1a** — Terminology and string-pinning housekeeping (TRIAGE T11, T24): standardize
  "Results region"/"Results column" comments (picker_view.go, picker.go) on GLOSSARY.md's
  fixed term "Results list" — done when: no comment in the Picker's code uses either variant —
  effort: 1 (needs: none; opportunistic, no dependency). T24 (review) flagged new M4.1/M4.2
  comments repeating the same variants; this task's criterion is codebase-wide and already
  covers them.
- **M7.1b** — Validation-pass test hygiene backfill (TRIAGE T1, T2, T3, T4, T23 / D56):
  (a) cross-reference M4.3's two paired ordinal-resolution tests in a comment so the coupling
  is explicit; (b) add a test asserting the Files Nav tree is unchanged after activating a
  Grep Result, mirroring M3.3's file-mode coverage; (c) isolate the disambiguation test's row
  assertion from full-frame comparison, or cross-reference M3.1b's row-isolated test as the
  coverage that actually isolates it; (d) give
  `TestPickerActivateOpensResultAsActiveDocumentOnly`'s expansion fixture at least one
  expanded subdirectory so its `reflect.DeepEqual` assertion exercises a non-empty map;
  (e) add a one-line comment on `TestGrepFilesBatchedResultsExactlyAtCapAreNotTruncated`
  noting it guards a future batching refactor, per D56 — done when: all five sub-items are
  addressed in their respective test files — effort: 2 (needs: none; opportunistic
  test-quality backfill from an earlier validation pass, no code dependency).
