# Decisions

## D1: Go with the grain of the libraries

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: `NOTES.md` specified UI and behavior in unusual detail before any code existed.
- Decision: Where intent and library behavior differ, try the library's way first. Where a
  spec'd nicety would require authoring a subsystem to obtain it, drop the nicety and raise it
  for triage. The spec expresses intent; it is not a design to reproduce exactly.
- Why: The project only feels reachable because Glamour/Bubble Tea do the hard parts. Fighting
  them to match a pre-code design is how the budget disappears.

## D2: The keymap is provisional and build-time-only

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: `NOTES.md` contains a near-complete keymap, including collisions (`J` as
  follow-link vs. `shift+j` as accelerated movement) and an inverted `j`/`k`.
- Decision: Deliverables are defined by their *actions*, not their bindings. Bindings must be
  cheap to change at build time; user-facing key configuration is out of scope indefinitely.
- Why: Fluency is discovered by dogfooding, not designed up front. Naming keys in the contract
  would freeze guesses.

## D3: Editing is a suspend-and-resume handoff, not a tmux pane

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: `NOTES.md` offered "via tmux or handoff process" as if they were variants; they are
  different products.
- Decision: C1 releases the terminal to `$EDITOR` (fallback `vi`), resumes on exit, re-renders
  from disk. Reading while editing (tmux) is deferred to Unscheduled.
- Why: Bubble Tea provides process handoff natively, so this is nearly free and needs no file
  watching. Reading-while-editing only becomes meaningful once more than one document can be
  on screen (C3).

## D4: The Document Viewer has no cursor

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: The spec'd cursor with vim motions exists only to anchor Select mode and link
  jumping, both deferred.
- Decision: Viewport scrolling only. No cursor, no selection, no copy in C1.
- Why: Mapping cursor positions across styled, wrapped, rendered output is precisely where
  nvim's markdown plugins fail; paying that cost for no C1 feature is indefensible. Revisit
  when select mode or wiki-links are actually scheduled.

## D5: YAML frontmatter is hidden by pre-stripping

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: Glamour has no frontmatter support and CommonMark renders the closing `---` as a
  setext heading underline, so an unstripped block becomes a spurious H2.
- Decision: Strip the frontmatter bytes before rendering, as Glow does.
- Why: Doing nothing is not a neutral default, it is a visible bug. Stripping is ~10 lines and
  no new dependency; frontmatter-aware parsing would mean bypassing `TermRenderer` and
  re-implementing its options. Stripping for display does not prevent parsing tags later.
  See `cycle/findings/frontmatter-rendering.md`.

## D6: Roots may not nest

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: Roots bound visibility and search scope; overlapping roots make "which root is this
  document in" ambiguous.
- Decision: Roots are unique by exact path and may not be a parent or a child of another root.
  A path argument inside an existing root activates the covering root instead; a path argument
  that is a parent of an existing root is rejected with a message.
- Why: Ambiguous scope would poison C2 search results and C4 per-root document state. Accepted
  sharp edge: `noggin ~` is rejected when `~/notes` is starred.

## D9: Does noggin ever become a general text-file viewer?

- Status: open
- Cycle: C1
- Date: 2026-07-30
- Context: C1 restricts visible files to a markdown-ish allowlist, but the vision mentions
  wanting to view source and other text files; the two positions were in direct conflict
  during brainstorming.
- Decision: **Open.** C1 is markdown-ish only. Unresolved: whether general text/source viewing
  is a later deliverable or permanently out of character for the product.
- Why: In a code repo, source files vastly outnumber documents, so admitting them works
  directly against "find my notes fast" — the core value. Deferred until dogfooding shows
  whether the need is real.

## D10: Tag notation is unresolved

- Status: open
- Cycle: C1
- Date: 2026-07-30
- Context: `NOTES.md` lists frontmatter / free text / metadata as options for tags, with
  parsing and caching efficiency as the deciding factor. Tags are C5.
- Decision: **Open.** Not needed before C5.
- Why: The C1 decision to hide frontmatter (D5) does not prejudge this — hiding it from the
  render is independent of parsing it for metadata. Worth deciding before C5 is bet on.

## D13: The markdown render style is resolved once at startup, not per render

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: `glamour.WithAutoStyle()` re-resolves dark/light on every render by querying the
  terminal's background colour. Inside a running Bubble Tea program the reply is consumed by
  Bubble Tea's input reader, so the first render blocks forever and the program draws no frame
  (`findings/glamour-in-bubbletea.md`).
- Decision: Decide dark / light / no-TTY once, before Bubble Tea takes stdin, and pass Glamour
  an explicit `WithStandardStyle`. More generally: no terminal query happens after the program
  starts.
- Why: Going with the grain (D1) means using the library's default until the default is a hang.
  The cost is that a terminal which changes its background mid-session keeps the style it
  started with, which nobody will notice. Constrains M4.1: returning from `$EDITOR` must not
  re-query either.

## D15: The pty harness is a checked-in deliverable, not a scratch script

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: `findings/pty-verification.md` calls the pty harness "the recipe for every later
  end-to-end check", but the harness itself only ever existed in `/tmp`. M1.4 had to rebuild it
  from scratch — the second time — and rebuilding it is where the time goes, not writing the
  assertions: v2's cell-diffing renderer needs a ~90-line screen emulator (ECH, CUF, DCH, tabs)
  before a single row can be measured, and a DCH bug in it masqueraded as a product focus bug.
- Decision: the harness lives in the repo at `test/pty/{harness.py,checks.py}` and runs via
  `make check-pty`, kept separate from `make check` because it needs a pty and is slower.
  Key bytes in it are an explicit, documented D2 carve-out: one `KEY` table naming the binding
  each byte stands for, because the harness is on the far side of a language boundary from
  `keymap.go` and must put real bytes on the wire.
- Why: it is the only thing that verifies what actually reaches a terminal — colour, cell
  geometry, the startup handshake and the frame not overflowing the screen are all invisible to
  `go test`, which sees no TTY. A verification asset that has to be reinvented every milestone
  is not a verification asset. The cost is a Python dependency in the check path, which the
  sandbox already has, and a second place that knows about keys.

## D16: Keep `expandTabs` under v2, and keep a negative control on it

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: D14 left open whether lipgloss v2's own tab handling made `expandTabs` redundant.
  Tested rather than assumed: glamour v2 still emits literal tabs into code fences, and while
  lipgloss v2 does expand them inside `Render`, `lipgloss.Width` still measures a raw tab as one
  cell. The frame therefore grows a row taller than the terminal instead of overflowing the
  pane border — same defect class, new symptom.
- Decision: keep `expandTabs`, with `tabStop` documented as coinciding with lipgloss v2's
  default `TabWidth`. Add `TestFrameStaysWithinTheTerminalWhileScrolling`, which walks the whole
  document a page at a time and holds the frame to exactly the terminal's size — a single
  screenful cannot see this, the offending row has to be scrolled to. `make check-pty` asserts
  the same thing structurally (the top border must not scroll off).
- Why: removing the call is a cheap, repeatable falsification of both guards, and both fail when
  it is removed (`view height = 25, want 24`; "frame border intact while scrolling"). A defect
  that survives a library major and changes symptom is exactly the kind that needs a test whose
  failure mode is understood, not a comment asserting it is handled.

## D18: M2.1's walk hand-rolls .gitignore, lives in `internal/tree`, tests on a temp-dir fixture

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: M2.1's handoff flagged three early calls: how to honour `.gitignore` (dependency,
  shell-out, or hand-roll — "worth a findings note either way"); where the filesystem package
  lives; and how the fixture expresses "gitignored" without the git-commit awkwardness of a
  committed file matching its own fixture `.gitignore`.
