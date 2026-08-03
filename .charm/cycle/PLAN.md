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

## M2 — The Picker modal shell

Delivers: S1, S2, S8 (open state), S12 (messages), S14 (unavailable message)

Mode-agnostic: the shell is driven by whichever search M1 exposes. At the end of M2 the Picker
opens, types, moves, previews nothing useful yet, and dismisses cleanly.

- **M2.1** — Picker state plus open/dismiss, reachable from every Focus including while the
  Navigator is hidden, always opening as File Picker / Active Root / empty Prompt — done when: a
  test opens the Picker from each Focus and with the Navigator hidden, and a frame comparison
  shows the pre-open frame is restored byte-for-byte after dismissal (same Active Document,
  Active Root, scroll positions) — effort: 2 (needs: M1.1)
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
- **M3.2** — Preview pane renders the selected Result's file as raw, unrendered text from the
  top, scrollable, following selection changes — done when: a test shows the preview content
  matches the file's raw bytes (no Glamour styling), starts at line 1, scrolls on the scroll
  keys, and re-targets when the selection moves — effort: 2 (needs: M3.1)
- **M3.3** — Activate a File Result: it becomes the Active Document and nothing else moves —
  done when: after activating a Result belonging to a non-active Root, a test asserts the Active
  Document changed while the Active Root, the Files Nav tree's expansion and its selection are
  unchanged — effort: 2 (needs: M3.2)

## M4 — Grep Picker end to end

Delivers: S4, S9 (grep preview), S10 (matched line), S11

- **M4.1** — Grep mode wired to M1.4; each Result row shows file, line number and the matching
  line — done when: a fixture search renders one row per matching line across the Scope with the
  right line numbers, and long lines are truncated to the pane rather than wrapping the layout —
  effort: 2 (needs: M1.4, M2.5)
- **M4.2** — Preview opens scrolled to the matched line with that line highlighted — done when: a
  test on a match deep in a long file asserts the matched line is within the visible preview
  window and carries the highlight style — effort: 3 (needs: M4.1, M3.2)
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