- Decision: (1) **Hand-roll** a focused `.gitignore` matcher (`internal/tree/gitignore.go`) —
  shell-out (`git check-ignore`) only works inside a real worktree and consults the whole repo
  ignore hierarchy, wrong for arbitrary roots; a dependency adds weight to a tracked dep graph
  and neither candidate clearly beats a testable subset. (2) The walk lives in a new
  **`internal/tree`** package (not `cmd/noggin`'s `package main`): pure, heavily unit-tested
  logic that M2.2's Files Nav pane consumes. (3) The acceptance fixture is built
  **programmatically under `t.TempDir()`**, not committed — hermetic and it sidesteps git
  refusing to add a file its own fixture `.gitignore` ignores.
- Why: all three are medium (reversible, contained) so they are logged, not asked. The
  gitignore subset and its deliberate omissions are documented in
  `findings/gitignore-matching.md`; if a real root reveals a gap the fix is a test case plus a
  few lines, not a dependency swap.

## D21: M3.2 starred-roots store is a line-based text file, garbage filtered by absolute-path shape, toggle writes through

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: M3.2 needed a concrete on-disk format for S2's "one plain, hand-editable file" that
  is tolerant of a missing or garbage file (empty list, no error) yet still round-trips a star
  across a process restart, and does not prune paths that no longer exist on disk.
- Decision: (1) **Format is one absolute path per line**, blank lines and `#`-prefixed comments
  ignored — the same shape as the `.gitignore` files noggin already parses, so it stays
  comfortable to hand-edit and repair. (2) **"Garbage" is defined structurally, not by trying
  and failing to parse**: `loadStore` keeps only lines that are non-blank, non-comment, and pass
  `filepath.IsAbs`; anything else (binary noise, relative paths, stray text) is silently
  dropped rather than surfaced as an error, which is what makes a garbage file and a missing
  file collapse to the same "empty list" outcome. (3) **No stat on load** — a stored path that
  no longer exists on disk is returned unpruned, matching S2 and mirroring D20's sub/parent
  detection not requiring the path to exist. (4) **`toggleStore` writes the whole list through
  immediately** on every star/unstar (load, flip membership, save) rather than batching, so a
  restart-survival test is just "call toggleStore, then loadStore again". (5) `storePath()`
  isolates `os.UserConfigDir()` (D7) from the testable `loadStore`/`saveStore`/`toggleStore`
  functions, which take an explicit path — tests never touch the real config directory.
  `main.go`'s `loadStarredRoots()` folds a non-resolvable config directory into the same empty
  result.
- Why: medium and reversible (a file format with no consumers yet outside this store), so
  logged not asked. Filtering by path shape rather than a stricter format (e.g. JSON) keeps the
  file trivially hand-editable per S2's explicit intent, at the cost of silently dropping a
  malformed line instead of reporting exactly which line was bad — acceptable since S2 asks for
  tolerance, not diagnostics.

## D22: M3.3 Roots Nav pulls forward a minimal footer message, ahead of M4.2

- Status: decided — Picker parent-path disambiguation dropped by C6/D117
- Cycle: C1
- Date: 2026-07-31
- Context: D20 parked S1's launch-rejection message on `model.message` with display deferred to
  M4.2's footer work, and T15 deliberately left M2.3's file-read-error path silent for the same
  reason. But M3.3's own acceptance criterion is explicit: "activating a broken root is refused
  with a footer message" — S3 names the refusal, not just the internal state.
- Decision: (1) **`renderFooter` now shows `m.message` when set**, falling back to the existing
  focus+quit hint otherwise — the minimum needed to surface a refusal, not M4.2's full design
  (focus-aware key hints, narrow-terminal truncation, clear-on-next-keypress all remain M4.2's
  job). (2) As a side effect, S1's launch-time rejection message (parked since D20) is now also
  visible on the first frame, since it uses the same field — a welcome side effect, not scope
  creep: the field existed for exactly this. (3) **`activateRoot` clears `m.message` on success**
  so a stale refusal does not linger once the user corrects it, but nothing else clears it yet
  (e.g. switching panes) — that generalized behaviour is M4.2's "clears on next keypress".
  (4) Roots Nav rendering: an active-root marker (`● `) and a starred marker (`★ `) as two-column
  prefixes (mirrors the Files Nav's marker column), a broken root gets a literal ` (missing)`
  suffix and a faint style rather than a symbol, and same-named roots are told apart by a
  left-truncated parent path (`leftTruncate`, `…/work/admin/`) — the exact form S3 asks for.
  (5) **Activate and Star are new/reused key bindings on the Roots Nav**: Activate (`enter`) is
  reused from the Files Nav — both mean "open the selected thing" — and a new `Star` binding
  (`s`) is added for the toggle, with cross-cutting effects (tree walk, disk I/O) handled in
  `model.activateRoot`/`model.toggleStar` rather than `rootsNav.handleKey`, mirroring how the
  Files Nav's own Activate is handled one level up (M2.3). (6) `model.roots` becomes the
  `rootsNav` pane itself (selection cursor plus rows), replacing the placeholder
  `model.{roots, activeRoot}` fields D20 introduced — nothing outside `model.go` referenced them.
- Why: medium and reversible, so logged not asked. Showing the message now is what the plan
  task's acceptance criterion requires; building M4.2's full footer redesign to get there would
  be scope creep in the other direction. The marker/key choices are cosmetic and cheap to retune
  under the provisional-keymap principle (D2's build-time-only constraint covers `Star` too).

## D23: Conventional extensionless document names match case-insensitively, like extensions

- Status: decided
- Cycle: C1
- Date: 2026-08-01
- Context: M2.1 review (T11-T14, TRIAGE.md) found `isReadable` (internal/tree/policy.go)
  lower-cases extensions before comparing (`.MD` matches) but compares the extensionless
  conventional names (`README`, `LICENSE`, `CHANGELOG`) by exact case only, so a real-world
  `Readme` or `readme` would silently not be shown. S6 doesn't specify either way.
- Decision: normalize case for the extensionless names too, matching the extension path's
  existing behavior. Folded into M2.4 alongside the rest of the M2.1 review batch (T11, T12,
  T14) rather than a standalone task.
- Why: low risk (a strict widening of what's shown, never a narrowing) and real-world value
  (README variants in the wild are inconsistently cased); consistency with the already-decided
  extension-matching behavior is the simpler rule to hold in one's head than "extensions
  normalize, names don't".

## D26: M5.3 Help Modal is real cell compositing via lipgloss v2's Canvas/Layer, not a full-screen replacement

- Status: decided
- Cycle: C1
- Date: 2026-08-02
- Context: S12 says the modal "opens over any pane"; lipgloss v2's `Place` only positions
  text within whitespace, it does not preserve a background. Rendering the modal as a
  full-screen takeover (blanking everything else) would satisfy the letter of the acceptance
  criterion more cheaply, but not "over".
- Decision: lipgloss v2 ships a cell-based `Layer`/`Compositor`/`Canvas` (undocumented in the
  quick-reference used at plan time, found in the vendored source). `help.go`'s `overlayHelp`
  builds the already-rendered frame as one full-size layer (z0) and the modal box as a
  second, centered, smaller layer (z1), and renders the compositor — the pane underneath
  stays pixel-visible around the modal's edges. The modal itself is content from
  `bubbles/help`'s `FullHelpView` (going with the grain, D1) fed `helpBindings(pane)`, a
  fuller-than-footer list that includes what `paneHints` deliberately omits (PrevPane; for
  the viewer, Cancel/FindNext/FindPrevious — see footer.go's comment). Input is trapped the
  same way `handleFindInput` traps an in-progress search: while `m.helpOpen`, only Help or
  Cancel do anything, everything else (including Quit) is swallowed. Fresh-context review
  (code-reviewer subagent) caught that a Compositor sizes its output canvas to the union of
  every layer's bounds, not the base layer's own size — an unclamped modal taller/wider than
  the terminal (e.g. the viewer's full binding list on a short terminal) would grow the
  rendered frame past what `View()`'s no-alt-screen drawing can show correctly. Fixed by
  wrapping the modal box in `lipgloss.NewStyle().MaxWidth(width).MaxHeight(height)` before
  compositing, so it is truncated rather than overflowing; `TestHelpOverlayNeverExceedsTheTerminal`
  locks this in.
- Why: true compositing is the only way to keep the pane visible around the modal, which is
  what "opens over" means; going with the grain (D1) again rather than hand-splicing strings.
  The canvas round-trip re-serializes SGR runs and trims trailing padding on unaffected rows
  — cosmetically identical on a real terminal, but byte-different, which is why
  `help_test.go`'s overlay test compares `ansi.Strip`'d, trailing-space-trimmed lines rather
  than raw bytes. This round-trip only ever runs while the modal is open (`render()` skips it
  otherwise), so it cannot affect any other milestone's byte-exact frame comparisons.

## D27: Naming convention — "Find" is within a document, "Picker" is across files

- Status: decided
- Cycle: C2
- Date: 2026-08-02
- Context: `/`-search, `NOTES.md`'s "Fuzzy Finder" / "File Search" / "Document Search" and this
  cycle's own name all overlapped. A convention was wanted, to be retrofitted and then held.
- Decision: Adopt the mainstream editor/browser split — **Find** = incremental search within
  one document (C1's `/`); searching across files is a **Picker**, per Telescope's vocabulary
  (Prompt / Results / Preview pane), instantiated as **File Picker** and **Grep Picker**.
  "Finder" is banned: `NOTES.md` also used it for the OS file manager.
- Why: C1's `Find` already conformed, so nothing needed renaming — the cheapest possible
  convention is the one already accidentally followed. Glossary updated at coin time.

## D28: ripgrep is a hard requirement, and its ignore rules are the truth for Pickers

- Status: decided
- Cycle: C2
- Date: 2026-08-02
- Context: `ripgrep` gives interactive-speed grep for free. It could also list candidate files
  (`rg --files`). Either both Pickers depend on it or only grep does.
- Decision: Both. `ripgrep` supplies the file list *and* the text search; without it the
  session reads normally but both Pickers are unavailable with a message. rg's ignore rules
  (`.gitignore`, `.ignore`, hidden defaults) govern what Pickers see, and are allowed to
  disagree with C1's hand-rolled tree walk (`knowledge/gitignore-matching.md`).
- Why: One source of truth for "which files exist" beats consistency with the tree. Two
  Pickers in the same modal disagreeing with *each other* would be the worse inconsistency.
  Accepted cost: filename find — the cycle's core value — dies on a machine without rg.
  Note rg does no fuzzy matching; scoring is in-process (`sahilm/fuzzy`, already a transitive
  dependency of `charm.land/bubbles/v2`).

## D29: The Grep Picker query is literal, not regex

- Status: decided
- Cycle: C2
- Date: 2026-08-02
- Context: Telescope's `live_grep` passes the query to rg as a regex. Under incremental typing
  that means a half-typed `func (` is an *invalid* regex, and `v1.2` matches `v1x2`.
- Decision: Literal only (`--fixed-strings`), with smart case matching C1's Find. No regex
  mode this cycle; deferred to Unscheduled.
- Why: For prose notes, predictability beats power, and error flashes mid-type are worse than
  a missing feature. Regex can be added behind a toggle later without changing the contract.

## D30: The Preview pane shows raw text, not rendered markdown

- Status: decided
- Cycle: C2
- Date: 2026-08-02
- Context: A markdown reader previewing unrendered text is jarring; but a Grep Result is
  `file:line`, and `knowledge/viewport-search-ansi-offsets.md` records that line and character
  offsets drift once content is ANSI-styled.
- Decision: Raw text for both Pickers. The Grep Picker opens the preview scrolled to the
  matched line, highlighted; the File Picker previews from the top.
- Why: Precise match context *is* the point of a grep preview, and raw text is trivially
  line-addressable. Rendering it is polish, and polish that re-opens a known-painful problem.
  See D33 for whether it is ever worth it.

## D31: The Picker is a jump tool, not a browser — stateless and query-only

- Status: decided — 3-character minimum revised to 2 by C6/D116
- Cycle: C2
- Date: 2026-08-02
- Context: Several small choices (memory, minimum query length, what the query matches, what
  Activate moves) each looked independent but all pull in one direction.
- Decision: No memory of any kind — every open is File Picker / Active Root / empty Prompt.
  A 3-character minimum before either Picker searches, so there is no "list everything" state.
  The File Picker matches the **file name only**, never path components. Activate opens the
  document and changes nothing else: not the Active Root, not the tree.
- Why: One coherent identity — you arrive knowing roughly what you want, and leave with a
  document open. Browsing is the Files Nav's job; remembering is C4's (recall). Deliberately
  unlike Telescope, which lists all files on open.

## D33: Does the Preview pane ever become rendered markdown?

- Status: open
- Cycle: C2
- Date: 2026-08-02
- Context: D30 chose raw text "for now, circle back to polish". Rendering it would look right
  but reintroduces the line-addressing problem for grep previews.
- Decision: **Open.** Raw in C2. Unresolved: whether rendered preview is worth solving
  line addressing under Glamour output, whether only the File Picker renders (inconsistent
  between modes), or whether raw is simply correct for a search tool and this closes as wontfix.
- Why: Dogfooding C2 will show whether raw previews actually feel wrong in use.

## D34: A Grep Result's landing spot is re-derived in the rendered document, not trusted from ripgrep's line number

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Planning M4.3/M4.4 (S10, S11). ripgrep reports line numbers against the **raw file**;
  `find.go`'s match positions (`findMatch`, `applyCurrentMatch`) live in the Document Viewer's
  **rendered** line space — Glamour re-flows and word-wraps, frontmatter is pre-stripped (D5),
  tabs are expanded, and markdown syntax characters vanish. The two line numbers do not
  correspond, so "scroll the viewer to rg's line 137" would land in the wrong place.
- Decision: the Preview pane (raw text, D30) *does* address rg's line number directly — that is
  exactly why the preview is raw. The Document Viewer hand-off does not: after `openDocument`,
  re-run `findMatches(renderedContent, query)` and select the match by **occurrence ordinal** —
  the Nth literal occurrence of the query in the raw file maps to the Nth in the rendered
  content — then `commitFind` + `applyCurrentMatch`. If the ordinal cannot be resolved (the
  query spanned syntax Glamour removed, or sat inside stripped frontmatter), degrade: open the
  document with the query as a Committed Find on whatever matches do exist (first match), and if
  there are none, open it plain at the top. Never fail the activation, never crash.
- Why: reuses the existing Find machinery rather than authoring a raw↔rendered line map (D1),
  and keeps S11 honest — the point of the hand-off is that next/previous traversal works, which
  requires the rendered-space matches anyway. Ordinal correspondence holds for prose, which is
  what noggin reads; the degradation path covers the markdown-syntax edge without a special
  case per markdown construct. Cost: an activated match inside e.g. `**bold**` may land on a
  neighbouring occurrence. Accepted, and worth a TRIAGE item if dogfooding shows it biting.

## D38: M1.4's grep searches listFiles' explicit file list, not a second rg directory walk

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: M1.3's handoff assumed grep would shell out to rg the same way `rgFiles` does
  (`rg` walking Scope's directories itself, carrying `--no-require-git`, D36). Verified against
  the sandbox's real rg (15.1.0) first: rg only applies `.gitignore`/hidden-file rules to
  directories it walks itself, never to files named explicitly on its command line — an
  explicitly-passed `.gitignore`d file is searched anyway.
- Decision: `grepFiles` calls `listFiles` for the file set and passes those paths to rg
  explicitly (`rg --fixed-strings --smart-case --json -- <query> <files...>`), rather than
  letting rg discover files under Scope's roots itself. Line/file/text output is parsed from
  `--json` (one line's `path`/`line_number`/`lines.text`), not the plain `path:line:text`
  text format, since a colon can appear in either a path or a line's text and JSON does not
  leave that ambiguous.
- Why: high value, no real risk (an implementation-internal call shape, reversible), so
  decided and logged rather than asked. A directory-walking rg invocation for grep would have
  needed its own re-derivation of the allowlist/ignore-rule filtering listFiles already does
  for the File Picker, and — per the finding above — would have leaked ignored/hidden files
  into grep results that the File Picker would never show for the same query, the exact kind
  of two-Pickers-disagreeing-with-each-other D28 already ruled out. Reusing listFiles' output
  is also literally the smart-case/fixed-strings query applied to "the Readable files in
  Scope" as SPEC (S4) states it, with no second definition of that set anywhere.
- Also: smart case is rg's own `--smart-case` flag rather than a reimplementation of
  `find.go`'s fold-and-compare (verified it gives find.go's exact semantic: all-lowercase
  matches either case, any uppercase makes it case-sensitive).

## D41: Picker search is a trailing-edge debounce over a monotonic query generation

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: M2.4 needed the first genuinely async work in the codebase (no in-repo precedent —
  `findings/codebase-picker-integration.md`): the Prompt must never block on rg, in-flight
  searches must be debounced, and results for a superseded query must never be displayed.
  Bubble Tea gives no way to cancel a `tea.Cmd` once started, so "discard" has to happen on
  arrival, not at the source.
- Decision: one `picker.seq` counter, incremented by every change to what would be searched
  (a keystroke now; M5.1/M5.2's mode/Scope switch later). Each change schedules a
  `tea.Tick(120ms)` carrying its seq (`pickerSearchMsg`); a matured timer dispatches the real
  search only if its seq is still current, and a returning `pickerResultsMsg` is applied only
  if its seq is still current. Same comparison serves both debounce and stale-discard. The
  counter is monotonic for the model's life — `dismissPicker`/`openPicker` bump it instead of
  resetting, so a search in flight across a dismissal cannot be accepted by the next Picker.
  The search Cmd is a pure function of copied-in values (`pickerSearchCmd`), touching no model
  state, since it runs concurrently with `Update`.
- Why: low-risk, reversible and entirely internal (no spec surface), so logged not asked. One
  counter is cheaper than the alternatives — a cancellable `context` threaded into M1's search
  layer (which would make it Bubble-Tea-aware for no gain, since the goroutine's result still
  has to be dropped on arrival) or a leading-edge debounce (which would search the first
  keystroke of every burst, i.e. the least useful query). 120ms is a build-time judgment call,
  not spec'd; S8's "no memory" is untouched because a generation count carries no information
  about mode, Scope or query.

## D42: An interstitial cycle is keyed C2a, not C6

- Status: decided
- Cycle: C2a
- Date: 2026-08-05
- Context: A short cycle of dogfooding tweaks was inserted ahead of the already-slotted
  `C3 compare`. Cycle keys are durable and already referenced by intent in durable docs
  (`DECISIONS.md` D-entries citing "(C3)", "C4 per-root document state", "Tags are C5";
  `GLOSSARY.md`'s "documents in C4"), so renumbering C3/C4/C5 was not available.
- Decision: allocate the suffixed key **C2a** and insert its row ahead of C3 in the roadmap
  ledger. Ledger order is independent of key order.
- Why: suffixed insertion is the sanctioned idiom for durable keys, and `C2a` sorts correctly
  in `archive/` between C2 and C3, which a bare `C6` would not. Cost: `C2a` reads as a
  follow-on to `find`, which it is not — accepted as cheaper than a positionally misleading key.

## D43: The render-mode toggle and per-line git gutter are cut without a roadmap entry

- Status: decided
- Cycle: C2a
- Date: 2026-08-05
- Context: C2a was pitched with a toggleable markdown render mode (wrapped "read" mode vs
  unwrapped "info" mode), a per-line git gutter, and exact-word Find highlighting.
  `cycle/findings/git-status-gutter-approach.md` established that the dual-mode approach does
  make the gutter tractable without forking Glamour (`WithWordWrap(0)` +
  `WithPreservedNewLines()`, plus a goldmark-based source-to-rendered line map, ~3/5 overall)
  — but also that (a) Find is already wrap-mode-independent (`find.go:186-190` matches against
  `ansi.Strip(content)` in rendered-line space), so inline-span highlighting is an orthogonal
  problem the mode does not solve and the ANSI-offset drift of
  `knowledge/viewport-search-ansi-offsets.md` still applies; (b) an unwrapped mode is unreadable
  without horizontal scrolling, which noggin has never wired up and which cuts against D4's
  "scrolling only, no cursor" framing; (c) a per-*file* indicator is mapping-free and cheap.
- Decision: cut the render modes, the per-line gutter, exact-word highlighting and horizontal
  scroll from C2a, keep only per-file `M`/`A` markers (C2a/S8) — and deliberately do **not**
  create ROADMAP Unscheduled entries for the cut items.
- Why: the mode toggle's remaining justification was the gutter alone, and it would have cost
  two divergent reading experiences plus a new scroll idiom for it. Omitting the roadmap
  entries is intentional: the user wants to discover whether per-file markers and refresh
  satisfy the underlying need before re-committing to the expensive version. If the need is
  still felt after C2a, it gets re-raised from scratch rather than inherited as a stale
  backlog item.

## D44: Peek is allowed to feel unfinished

- Status: decided
- Cycle: C2a
- Date: 2026-08-05
- Context: C2a/S7 renders the Files Nav selection in the Document Viewer without changing the
  Active Document. That creates a second viewer state, which collides with globals defined on
  the Active Document — most sharply `e` (Edit), where acting on the Peeked document
  contradicts the glossary and acting on the Active Document is astonishing.
- Decision: keep Peek deliberately inert — no scrolling, no Find, `e` a silent no-op, cleared
  by any selection change, Focus change, Picker open, second `p` or Refresh, with `<enter>` the
  only promotion path. One concession to UX: an inverted header naming the Peeked document.
- Why: Peek is an experiment and a cheap stand-in for the still-deferred Follow mode. Choosing
  "feels broken at the edges" over "premature machinery" keeps it a few hours of work instead of
  a redesign of the Active Document contract. If Peek proves itself, its edges get specified
  then — with real usage as evidence.

## D46: Git status markers shell out to `git`, resolved via repo toplevel

- Status: decided
- Cycle: C2a
- Date: 2026-08-06
- Context: M4.1/S8 needs per-file modified/new markers, resolved relative to the containing
  repository's root, for a Root that may or may not be inside a git repository. The
  gitignore-matching precedent (`.charm/knowledge/gitignore-matching.md`, D18) rejected
  shelling out to `git check-ignore` for a *different* reason — it needs an arbitrary,
  possibly non-repo directory to behave like a git worktree, which arbitrary noggin Roots
  are not. S8's ask is the opposite: it explicitly wants real git semantics, only within a
  real repo, with staged/unstaged/deleted/renamed/submodule/worktree cases pre-declared out
  of scope — a much shallower contract than gitignore matching, and one hand-rolling would
  mean reimplementing (index parsing, blob hashing) for no correctness gain.
- Decision: `gitstatus.go` shells out to the real `git` binary (`exec.LookPath`-detected, the
  same degrade-quietly shape as C2/S14's `detectRipgrep`), two calls per Root load: `rev-parse
  --show-toplevel` (doubles as the "inside a repo at all" check) then `status
  --porcelain=v1 --untracked-files=all --no-renames -z -- .`. Porcelain paths are always
  toplevel-relative regardless of invocation directory (verified empirically, git 2.53) — so
  toplevel resolves the absolute path per S8's "relative to the containing repository's root,
  not the Root" wording, and pathspec `.` (with `-C root`) scopes results to files under the
  Root without walking the whole repo status separately. `--no-renames` disables rename
  detection so every changed path is one flat entry (no orig-path continuation field to parse
  and discard), matching S8's "no renamed marking" directly. Computed fresh in
  `model.openFolder` on every call (launch and every Root switch) — not cached, unlike M3.1's
  per-Root memory — since the point is current status, not a remembered browse state.
- Why: git binary is a reasonable runtime dependency the same way ripgrep already is (C2/S14);
  hand-rolling git's index/worktree diff algorithm is a different order of complexity than
  hand-rolling `.gitignore` glob matching, with no upside since S8 deliberately wants real git
  behaviour, simplified, not an independent approximation of it.
- Addendum (2026-08-09, replan, TRIAGE T8): unlike D41's `rg` calls, which moved off the
  Update thread specifically because they run on every keystroke behind a debounce, D46's two
  `git` calls run synchronously from `openFolder`/`refresh`, gated on discrete user actions
  (Root switch, launch, explicit `R`) rather than typing — the same synchronous-and-gated
  shape `tree.Walk` already has, not an oversight relative to D41's async precedent.

## D47: S5's two-line rows drop the old bold-means-active convention

- Status: decided
- Cycle: C2a
- Date: 2026-08-07
- Context: M8.1/S5 makes the base name bold on line 1 "unconditionally" (S5's own wording),
  the path unbold on line 2. Pre-S5, the single-line layout used the *same* Bold attribute to
  mean "this is the active root" (`rootsActiveStyle`, applied only to `idx == r.active`), with
  the `●` marker as a second, redundant cue for the same fact. Literally implementing S5's
  "name in bold" as unconditional (every row's name line bold, active or not) removes bold as
  an active-root signal — a row's only remaining tell for "is this the active root" becomes
  the `●` marker alone.
- Decision: implemented literally: `rootsNameStyle`/`rootsPathStyle` (rootsnav.go) are the
  same for every row regardless of `idx == r.active`; broken rows still get `rootsBrokenStyle`
  (Faint) on both lines, and the selection cursor still reverses both lines. No attempt to
  preserve a second active-only visual cue beyond `●`.
- Why: S5's acceptance text ties bold specifically to the name text, not to active status; a
  row that is both active and bold-for-being-active *and* bold-for-being-a-name-line would be
  indistinguishable from a bold-for-being-a-name-line-only row anyway, so keeping the old
  active-bold rule on top would have been a no-op requiring extra code, not an additional
  signal. `●` was always present as an independent cue and is unaffected by this. Flagged here
  rather than silently changed since it is a real, user-visible reduction in redundant
  signalling (from two cues to one) that a fresh session might otherwise wonder about when
  diffing `rootsnav.go` against C1-era behaviour.

## D49: S10's staleness check is a directory-mtime heuristic, and its indicator is a Footer suffix

- Status: decided
- Cycle: C2a
- Date: 2026-08-08
- Context: PLAN.md M10.1/SPEC.md S10 require a periodic (~5-10s) check that reports "the
  Active Document or the Active Root's tree has changed on disk" without refreshing or
  re-walking anything, and must not perceptibly cost anything on a large Root; SPEC.md's own
  "Left to the plan" list names the indicator's surface (Footer vs elsewhere) as unspecified,
  and PLAN.md's M10.1 entry does not pin down a detection mechanism either — both were left
  for this task to decide.
- Decision (detection): the periodic check compares a snapshot (staleBaseline,
  staleness.go) against the disk: one os.Stat per directory the already-loaded tree knows
  about (root plus every directory node under it, walked in memory — no os.ReadDir call),
  plus a direct Stat of the Active Document's own path for its mtime and size. Any mismatch,
  or any known path that no longer stats, means "stale." It never calls tree.Walk. The whole
  check runs off the Update goroutine (a tea.Cmd, mirroring picker.go's async search/D41).
  Unlike that search's own debounce, a matured periodic timer always dispatches a check
  against whichever baseline/generation are current *right now*, never ones captured at
  schedule time — a superseded-timer discard (picker.go's own shape) would let a
  resetStaleBaseline landing mid-flight push the next real check out to roughly two
  intervals instead of the one M10.1's acceptance criterion names. A generation counter is
  still used, but only at the result-applying end: a completed check's verdict is discarded
  if the baseline it was computed against has since been superseded, so a slow check can
  never stomp a state that has already moved on.
- Decision (indicator surface): the Footer. Unlike a transient m.message (D22), which
  replaces the pane's key hints outright, the indicator is appended to whatever hint line
  would already show (pane hints or the open Picker's) as one more " · "-joined segment,
  dropped if it would not fit at the current width — S10's fixed wording, never displacing
  the hints someone might need to act on something else while it's showing. It loses to an
  in-progress find and to a transient message, both of which still need the whole line.
- Why (detection): a full recursive Stat of every file, or re-running tree.Walk, is exactly
  the cost S10 forbids on a large Root; a directory's own mtime changing on file
  add/remove/rename inside it is a standard, cheap signal for "this subtree changed" that
  needs no new directory listing since the caller already has every directory's path from
  the last real walk. It is deliberately not exhaustive: an in-place edit to a *different*,
  non-Active file that touches neither its parent directory's entry list nor the Active
  Document itself goes unnoticed until the next refresh. Catching that case would require
  statting every file, which reintroduces the cost this decision exists to avoid. The Active
  Document has no such gap — it is statted directly, every check, regardless of the
  directory heuristic. Reusing `git status` (already shelled out to for S8, D46) was
  considered and rejected as the general tree-changed signal: it would leave non-git Roots
  with no coverage at all for anything but the Active Document, a bigger gap than the
  heuristic's own, for no simplification (the git path still needs the Active Document Stat
  either way, and mixing S8's status computation with S10's unrelated concern muddies both).
- Why (indicator surface): the Footer is the one region already shown to every pane
  regardless of focus (S11), the same reasoning FocusRoots/FocusFiles/Refresh already being
  global bindings relies on; appending rather than replacing follows from S10 itself being a
  standing condition (persists until a refresh), not a one-shot notice — D22's transient
  shape is the wrong fit for something that must still be visible many keypresses later.

## D54: no minimum-width guard for the DV1/DV2 50/50 split

- Status: decided
- Cycle: C3
- Date: 2026-08-08
- Context: `layout.go`'s `computeLayout` already guards the Navigator-vs-viewer boundary
  (`minViewerWidth`, collapsing the Navigator to 0 on narrow terminals) but has no equivalent
  concept for splitting the viewer region itself in two. SPEC.md fixes the split at exactly
  50/50 with no ratio adjustment, and neither SPEC.md's Out-of-scope list nor VISION.md's
  boundaries say what should happen when a terminal is too narrow to carry two useful halves
  — a genuine gap, not an oversight either doc owns.
- Decision: no guard. Each half always renders at its exact 50/50 share of the viewer region
  (rounding down, remainder to the left half), however small; there is no collapse-to-one-
  viewer behavior analogous to the Navigator's. `minViewerWidth` continues to gate the
  Navigator-vs-(combined)-viewer boundary only, unchanged and untouched by DV2.
- Why: S1 already makes the split's fixed-ness a deliverable ("exactly two, exactly 50/50
  left-right, fixed" is explicitly Out of scope to change); inventing a second collapse
  threshold that silently drops DV2 on narrow terminals would contradict "DV2 exists once
  created, session-only" (S8) with a rendering-driven exception SPEC.md never asked for, and
  would need its own restore/re-split semantics on a subsequent resize — real design surface
  for a case the spec doesn't flag as a concern. Rendering degrades gracefully already
  (glamour/lipgloss wrap or clip narrow content; this is the existing single-viewer behavior
  at small widths too, just applied to a narrower region) — acceptable at the very narrow
  end, revisit only if it becomes a real reported complaint.

## D55: `shift+` reverted to `ctrl+` for S3/S5; T3 closed by making the fallback non-printable

- Status: decided
- Cycle: C3
- Date: 2026-08-08
- Context: D53 chose `shift+enter` (+ `L` fallback) for S3's targeting modifier, and S5's
  `shift+d`/`D` (FocusDV2) followed the same prefix for consistency, after an earlier `ctrl+`
  pitch was dropped: a fully-`ctrl+`-prefixed scheme needs `ctrl+d` for "jump to DV2"
  (mirroring `d` for DV1), which collides with the pre-existing `ctrl+d` (HalfPageDown).
  Separately, TRIAGE T3 found D53's `L` fallback can't be reused in the Picker —
  `handlePickerInput`'s Prompt is always-typing, so a printable fallback key risks being typed
  into the query instead of triggering the action.
- Decision: revert to `ctrl+` as the one consistent modifier across S3 (Files Nav's
  `ctrl+enter`/`ctrl+p`; the Picker's `ctrl+enter`) and S5 (`ctrl+o`) — the DV1 focus key
  itself moves from `d` to `o` (S5), freeing `ctrl+o` for FocusDV2 and avoiding the `ctrl+d`
  collision that killed the original pitch. D53's `L` fallback for `ctrl+enter`'s terminal
  ambiguity (identical risk to `shift+enter` — switching modifiers doesn't change it, D53
  already noted this) is replaced with `ctrl+l`: a raw control byte, not a printable
  character, so it can never be typed into the Prompt as query text. This closes T3 outright
  — no Picker-specific carve-out (`alt+enter`, a second non-printable-only path, etc.) is
  needed, `ActivateDV2` is one binding used identically in both Files Nav and the Picker.
  `ctrl+o` (FocusDV2) needs no fallback of its own: unlike `ctrl+enter`, a plain Ctrl+letter
  is a raw C0 control byte delivered identically by every terminal regardless of
  Kitty/CSI-u/modifyOtherKeys support, so there is nothing to degrade.
- A candidate considered and rejected: keep S5's focus-jump key as `d`/`shift+d` and give
  FocusDV2 its own `ctrl+.`-style binding instead of renaming the base key. Ctrl+ on
  non-letter punctuation has no legacy VT100/CSI encoding at all — Ctrl clears bit 6 of a
  0x40–0x5F byte, and `.` is 0x2E, already below that range — confirmed against the vendored
  `ultraviolet` decoder (`decoder.go`'s `parseControl`/`LegacyKeyEncoding`), which has no code
  path producing `Mod: ModCtrl` for `.` outside Kitty/CSI-u/modifyOtherKeys. Its degradation is
  worse than `ctrl+enter`'s (byte-identical to plain `enter` is at least a well-formed
  keystroke): on GNOME Terminal/VTE, Alacritty and likely Windows Terminal/Terminal.app it
  falls through to plain `.`; on a bare tty it can produce nothing; relayed through tmux/screen
  from a Kitty-capable outer terminal it can be silently dropped. Renaming the base focus key
  (`d`→`o`) avoids ever needing a Ctrl+punctuation binding.
- Why: a binding choice, not a spec-deliverable change (D2) — S3's and S5's contracts
  (which action, what it does) are unchanged, only which physical chords carry them. Recorded
  as a decision, not silently reworked, because SPEC.md's S3/S5 text names specific keys
  (unlike most deliverables, which describe actions only) and because it reopens two
  milestones already marked `done` (C3/M1.3, C3/M1.4) — see PLAN.md's M1.3a/M1.4a. Supersedes
  D53's chord choice; D53's fallback-is-necessary reasoning and its terminal-disambiguation
  research (Kitty protocol/CSI-u/modifyOtherKeys) stand unchanged.

## D59: DV1's resize find-dropped wording stays unprefixed; DV2's is named, combined drops join with a comma (TRIAGE T5)

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: M1.6a extends `resize()`'s find-dropped message (`"…; search cleared"`/`"…; search
  dismissed"`) to DV2, which the plan task left as an open call: "decide how two simultaneous
  drops combine into the single `m.message` slot."
- Decision: DV1's own wording is left exactly as it already was — unprefixed — when DV2 drops
  nothing, so every pre-M1.6a test asserting the literal "search cleared"/"search dismissed"
  text keeps passing unchanged. DV2's drop is always named explicitly ("dv2 search
  cleared"/"dv2 search dismissed"), lowercase per D57's hint-style precedent, since an
  unprefixed message would be ambiguous once two viewers exist. When both viewers drop in the
  same resize, the two clauses join into one message: `"<cause>; dv1 <drop>, dv2 <drop>"`.
- Why: medium risk×value (a messaging-only choice, but one that touches every existing
  resize-message test's exact wording) — decided and logged rather than asked. Preserving
  DV1's unprefixed wording is what makes the change backward-compatible for free; naming DV2
  explicitly rather than leaving it equally unprefixed is the only option that stays
  unambiguous once a message might need to describe either or both viewers at once.

## D60: Stale row supersedes D49's Footer-suffix surface; color reuse, message coexistence, low-height precedence (C3/M5.1)

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: M5.1 relocates S10/C2a's staleness indicator off the Footer's hint line into its
  own dedicated row (S9), superseding D49's "indicator surface: Footer" half (D49's detection
  mechanism, staleBaseline/staleness.go, is untouched and still current). S9 pins wording,
  "occupies no space otherwise" and styling ("highlight color, inverse text") but leaves three
  things open that D49 previously had opinions on: which highlight color, how the row
  interacts with a transient m.message (D22) now that they are different lines, and what
  happens to the row/Footer split at a terminal too short to fit both.
- Decision (color): reuse `lipgloss.Color("214")`, the Help Modal border's own color
  (`modalStyle`, help.go) — noggin's one existing "this needs your attention" color — rather
  than introduce a second, arbitrary highlight color. `staleRowStyle` (style.go) is
  `Foreground(214).Reverse(true)`, matching S9's "highlight color, inverse text" by swapping
  the specified foreground into the row's background via `Reverse`.
- Decision (message coexistence): the row and a transient Footer message (D22) now show
  simultaneously — they no longer share one line, so D49's "message wins, suffix dropped"
  rule (which existed only because both competed for the same Footer line) has no successor.
  `renderFooter` no longer looks at `m.stale` at all.
- Decision (low-height precedence): `computeLayout` reserves the Footer's own height first,
  then gives the stale row whatever remains — at a terminal shorter than 2 rows the row is
  silently dropped rather than displacing the Footer, the same "never push a more essential
  region out" precedence the frame's other regions (Navigator vs. viewer, `minViewerWidth`)
  already follow at extreme sizes.
- Why: low-to-medium risk×value, all three are narrow implementation choices with an obvious
  "least surprising" answer and no real alternative reading of S9 — decided and logged rather
  than asked, per the effort-scaling rule.

## D61: Document memory is one global, path-keyed store — shadow Roots rejected

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: C4 brainstorming. Starring documents per Root produced a chain of bad cases (star in
  an ad-hoc Root, then open a parent folder as a Root next session; star the Root; un-star it).
  A "shadow Root" concept was proposed to hold recency for ad-hoc and un-starred Roots: a
  retained-but-unstarred Root class with a TTL, a quantity cap, and a carve-out letting shadow
  Roots nest.
- Decision: one persisted store (S1) keyed by absolute path, holding `{starred, lastViewed}`,
  **not attached to Roots**; per-Root Starred and Recent lists are derived by path prefix. No
  Root lifecycle event destroys document state. Shadow Roots are rejected outright — not
  deferred, since nothing remains to build. The per-Root cap of 100 becomes a *view* cap; the
  store's own bound is separate housekeeping needing no Root attribution. The starred-Roots
  store moves from config (C1/D7) to the same data location, closing that Unscheduled item, as
  a clean break with no migration.
- Why: derivation is unambiguous because Roots may not nest (D6) — the same rule shadow Roots
  would have had to violate, reintroducing exactly the "which Root is this document in"
  ambiguity D6 exists to prevent. Shadow Roots would have added a durable concept with a full
  lifecycle and time-based expiry (in an app whose refresh philosophy, C2a, is "manual only,
  never watches anything") for **zero visible surface** — while a global store delivers the same
  outcome by deleting a rule rather than adding a concept. Reverses the earlier intent that
  ad-hoc Roots' recency be session-only: it is now persisted like everything else, which is the
  trade that makes this work. Also rejected: a confirm dialog offering to purge document stars
  when a Root is un-starred — the signal is weak (un-star means "stop auto-adding at launch",
  not "done with these documents"), it is irreversible, it would take out stars created before
  the Root was ever starred, and it would need modal-dialog machinery noggin does not have
  (cf. D26's cost). Bulk un-starring is done one document at a time from the Picker's Starred
  scope, which is also what keeps orphaned stars reachable.

## D62: Wiki-links leave C4; the document cursor is bought once, for everything that needs it

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: The roadmap's C4 line named wiki-links alongside history, starring and recency.
  Following a link needs a notion of "which link is current" — precisely the Document Viewer
  cursor D4 refused, in ANSI-styled rendered output where
  `knowledge/viewport-search-ansi-offsets.md` records offsets drifting.
- Decision: wiki-links and URL following are cut from C4 and moved to ROADMAP Unscheduled,
  bundled with Select mode + copy (also from C1) as one cursor-dependent group. A link Picker
  (a modal list of the document's links, needing no cursor) was floated as a cursor-free
  alternative and deliberately not taken this cycle.
- Why: C4 is already the largest scope bet since C1 without it, and the honest cost of
  wiki-links is not the parsing — it is the cursor. Whatever eventually justifies that cursor
  should pay for every feature that needs it at once, rather than one feature smuggling in
  half of it.

## D64: The Picker's Scope axis becomes file-set-based, not Root-set-based (C4/M5.1)

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: Planning M5.1 (S8). `liveRootsInScope` (`cmd/noggin/scope.go:26`) resolves a Scope to
  a set of live Roots, and both `listFiles` and `grepFiles` search from those Roots outward
  (`cmd/noggin/files_search.go:57,375`). `Starred` scope has no Root set at all — it is a set of
  paths from the document store, including orphans under no Root.
- Decision: refactor Scope resolution to yield a *file set*, with `Active Root` / `All Roots`
  expressed in those terms, rather than special-casing `Starred` as a branch beside the Root
  path. Both Pickers get the third value in the same change (S8), keeping the mode-switch
  invariant.
- Why: a `Starred` branch bolted beside a Root-shaped abstraction would leave two search paths
  to keep in agreement — exactly the drift D34-era work already paid for once. One resolution
  shape for all three values is cheaper than a carve-out.

## D66: dataDir() hand-rolls the data-location split rather than a dependency (C4/M1.1)

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: S1 requires a data location distinct from `os.UserConfigDir()`; the standard library
  has no `UserDataDir()`. Options were a small XDG dependency (e.g. `adrg/xdg`) or a hand-rolled
  function.
- Decision: hand-roll `dataDir()` in `cmd/noggin/store.go`, mirroring the exact structure and
  style of the stdlib's own `os.UserCacheDir()`/`os.UserConfigDir()` (same `runtime.GOOS` switch,
  same env vars and fallbacks): `$XDG_DATA_HOME` or `$HOME/.local/share` on Unix,
  `~/Library/Application Support` on Darwin, `%LocalAppData%` on Windows, `$home/lib` on Plan9.
- Why: D1 (go with the grain of the libraries) favors the stdlib's own idiom over a new
  dependency for four lines of well-specified, rarely-changing logic; consistency with
  `os.UserConfigDir()`'s existing style (already read by every contributor) keeps the two
  functions easy to compare and trust. Medium and reversible — logged, not asked.

## D67: Document store uses a tab-separated two-field line format (C4/M1.2)

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: M1.2's document store needs a per-line encoding for `{starred, lastViewed}` keyed by
  an absolute path that may itself contain spaces. The Roots store's plain one-path-per-line
  format (D7) doesn't extend to a second field without a delimiter choice.
- Decision: `<starred:0|1>\t<lastViewed:RFC3339Nano|-> \t<absolute path>`, one record per line,
  tab-separated, in its own file (`documents`, `docStoreName`) alongside the Roots store under
  the same data directory (M1.1). Parsing drops any line that fails to parse (wrong field
  count, bad flag, bad timestamp, non-absolute path) rather than failing the whole load, same
  tolerance as `loadStore`. A duplicate path across lines has the later line win.
- Why: tabs are already excluded from valid file paths, so they're an unambiguous, zero-escaping
  delimiter — no JSON/encoding dependency needed (D1), and the format stays greppable/inspectable
  even though S1 does not require it to be hand-edited the way the Roots store is. Medium and
  reversible (no consumers outside this file yet) — logged, not asked.

## D68: Document store bound fixed at 1000 records (C4/M1.3)

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: S4 says the store as a whole is bounded and the bound is fixed, not configurable
  (S1's Out-of-scope list), but no cap number appears anywhere in SPEC.md, VISION.md or
  DECISIONS.md — it was left for M1.3 to pin.
- Decision: `docStoreCap = 1000` (`cmd/noggin/docstore.go`). `evictDocStore` walks records
  oldest-`LastViewed`-first once the map exceeds this, deleting unstarred candidates and
  clearing (not removing) a starred candidate's timestamp, stopping as soon as the map is back
  within cap.
- Why: 10x the per-Root `Recent` view cap (100, S6) — generous for a store spanning every Root
  a session has ever touched, while keeping the on-disk tab-separated file (well under 100KB at
  this row size) and any future in-memory scan trivially cheap. Medium and reversible (a
  constant with one call site, no persisted format dependency on its value) — logged, not asked.

## D69: The Picker's star action is a new `PickerStar` binding (alt+s), not a reuse of `keys.Star`

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: M2.5/S2 adds a third star surface, the Picker's selected Result. The Roots Nav,
  Files Nav and a focused Document Viewer all bind the same `keys.Star` (`s`, a bare printable
  letter) since none of them is a typing surface. The Picker's Prompt is "always-typing"
  (`handlePickerInput`'s own doc comment, S2's own contract): every key not in a fixed
  non-typing set falls through to the Prompt as query text, which is exactly why `PickerUp`/
  `PickerDown` avoid `j`/`k`, `PickerMode`/`PickerScope` use Tab/Shift+Tab instead of a letter,
  and `ToggleHidden` uses `alt+h` — reusing bare `keys.Star` here would make every `s` typed
  into a query toggle the selected Result's star instead of being typed.
- Decision: a new `keyMap` field, `PickerStar`, bound to `alt+s` (mirroring `ToggleHidden`'s
  own `alt+` prefix for the identical reason), wired only in `handlePickerInput` and added to
  `pickerHints()` (Footer). It is deliberately **not** added to `helpBindings`'s per-pane lists
  the way `PickerUp`/`PickerDown`/`PickerMode`/`PickerScope`/`ToggleHidden` are: its `Help()`
  description ("star") collides with plain `keys.Star`'s own, which
  `TestHelpModalListsThePaneOwnBindings` already pins as *absent* from the Files Nav's and the
  Document Viewer's own Help Modal (neither pane binds plain `keys.Star` itself, only Roots Nav
  does). Reworking that pinned wording to make room was judged not worth it for one binding;
  the Footer's own live hint line is `PickerStar`'s discoverability surface for now — logged as
  a real, smaller gap in TRIAGE rather than silently accepted.
- Why: medium and reversible (an internal keymap choice, no spec surface — S2 says "acting on
  the selected Result", not which physical key), decided and logged rather than asked. A
  distinct binding is cheaper than teaching the Prompt's typing-vs-action dispatch a
  per-context exception for one key, and follows the codebase's own established precedent
  (`PickerUp`/`PickerDown`'s doc comment) for exactly this situation.

## D73: M3.1's recording does not cover launch-time document opens

- Status: decided
- Cycle: C4
- Date: 2026-08-10
- Context: S4 names two recording triggers verbatim — "Activated (in DV1 or DV2)" and "opened
  via the `$EDITOR` handoff" — and M3.1's own acceptance criterion only asks for tests
  covering Activate (both viewers) and the edit handoff, plus Peek's exclusion. `applyResolution`
  (S1's launch-time `openDocument`/`openDocumentInto` call for a document named on the command
  line) shares `openDocumentInto` with `activateInto`/`finishEdit`, so recording could have been
  centralized there and picked up launch opens "for free" — but neither S4 nor M3.1/M3.2 ever
  say a launch open counts, and M3.2's own wording ("a test Activates a document from an
  ad-hoc Root...") reuses "Activates", not "opens", reinforcing the narrower reading.
- Decision: `recordViewed` is called explicitly at the two named call sites
  (`activateInto`, `finishEdit`) rather than inside the shared `openDocumentInto`/`openDocument`
  — a launch-time document open does not stamp `LastViewed`.
- Why: low risk, straightforward reading of the spec's own enumerated trigger list; the
  alternative (recording on launch too) is a small, easily-reversible extension if the user
  wants it, not a foreclosed path — noted here rather than silently assumed either way.

## D79: Recent's "no session Root" bucket is capped at recentPerRootCap like any other, and All Roots is computed live per render — M4.2

- Status: decided
- Cycle: C4
- Date: 2026-08-11
- Context: S6 says "`Recent` shows at most 100 documents per Root; `Recent`/`All Roots` merges
  each Root's list sorted by timestamp with no further cap." It does not say what happens to a
  document last viewed while covered by a Root no longer in the session, or never covered by
  one at all — Recent's own scope set (S5) has no dedicated Orphans value the way Starred's
  does, so these "orphan" records have no Active-Root-scope home, yet S6's own "two right-hand
  label forms" line only makes sense if they can appear somewhere in Recent. Separately,
  `filesNav` had no existing pattern for a flat list whose data (the S1 store, the session's
  Roots) lives on `model`, not the pane itself.
- Decision: treat "no covering Root" as its own bucket, capped at `recentPerRootCap` (100)
  exactly like a real Root's bucket, then folded into `All Roots`'s merge with every other
  bucket — this is where an orphan document surfaces in Recent, and the source of its
  left-truncated-parent-path label form. `Active Root` scope never reads this bucket (an
  explicit `activeRootPath == ""` guard, not a `buckets[""]` lookup that would coincide with
  it). Row data itself is never cached on `filesNav`: `model.refreshFilesRecentSource` pushes a
  fresh `(docs, roots, activeRootPath)` snapshot in at every point that could change (`withDocs`,
  `withDocRecord`, `openFolder`, `applyResolution`, `refresh`), and `recentRows` recomputes the
  actual rows from that snapshot on every `rowsView` call — the same "no cache to go stale"
  shape M2.11 chose for Picker-row star markers over M2.7's Files Nav marker cache, picked here
  because mode/scope can change on every keypress (`cycleMode`/`cycleScope`) with no natural
  place to intercept that the way a document load or star toggle already has one.
- Why: the orphan-bucket cap is the simplest reading that makes S6's two-label-forms clause
  true without inventing a second cap value nowhere named in SPEC.md; capping it is also
  strictly safer than not (bounds the merge's own cost) and symmetric with every other bucket.
  The live-recompute choice avoids adding a second, harder-to-keep-consistent cache alongside
  `status`/`starred`'s existing one, at a cost (`recentRows` re-scans `m.docs` and re-derives
  per-path Root coverage every render) the store's own `docStoreCap` (1000 records) keeps cheap
  regardless — the same reasoning D75 already accepted for `filesStarredMarks`' full recompute.

## D80: activateInto/toggleDocStar resolve against the flat row, not the tree, while Recent/Starred is showing — M4.3, narrows TRIAGE T21

- Status: decided
- Cycle: C4
- Date: 2026-08-11
- Context: M4.3's own acceptance wording requires two behavioral tests — "activating a broken
  star is refused the way a Broken Root is" and "un-starring drops the row" — but M4.3 depends
  only on M4.1/M2.1, not M4.5 (the task that gives flat-list rows Activate/Peek/Edit/DV2-
  targeting generically, per its own acceptance wording). Before this task, `activateInto` and
  `toggleDocStar` both resolved the "selected document" via `filesNav.selectedNode()` — the
  Tree's own row set — regardless of Browse Mode (TRIAGE T21's exact citation), so there was no
  way to reach "activate this Starred row" or "un-star this Starred row" at all once `starredRows`
  existed, let alone test the refusal M4.3's wording names.
- Decision: give `filesNav` a `selectedFlatRow()` accessor (the flat row at `f.selected` while
  `flatMode()`, mirroring `selectedNode()`'s tree-based shape) and route `activateInto` and
  `toggleDocStar` through it first when the pane is in Recent or Starred mode, falling back to
  the existing tree-based path only for Tree. A broken flat row's own refusal
  (`activateFlatRowInto`) mirrors `activateRoot`'s wording exactly (`"%s no longer exists; not
  activated"`). This is scoped to exactly these two actions; Peek, Edit, and ctrl+ DV2-targeting
  keep resolving via `selectedNode()` (still wrong while Recent/Starred is showing) until M4.5
  gives them the same fix, and Expand/Collapse are untouched (M4.5 makes them explicit no-ops).
- Why: M4.3's acceptance criterion cannot be met, even minimally, without some path to attempt
  activating/un-starring a flat row — deferring both to M4.5 would leave the task's own stated
  done-condition untestable. Scoping the fix to Activate+Star only (not Peek/Edit/DV2-targeting)
  keeps this from silently absorbing M4.5's broader mandate while still resolving T21's root
  cause for the two actions M4.3 actually needs; T21 itself is unchanged here since Peek/Edit
  remain broken exactly as it describes, but a future replan should read T21 as narrowed to
  those two actions, not Activate/Star, once this lands.

## D82: Flat-row selection index is re-clamped wherever the effective row set can shrink — M4.7

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: code review (TRIAGE T23), high severity — `f.selected` is a raw index shared across
  Tree rows and both flat-list row sets, but only `move()` re-clamps it. `cycleMode`/
  `cycleScope` never touch it, and `setRoot`'s M4.6 memory-restore computes it by walking
  `visibleRows()` (Tree-only) even when restoring into Recent/Starred. Selecting a lower Tree
  row then cycling to a shorter Recent/Starred list — or un-starring the last-visible selected
  row — leaves `f.selected` pointing past the end of the new row set, so `selectedFlatRow()`
  returns `ok=false` and Activate/Peek/Star silently no-op with no footer message, undermining
  M4.5's and M4.6's own acceptance wording. A related stale doc comment (TRIAGE T24, on
  `selectedFlatRow`) claims Peek/Edit/DV2-targeting still go through `selectedNode()` "until
  M4.5 gives them the same fix" — no longer true since M4.5 landed `togglePeekFlatRowInto`.
- Decision: accepted as a new task, M4.7 — clamp/re-resolve `f.selected` against the current
  mode/scope's `rowCount()` at every point that can change the effective row set
  (`cycleMode`, `cycleScope`, `setRoot`'s flat-mode restore, and any row-set-shrinking store
  write reachable from a flat list, i.e. un-starring the selected Starred row), mirroring the
  `nearestSurvivingIndex` treatment Tree refresh already gets; and correct T24's stale comment
  in the same pass since it documents the state this task changes.
- Why: high severity and directly reachable in normal use (switch mode with a lower Tree
  selection, or un-star the last visible Starred row) — a real functional bug in delivered
  work (M4.5, M4.6), not a theoretical gap, and it silently no-ops with no error message,
  which is worse than a visible failure. Not blocking the current cursor (M5.1 has no
  dependency on flat-row selection correctness), but cheap enough (effort 3) and important
  enough to fix before building more on top of the flat-list machinery. TRIAGE T21 — kept open
  by build only for replan to confirm/clear — is dropped here as already-handled: the actions
  themselves (Activate/Peek/Star/Edit, Expand/Collapse-as-no-ops) are all correctly wired per
  D80/M4.5; the remaining gap is precisely T23's index-clamping bug, not a T21-shaped one.

## D90: M6.1's history mechanics — key chords, launch seeds history, boundary is silent

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: M6.1 (per-viewer document history, S9) left several small implementation choices
  unspecified by SPEC.md/PLAN.md: (1) which two keys drive Back/Forward; (2) whether the
  launch-time document (`applyResolution`'s `openDocument`) seeds history entry 0, or only
  later Activates do; (3) whether hitting either end of history (nothing to go back/forward
  to) is a silent no-op or a footer message; (4) how to keep `openDocumentInto`'s ~39
  existing call sites (mostly test fixtures) untouched while still recording history only at
  the "real navigation" ones.
- Decision: (1) `[`/`]` (HistoryBack/HistoryForward), unused elsewhere in the keymap and free
  of the Picker Prompt's always-typing constraint since neither the Files Nav nor a focused
  viewer types. (2) Launch seeds history like any other Activate (`openDocument` routes
  through the same `model.navigate` wrapper) — browser parity (the first page load is history
  entry one) and simplest to implement uniformly. (3) Silent no-op at either boundary, no
  message — mirrors FindNext/FindPrevious's existing "nothing to do" shape for an empty match
  list, not activate()'s "silence is not an error" refusal-with-message shape, since walking
  off the end of history is not a refusal of anything the user asked for. (4) A new
  `model.navigate(p, d)` wrapper around the unchanged `openDocumentInto`, called only from the
  Files Nav's `activateInto`/`activateFlatRowInto`, the Picker's `activatePickerInto`/
  `activateGrepPickerInto`, and `openDocument` — `openDocumentInto` itself and every one of its
  other ~39 call sites (finishEdit, `walkViewerHistory`'s own back/forward re-open, and every
  existing test fixture) are completely untouched.
- Why: (1)/(2)/(3) are low-risk product-shape judgment calls with no acceptance-criterion
  wording to contradict; (4) avoids a 39-site mechanical signature change for a two-line
  wrapper, keeping the diff scoped to the five call sites that actually changed behavior.

## D91: M6.2's scroll memory lives in `viewer.open()` itself, not only in the history walk

- Status: decided
- Cycle: C4
- Date: 2026-08-12
- Context: M6.2 (per-document scroll memory, S9) is asked for in terms of history's own Back/
  Forward, but every path into a document — a plain Activate, the Picker's Activate, launch,
  and history's own `goBack`/`goForward` re-open — already funnels through the single
  `viewer.open()` method (M6.1's own `navigate`/`walkViewerHistory` design). Putting the new
  `docOffsets` map's save-before-leaving/restore-on-arrival logic there, rather than only in
  `model.walkViewerHistory`, covers all of them for free instead of duplicating the same pair
  of map operations at each call site. One visible side effect: re-Activating a document that
  is already the Active Document (no history motion at all) now preserves its scroll position
  instead of `open()`'s previous unconditional `GotoTop()` — since leaving and re-arriving at
  the same document banks and immediately restores the same offset, this is a no-op by
  construction, not a special case.
- Decision: implement scroll memory inside `viewer.open()`, keyed by path, rather than adding
  a second, history-specific code path. `refreshDoc`/`close` are deliberately left untouched —
  neither one ever leaves a document (S2's own preserve-scroll contract; a vanished-file close
  has nothing worth remembering), so `open()` remains the only place a document is ever left.
- Why: the acceptance criterion ("remembered per document rather than per history entry") is
  a property of the document, not of the history mechanism specifically, and the reuse avoids
  a second, parallel bookkeeping path with its own chance to drift out of sync with the first.
  The one behavior change (same-document re-Activate no longer resets to top) is untested
  today, matches the same "remembered per document" spirit the acceptance wording asks for,
  and is strictly a smaller surprise than the alternative (Activate keeps a reset-to-top
  special case that Back/Forward on the very same document does not).

## D92: The north star dilates — noggin is a browser of document collections

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: C4a was pitched as a Wikipedia reader, which the existing north star ("several
  sources of *notes and docs*") did not admit: Wikipedia is remote, read-only, unbounded and
  pathless. The original framing was a deliberately narrow hedge against the value not
  materialising; four delivered cycles have since made the value credible.
- Decision: widen the north star to **a browser of document collections** — a thoughtspace
  for navigating, synthesizing and remembering knowledge from several collections at once.
  Collections are admitted for a specific value they carry (Notes = synthesis target and
  personal information; Wikipedia = authoritative reference; Dictionary = the medium
  everything else is written in), never as a clearinghouse. Starring and link-following are
  the cross-collection features where value compounds.
- Why: "everything I use to think" is messy and scales irregularly; "a browser of document
  collections" is tightly scopeable feature-wise yet grows naturally to man pages, cheat
  sheets, docsets and dictionaries. Consumption and navigation stay the whole job.

## D93: LLM chat is cut from the vision — creating documents is outside noggin, like editing

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: A TUI LLM chat had been floated as a peer source alongside Notes and Wikipedia.
  Interrogating it against the dilated north star showed chat is not a collection: it
  *creates* documents rather than serving them, and needs create-new, model selection,
  live-mutating content and an input field that owns the keyboard.
- Decision: cut chat as a noggin feature. The same editor boundary applies — hand off to a
  dedicated chat TUI, and read its saved transcripts back as an ordinary collection. A
  handoff key plus transcript-reading sits on ROADMAP Unscheduled.
- Why: chat is a solved problem with good existing TUIs; absorbing it would import an
  authoring interaction model into a consumption-only product and break the Root abstraction.

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

## D95: C4a is a suffixed interstitial key; C5 organize is not renumbered

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: This cycle was inserted ahead of the already-slotted `C5 organize`. Renumbering was
  briefly agreed in conversation, then withdrawn: C5 is already referenced by intent in
  `DECISIONS.md` ("Tags are C5", twice), `knowledge/frontmatter-rendering.md` and two archived
  cycle docs — and D42 already settled this exact situation for C2a.
- Decision: allocate **C4a** and insert its row ahead of C5 in the ledger; C5 and C6 are
  untouched, with `C6 collections` newly allocated.
- Why: consistency with D42 and the never-renumber rule. Same accepted cost as D42: `C4a`
  reads as a follow-on to `recall`, which it is not.

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

## D98: CC BY-SA attribution for rendered Wikipedia articles

- Status: decided
- Cycle: C6 (opened C4a 2026-08-14, resolved here)
- Date: 2026-09-08
- Context: SPEC OQ2's original framing ("attribution obligations appear to attach to
  redistribution rather than purely local personal use") was checked against Wikipedia's own
  reuse guide (`Wikipedia:Reusing_Wikipedia_content`) and the CC BY-SA 4.0 legal code during
  C6 build. Two findings: (1) a URL/hyperlink back to the article is one of CC BY-SA's own
  explicitly-sufficient options for author attribution, but attribution alone is not the full
  obligation — a licensing notice naming the license, and an indication that the content was
  modified (noggin's HTML→Markdown conversion is a modification, not a verbatim copy), are
  also required; (2) the "purely local personal use is clear" assumption does not hold as
  confidently as OQ2 hoped — CC BY-SA's "Share" trigger is defined broadly (reproduction,
  public display, communication), and comparable reader apps (Kiwix, Wikiwand) attribute even
  ephemeral, non-persisted, single-user renders, so noggin should not rely on that exemption.
- Decision: **Amended SPEC.md with S9a** — the Wikipedia-article location line reads
  `CC BY-SA 4.0 · <canonical URL> · Reformatted to markdown`, covering all three obligations
  (attribution via the URL, a licensing notice, and the modification indication) on every
  article view, at no extra row cost (S9 already reserved the line). Closes OQ2/D98.
- Why: cheap to satisfy fully now (one line, already-planned real estate) versus leaving a
  live legal-compliance gap open into a feature this cycle is building for real use, on an
  assumption ("local display is exempt") that turned out weaker than believed.

## D112: C4a spike verdict — GO, with a known limitation

- Status: decided
- Cycle: C4a
- Date: 2026-09-03
- Context: SPEC S7a's cycle-ending obligation — an explicit go/no-go/retune-exhausted
  verdict against the user's A1–A5 readability bar, informed by M3.2's per-article A1–A4
  read, S6a's two costed experiments (maths M3.4, images M3.5), and every retune task since
  (M3.3a, M3.3b, M2.4a, M3.5a, M3.7, M3.8, M3.9). Full detail:
  `cycle/findings/m4.1-conversion-record.md` (consolidated verdict record, re-verified
  post-M3.9), `cycle/findings/sample-judging.md` (M3.2's per-article A1–A4 verdict),
  `cycle/findings/maths-experiment.md`, `cycle/findings/images-experiment.md`.
- Decision: **GO — with a known limitation.** All four S5 samples clear A1 (readable), A2
  (headings/structure) and A4 (tables); three of four clear A3 (no gaps/artifacts) outright.
  `E (mathematical constant)`'s residual A3 shortfall — 39% of its formulae fall back to
  raw-TeX source text sitting mid-sentence, per `maths-experiment.md` — is confined to that
  one article and falls squarely within A5's own worked example (SPEC.md: "if
  `E (mathematical constant)` reads as swiss cheese while the other three clear A1–A3, that
  is a *go with a known limitation*, not a no-go"). The user's own bar ("I would read this
  here instead of opening a browser") is cleared. This gates `ROADMAP.md`'s "Wikipedia as a
  real collection inside noggin" and "A dictionary / thesaurus collection" Unscheduled
  entries and cycle `C6` (collections) — all three may now proceed to their own specs. This
  decision commits to none of their shape: per S6a/S7a, "nothing decided here fixes the
  shape of the eventual real feature; a later cycle with its own agreed spec decides whether
  and how either ships," and that stance extends to the whole verdict, not just the two named
  experiments.
- Why: the retune budget closed on legitimate diminishing returns, not on running out of
  cycle — M3.2's single-pass A1–A4 read found exactly one shortfall (E's maths gaps, carved
  into its own experiment), and every subsequent shortfall discovered mid-experiment or
  mid-review was fixed in its own task except three accepted known limitations (D107's
  half-open-interval gap, D108's cross-paragraph closer-masking, D110's sister-project
  scope-widening — none observed in any S5 sample) and one that was fixed one task later
  than the rest (M3.9, the maintenance-template/`{{rp}}` leak, D111). Both S6a experiments
  independently recommend carrying their approach forward (`maths-experiment.md`,
  `images-experiment.md`), so the swiss-cheese risk that originally widened this cycle's
  scope (D104) is substantially addressed rather than merely documented. A *go* verdict
  fixes nothing about C6's eventual shape (root/pane/key/search/collections model, per
  SPEC's own Out-of-scope list) — it only certifies that the underlying conversion quality
  clears the bar worth building on. OQ1/D97 (whether the spike's conversion code itself
  survives the cycle) is unblocked by this verdict but still decided at `/charm-close`, per
  its own entry, not here.


## D113: This cycle is C6 sources; C5 organize is withdrawn, not reused

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: Priorities moved: the collections work should run before tags/outline. The user
  first proposed reusing the unstarted `C5` key for it. But archived docs and D95 already say
  "Tags are C5", and D94 had already scheduled the collections question as `C6`.
- Decision: allocate this cycle as **C6 sources**; mark `C5 organize` **withdrawn** in the
  ledger with its goals moved to Unscheduled; the C5 key is retired and never reused.
- Why: least disruptive — nothing that points at C5 or C6 needs re-pointing, and the
  never-renumber rule holds even for unstarted keys. Ledger order remains independent of key
  order (D42, D95).

## D114: Sources, not Collections — Root stays, Query Source is the new class

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: C4a left "Collection" provisional and D94 left open whether a typed Root is "just a
  source with conditional UX, or needs a distinct layer". The user proposed renaming
  collection → source, root → file source, Wikipedia/dictionary → query source.
- Decision: **Source** is the umbrella entity (Roots Nav → Sources Nav). **Root** keeps its name
  and meaning. **Query Source** is the second class: singular per kind, always present, never
  created/removed/starred, no Tree, no grep, reached only by search. "Collection" is retired
  as an in-app term and kept only as north-star language. D94's open question resolves to:
  a typed Source with conditional UX, no distinct layer.
- Why: "Root" is load-bearing across the glossary (Active/Ad-hoc/Broken Root, Orphan, All
  Roots) and is exact for a folder; renaming it buys nothing. The unifying observation is that
  neither class persists its content — only pointers, stars and timestamps — and the real axis
  of difference is bounded-and-crawlable vs. not, which is what dictates Tree and grep.

## D115: Query Source documents in Recent are treated exactly like files

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: The pitch suggested a Wikipedia-specific time/count cap or session-scoped Recent.
- Decision: no special treatment — same persistence, per-Source cap, recency-only ordering.
- Why: Recent is the only way back to an unstarred article other than re-searching, which
  makes persistence more valuable for articles, not less; a divergent rule would also break the
  "restate for the dictionary without rewording" test the spec is built around.

## D116: Picker minimum query length is 2 (revises D31's 3)

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: D31 fixed the File/Grep Picker threshold at 3 characters.
- Decision: 2 characters in every Picker mode including Wikipedia; searches remain debounced,
  with a longer wait in Wikipedia mode because each search is a network call.
- Why: user judgment that 3 feels laggy; the existing debounce (D41) already prevents
  per-keystroke work, so the threshold was carrying less than it appeared.

## D117: Root parent-path fragment dropped from Picker rows; editable Root names replace it

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: D22 disambiguated same-named Roots in Picker rows via a left-truncated parent path.
- Decision: drop the fragment (row = Root name + path relative to the Root). Persisted Roots
  gain an editable name (default: folder base name); the Sources Nav's path line remains the
  disambiguator there. Ad-hoc Roots have no name until starred.
- Why: the fragment costs width on every row to solve a rare case that a user-chosen name
  solves better. Accepted loss: two same-named, un-renamed Roots are indistinguishable in
  Picker rows.

## D118: Focus indicated by background — inactive panes lighter, as an evaluation

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: Focus is currently border colour only (205 vs 240, 256-colour palette, no
  backgrounds anywhere). The user wants to evaluate background-based focus before the roadmap's
  style pass. "Active pane darker" was the first framing.
- Decision: tint **unfocused** panes with a subtly lighter background; the focused pane keeps
  the terminal's own background; border cue retained. Provisional until the style pass.
- Why: darkening the focused pane is a no-op on already-black terminals; lightening the
  others is visible regardless of theme. Known risk for build: Glamour output carries its own
  ANSI resets, so an inactive Document Viewer may show tint "holes" — acceptable in an
  evaluation.

## D119: Articles carry no Preview and are never grepped; File·Starred is the only cross-class search

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: Each article Preview or grep would require a network fetch (or persisted content,
  which contradicts D114's pointer-only rule).
- Decision: Wikipedia Picker shows a placeholder instead of a Preview; Grep never includes
  Query Sources in any scope; the File Picker's `All Roots` scope is files only; `Starred` scope
  matches starred articles by title. With a Query Source active, File/Grep skip the
  `Active Root` scope rather than offering an empty one. Cross-Source *lists* in the Files Nav
  (`All Starred`, `All Sources`) do include articles — those are store lookups, not fetches.
- Why: keeps every network call user-initiated and explicit (search, Activate, Peek, Refresh);
  no dead-end scopes.

## D120: The C4a spike's conversion pipeline moves into an internal package; `cmd/wikispike` stays as a thin CLI

- Status: decided
- Cycle: C6
- Date: 2026-09-07
- Context: D97 kept `cmd/wikispike` untouched and isolated ("never imported by cmd/noggin")
  until a cycle spec'd a real Wikipedia feature. C6 does; S3 requires the conversion at the
  quality D112 certified, and `cmd/noggin` must import it.
- Decision: move the fetch-agnostic transforms (math, images, chrome strip, links, convert,
  cleanup) into an `internal/` package alongside the new search and canonical-fetch code
  (PLAN M2). `cmd/wikispike` is not deleted: it becomes a thin CLI over the package, since the
  scratch-root judging workflow from C4a remains the cheapest way to eyeball conversion
  quality across many articles. No duplicate copy of the pipeline survives.
- Why: D97 explicitly deferred extraction to "when a future cycle actually specs a real
  Wikipedia feature"; the isolation rule it preserved has served its purpose. Keeping the CLI
  costs a few lines and keeps the spike's `wikispike <title> <out>` recipe (and the
  `knowledge/wikipedia-*` findings that cite it) usable for regression judging.

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

## D122: M2.3's fetch cache only hits for already-canonical requested titles — is that guarantee strong enough?

- Status: open
- Cycle: C6
- Date: 2026-09-08
- Context: `internal/wikipedia.Client.Fetch` (M2.3) derives its cache *lookup* key from the
  requested title before the request, but its cache *write* key from the canonical URL the
  response resolves to. A fetch for a title that turns out to be a redirect therefore never
  hits cache on repeat, even within the same session — only fetches of a title already
  canonical do. All five of M2.3's own done-when test scenarios pass under this behaviour
  (surfaced by code review, TRIAGE T13, and independently confirmed at validate). PLAN.md's
  prose ("caches the result... keyed by canonical URL") reads as a stronger guarantee than
  this. S3 only requires this to hold for titles noggin itself records, since Starred/Recent/
  Refresh always reopen by canonical title (per M2.3's own PROGRESS.md rationale) — but that
  is currently an implementation comment, not a tested invariant at any layer.
- Decision: **Open.** Not blocking M2 or M3; revisit once M4.2 (async fetch-and-open) wires a
  real caller and it's clear whether any caller path can hand `Fetch` a non-canonical title
  more than once per session. If so, decide whether to widen the cache key (e.g. a secondary
  index by requested title) or accept the narrower guarantee as by-design.
- Why: Cheap to defer — no user-visible behaviour depends on it yet, and M2.3's tests already
  pin the current (narrower) contract, so nothing regresses silently while this is open.

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

## D124: `Search`'s package-level client vs `Fetch`'s `Client` struct — no shared config surface

- Status: decided — revisited at M3.3, kept separate
- Cycle: C6
- Date: 2026-09-08
- Context: TRIAGE T9 (code review). `internal/wikipedia.Search` is a free function backed by
  a package-level mutable `searchHTTPClient`; `Fetch` is a method on an explicit `Client`
  with its own `http.Client`. Nothing today unifies timeout, transport or User-Agent
  configuration between the two call shapes.
- Decision: **Open.** Not blocking M2 or M3.1/M3.2 (M3.3 is the first task that wires a real
  injectable client into the Picker for search, and M4.2 does the same for fetch) — revisit
  at M3.3, when it will be clear whether the Picker's injectable-client need pushes toward
  unifying `Search` onto `Client` (or a shared config struct) before more call sites accrete.
- Why: Cheap to defer — both shapes work and are independently tested; deciding now would be
  speculative ahead of M3.3's actual caller shape.
- Revisit (M3.3): The Picker's injectable-client need turned out not to require unification.
  `cmd/noggin`'s `pickerWikipediaClient` interface asks only for a `Search(query, limit)`
  method, so `*wikipedia.Client` satisfies it with a one-line delegating method
  (`internal/wikipedia/search.go`) that calls straight through to the existing package-level
  `Search`/`searchHTTPClient` — no shared `http.Client`, timeout or transport needed. Kept
  separate: search.go's own documented reason (`searchHTTPClient`'s doc comment — Fetch's own
  timeout is scoped to Fetch's own acceptance criterion) still holds, and forcing a shared
  config now would be speculative ahead of M4.2 actually needing one. Revisit again only if
  M4.2's own fetch-client wiring surfaces a concrete need (e.g. a single configurable
  User-Agent/proxy for both calls); until then this stays decided, not open.

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

## D126: M3.4's Document Store key for a Wikipedia Result never yet fetched is a title-derived URL, not a confirmed canonical one

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: M1.5/M4.1 fix the Document Store's article key as the canonical URL, but M3.4's
  Picker star must act on a `wikipedia.SearchResult` (title search, M2.2), which carries no
  URL at all — only `Fetch` (M4.2) learns the canonical URL, by following redirects. The
  handoff into M3.4 (PROGRESS.md) flagged this as needing "a concrete answer" before assuming
  title is an adequate key.
- Decision: Star a Wikipedia Result by deriving its key from the title alone via a new
  exported `wikipedia.ArticleURL(title)` — the same escaping (`articleURLFor(wikiBase,
  title)`) `Fetch`'s own request-building and `canonicalFromURL` already use for an
  already-canonical title, so the derived URL is byte-identical to what a real Fetch of the
  same (non-redirect) title would resolve to (locked in by
  `TestArticleURLMatchesCanonicalFromURLEscaping`, `internal/wikipedia/fetch_test.go`). A
  title that turns out to be a redirect page (title search matches real page titles, so this
  is rare) gets a record keyed by the pre-redirect URL; a later Fetch (M4.2) resolves the true
  canonical URL under what would be a second, unmerged record — not reconciled by this task.
- Why: Correct for the overwhelmingly common case at zero extra network cost (no fetch-before-
  star), reuses existing escaping logic instead of a second hand-rolled one that could drift,
  and mirrors D122's own accepted precedent — "noggin's own records always hold the canonical
  title once an article has been opened", i.e. before that point a narrower guarantee is
  already the codebase's own accepted shape, not a new relaxation invented for this task.
  Fetching first would also block Star on the network, which S2 does not ask for.
- Revisit: if M4.2's real fetch usage or a future validate surfaces the redirect-title edge
  case as a live problem (e.g. a starred article silently "loses" its star to Activate landing
  on a different key), decide there whether to reconcile keys on fetch (migrate the derived-URL
  record to the resolved canonical one) rather than leaving two.

## D127: `docRecord`/`withDocRecord` route through a new `docStoreKey` helper instead of calling `filepath.Clean` directly

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: Found while implementing M3.4. `model.go`'s `docRecord`/`withDocRecord` — the only
  read/write path for the S1 document store, shared by every star/recency caller including the
  Picker's Wikipedia branch this task adds — unconditionally ran every key through
  `filepath.Clean`, written when only file paths existed. `filepath.Clean` collapses a URL's
  `"//"` scheme separator into `"/"` (confirmed:
  `filepath.Clean("https://en.wikipedia.org/wiki/X")` → `"https:/en.wikipedia.org/wiki/X"`),
  so the first article record M3.4 ever wrote through this path would have been silently
  corrupted on every read/write — a latent bug since M1.5 introduced the article key shape,
  surfaced only once a real caller (M3.4) exercised it.
- Decision: Added `docStoreKey(path string) string`, which returns an article's canonical URL
  untouched (detected by an `"://"` substring, absent from any file path this codebase itself
  constructs) and `filepath.Clean`s everything else, and pointed both `docRecord` and
  `withDocRecord` at it instead of the bare `filepath.Clean` call.
- Why: Directly blocked M3.4's own acceptance criterion (a starred article's persisted record
  would not survive its own key). Low-risk, narrowly scoped fix (one helper, two call sites,
  no change to file-record behaviour — `docstore.go`'s on-disk load/save path already
  branches by an explicit `<kind>` field, unaffected). Fixing in place rather than deferring to
  TRIAGE since it sits on M3.4's own critical path, not an unrelated tangent.

## D128: M3.5 reuses `pickerFile` (not a new type) to represent a starred article in File mode's Starred scope, with the two now-reachable readers explicitly guarded

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: S6/M3.5 needs File mode's Starred scope to include a starred article alongside
  starred files ("the only File search that spans both classes"). `pickerFile` already carries
  an `orphan` flag for one cross-cutting Starred-scope case (M5.1); M3.5 adds `article`/`title`
  the same way rather than introducing a second candidate type or a discriminated union.
  Fresh-context code review (code-reviewer subagent) flagged this as pushing the "is this
  actually a file" invariant onto doc-comment discipline alone (medium), and, more
  concretely, found that two existing `pickerFile` readers this reuse newly makes reachable —
  `activatePickerInto`'s fileMode branch and `syncPreview` — were not taught about `article`,
  so selecting or activating an article row surfaced a raw filesystem error
  ("cannot open https://...: no such file or directory") rather than anything resembling
  intended behaviour (high — an observably broken interaction, not just an untested one).
- Decision: Kept the `pickerFile` reuse (a discriminated-union refactor is a larger change
  than M3.5's own scope justifies, and the doc-comment-discipline risk is real but not
  fundamentally different from `orphan`'s own precedent). Fixed the two reachable readers
  directly, since M3.5 is what made them reachable: `activatePickerInto` refuses with a plain
  Footer message ("<title>: opening a Wikipedia article from here isn't supported yet") and
  dismisses the Picker, the same "message + dismiss" shape a genuine read-error Activate
  already has; `syncPreview` shows a fixed placeholder string instead of attempting
  `os.ReadFile` on the URL. Neither fetches or renders the article — that is still M4.2's job
  — but neither surfaces a filesystem-error string for a thing that was never a file.
- Why: The full fix (fetch-and-render from this call site) is M4.2's own acceptance
  criterion, already covering "Activating a Wikipedia Result"; widening M4.2 to also read
  "from any Picker mode/scope that can produce an article row" rather than reopening M3.5 or
  building fetch-and-open early keeps the milestone boundaries the plan already drew. Leaving
  the interim state as a raw OS error, though, was not an acceptable degrade for something
  M3.5 itself makes newly selectable — a clean, honest refusal costs two small guards and
  removes a confusing failure mode without pre-building any of M4.2's actual work.
- Revisit: at M4.2, confirm its acceptance criteria are written broadly enough to cover
  Activate from File mode's Starred scope (this task's own new path) and not only
  `wikipediaMode`'s Results list — replace both guards added here with the real fetch/render
  behaviour at that point, not before.

## D129: M4.1's `document` reuses `path` as a dual-purpose file-path/article-URL field, with a viewer-scoped `articleCache` standing in for M4.2's not-yet-built fetch

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: M4.1 needs `document` (and everything keyed by its identity — `viewer.history`,
  `docOffsets`, S9 Document History's goBack/goForward) to work for either a file or an
  article without a network fetch, which is M4.2's own job. `docStoreKey` (D127, model.go)
  had already generalized the document *store*'s key to hold a file's absolute path or an
  article's canonical URL, told apart by a `"://"` substring rather than a threaded bool.
- Decision: `document.path` is reused the same way for the in-memory type (`isArticleKey`,
  shared by `document.go` and `docStoreKey`, one definition); `document` gains only a
  `title` field (an article's display name — `filepath.Base` on a URL would show a slug, not
  the real title) and `isArticle()`/`label()` helpers. History/`docOffsets`, both already
  keyed by `.path`, needed no structural change. The one real gap — `model.walkViewerHistory`
  re-reads a file fresh from disk on Back/Forward, but there is nothing on disk for a URL —
  is filled by a new `viewer.articleCache map[string]document`, populated by `viewer.open`
  alongside `docOffsets` and consulted instead of `loadDocument` when the history entry is
  URL-shaped. Since every path ever in `v.history` was first opened through `viewer.open`
  (`navigate`'s only underlying call, per its own doc comment), the cache can never miss in
  practice — a miss is treated as a broken invariant (a Footer message, not a panic), not a
  design gap to route around.
- Why: This keeps M4.1 to exactly its own scope — no `wikipedia.Client` call anywhere in this
  task, matching D119's "every network call is user-initiated (search, Activate, Peek,
  Refresh)" by omission (History walking was never on that list, and should not silently
  become network-triggering by accident). It also means M4.2's fetch-completion handler,
  whatever shape it takes, only has to call the existing `navigate`/`open()` path once per
  fetch — walking back to an already-seen article afterwards is then free, no re-fetch, no
  new code.
- Revisit: at M4.2, confirm the fetch-completion handler does route through
  `navigate`/`openDocumentInto` (as Activate/Peek's existing file paths do) rather than
  writing `dv1.doc`/`articleCache` directly — that is what keeps this task's "a URL can never
  reach `v.history` without a matching `articleCache` entry" invariant true going forward.
  Fresh-context review (code-reviewer subagent) also flagged two related landmines, both
  already covered by M4.3's own acceptance criterion rather than this task's: once M4.2 lands,
  `R` on an open article would call `loadDocument`'s disk-read path and fail loudly instead of
  refreshing (S3's "R re-fetches any open Wikipedia article"), and `e` would hand `$EDITOR` a
  URL as a literal filename (S3's "e on an article is refused"). `staleness.go`'s `os.Stat` on
  a URL already fails and is silently omitted by its own pre-existing "omit, don't
  half-track" rule, which happens to already satisfy S3's "staleness indicator never fires for
  an article" — noted in a code comment there so it is not mistaken for a bug and "fixed" into
  tracking a URL.

## D130: M4.2's "Peeking it" is a Files-Nav-triggered Peek (M5.2's own job to wire up), not a new Picker-internal Peek key; the fetch mechanism itself is built generically now and exercised directly by this task's tests

- Status: decided
- Cycle: C6
- Date: 2026-09-08
- Context: PLAN M4.2's own text ("Activating a Wikipedia Result or Peeking it") reads, taken
  in isolation, as if Peek should be reachable on a Wikipedia Picker Result. But S2 is explicit
  that "There is no Preview: the Preview pane shows a short placeholder instead of article
  content" for wikipediaMode, and GLOSSARY.md's own Peek entry defines it as "render the Files
  Nav's selected document ... Distinct from the Picker's Preview pane" — Peek has never been a
  Picker concept, and a new one could not use the `p` key without colliding with the Picker's
  own always-typing Prompt (handlePickerInput). Re-reading S3 ("Activating a Wikipedia Result, a
  Starred or Recent article row, or Peeking one") resolves this: "one" refers back to "a
  Starred or Recent article row" (a Files Nav flat-list row, M5.2's own deliverable), not "a
  Wikipedia Result" — Activate reaches both surfaces, Peek only the Files-Nav one, matching
  Peek's Files-Nav-only nature exactly.
- Decision: `model.fetchArticleInto(target pane, title string, peek bool) (model, tea.Cmd)` and
  its completion half `applyArticleFetch` (article_fetch.go) support Activate and Peek
  semantics symmetrically, but only Activate is wired to a real key this task (the Picker's own
  wikipediaMode Result and File mode's Starred-scope article pseudo-candidate, both via
  `activatePickerInto`). Peek's half is exercised directly by this task's own model tests
  (`article_fetch_test.go`) with no UI trigger yet, the same "test the mechanism ahead of its
  caller" shape M4.1's `articleDocument` already set for this cycle (document.go's own doc
  comment) — M5.2 is expected to call `fetchArticleInto(..., peek: true)` from a real Files Nav
  Peek keypress on an article row once that row exists.
- Why: Building a second, Picker-scoped Peek concept now would contradict S2's explicit "no
  Preview" rule and GLOSSARY's settled definition, for a UI surface (Files Nav article rows)
  that does not exist until M5.2 regardless — there would be nothing to test the keybinding
  against yet. Building the underlying mechanism generically now, ahead of that keybinding,
  keeps M4.2 to its own effort estimate while giving M5.2 nothing further to invent.
- Revisit: none expected; M5.2 should confirm this reading holds once Files Nav article rows
  land, and wire its Peek keypress straight to `fetchArticleInto(..., peek: true)`.

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
