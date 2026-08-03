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

## D7: The starred-roots store lives in the config directory, hand-editable

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: It is arguably application data, not user configuration.
- Decision: Treat it as config for now, in a plain hand-editable format.
- Why: While iterating, being able to fix a moved or broken root in an editor is worth more
  than taxonomic correctness — it is also the only repair mechanism, since in-session root
  management is out of scope. Move it to a data location once it stops needing hand edits.

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

## D14: Adopt Charm v2 now, at the skeleton, not later

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: C1 was planned against Bubble Tea v1.3.10 because v2 was pre-release at the time
  (`PROGRESS.md` handoff). It is not any more: `bubbletea/v2` v2.0.8, `bubbles/v2` v2.1.1,
  `lipgloss/v2` v2.0.5, `glamour/v2` v2.0.1 are all stable, under a new module path
  (`charm.land/…`, not `github.com/charmbracelet/…`). bubbles v2 and lipgloss v2 hard-require
  bubbletea v2, so this is one all-or-nothing swap.
- Decision: Migrate all four to v2 as a new task M1.4, ahead of M2.1, before any further
  surface is built. Re-scoped M5.1 down (3 → 2) because `viewport` v2 ships
  `SetHighlights`/`HighlightNext`/`HighlightPrevious`/`EnsureVisible`. D13's rule ("no terminal
  query after `Run()`") stands unchanged; only its cause is removed.
- Why: Cost scales with surface, and surface is at its minimum — three tasks, ~600 lines of
  non-test code, and the migration is mechanical (`View() tea.View`, `tea.KeyPressMsg` with
  `Code`/`Text`, viewport getters, `lipgloss.Color` as a function). Paying it at M2.1 is
  cheapest; paying it at M5 means redoing the tree, the roots panes and the footer. Staying on
  v1 is the expensive option, not the safe one: v1 is now the dead branch, the D13 deadlock we
  worked around is designed out of v2, `viewport`'s new highlight API is most of S8, and
  `tea.WithWindowSize`/`WithColorProfile` relieve the pty-harness pain
  (`findings/pty-verification.md`). Accepted risks: a young major with visible churn (8 patch
  releases), a pre-1.0 `ultraviolet` dependency, and the pty recipe needing re-verification
  against v2's different startup queries. Full migration surface, verified against the real
  modules, in `findings/bubbletea-v2-migration.md`. This is a dependency decision, not a spec
  change — no deliverable, criterion or key changes meaning.

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

- Status: decided
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

- Status: decided
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

## D32: S5 (match highlighting) is this cycle's pre-agreed cut

- Status: decided
- Cycle: C2
- Date: 2026-08-02
- Context: Highlighting fuzzy match positions inside a styled Result row, and a matched
  substring inside a possibly-truncated grep line, is the same ANSI-offset class of problem
  C1 already lost time to (`knowledge/viewport-search-ansi-offsets.md`).
- Decision: Keep grep in scope; if the cycle runs long, drop S5 first rather than trimming the
  Grep Picker.
- Why: Naming the cut before the pressure arrives is what stopped C1 trimming the wrong thing
  (C1/D8 pre-agreed cuts, none of which were needed). S5 is polish; grep is the substance.

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

## D35: C2's Picker hand-duplicates the modal dispatch again rather than building the shared abstraction

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: C1's code review left a roadmap item: a shared exclusive-input-mode dispatch
  abstraction across modals, "only worth it if a third modal is ever added" (C1/T31). The Picker
  is that third modal, so the condition it named has arrived.
- Decision: still hand-duplicate. M2.1/M2.3 add a third `if` gate in `Update` and a
  `handlePickerInput` in the shape of `handleFindInput`, and the roadmap item stays Unscheduled.
- Why: the three modals are less alike than the count suggests — Help swallows everything, Find
  types into one field and owns the footer, the Picker types into a field *and* drives a results
  list, a preview viewport, a mode and a Scope. The common factor is about four lines of
  ordering. Refactoring toward an abstraction while the third case is still being discovered
  would generalize from guesses; if a fourth modal arrives, the shape will be evident.

## D36: M1.2's `rg --files` runs with `--no-require-git`, and the Readable allowlist is exported from `internal/tree` rather than duplicated

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M1.2 (scoped, allowlist-filtered file listing). Two small implementation
  calls surfaced that the plan didn't anticipate. (1) Verified against the sandbox's real rg
  (15.1.0): `rg --files` only honours a `.gitignore` file when the target directory is inside
  a git repository (has a `.git` ancestor) — a `.gitignore` in a plain, non-git folder is
  silently *not* applied by default. A noggin Root is just a folder, not necessarily a git
  checkout. (2) `internal/tree/policy.go`'s `isReadable` (D23's allowlist) is unexported, but
  `cmd/noggin`'s new search layer needs the identical check applied to rg's output — the same
  allowlist, not a second copy of it.
- Decision: (1) Pass `--no-require-git` on every `rg --files`/grep invocation the search layer
  makes, so a Root's own `.gitignore` counts regardless of whether it's a git checkout — this
  is what makes D28's "rg's ignore rules govern" hold for a Root in general, not just a Root
  that happens to be a git repo. (2) Export `isReadable` as `tree.IsReadable`; `cmd/noggin`'s
  `listFiles` (`files_search.go`) calls it directly rather than re-declaring the allowlist,
  keeping policy.go's "the one place it lives" comment true across the package boundary.
- Why: both low-risk and reversible (a CLI flag, an export), so logged rather than asked.
  Skipping (1) would mean a Picker in a non-git notes folder ignores its own `.gitignore` for
  no reason a user would expect, which is a worse surprise than the flag is a cost. Skipping
  (2) would mean the allowlist drifts the first time someone edits one copy and not the other
  — D23's whole point was there being one place to edit. M1.4 (grep) will shell out to rg the
  same way and should carry `--no-require-git` too.

## D37: The Picker result cap is 200, shared by both Pickers

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M1.3 (fuzzy-match and cap the File Picker's candidates). S12 requires a
  cap and a truncation flag but SPEC/plan never fixed a number, and M1.4 (grep) needs the same
  cap so both Pickers report truncation consistently.
- Decision: `pickerResultCap = 200`, defined once in `files_search.go` for M1.3's `fuzzyFiles`
  and reused as-is by M1.4's grep search.
- Why: low-risk and reversible (one constant, no stored state depends on its value) — logged
  rather than asked. 200 is generous for a real note collection's file-name matches while still
  bounding a worst-case All Roots search; revisit if M1.4's grep (line-per-match, not
  file-per-match) turns out to want a different number.
- Also this task: `github.com/sahilm/fuzzy` (already present in the module graph as a
  transitive dependency of `charm.land/bubbles/v2`, per D28) is now imported directly by
  `cmd/noggin` and promoted to a direct `require` in `go.mod` via `go mod tidy` — no new
  dependency, just its existing presence made explicit.

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

## D39: The Picker opens on ctrl+p; Cancel (esc) dismisses it

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M2.1 (Picker state plus open/dismiss). SPEC/PLAN name no key for opening
  the Picker (S1 only requires "openable from any Focus"); keymap.go's existing bindings
  (`/`, `e`, `ctrl+n`, `?`, `tab`) were all taken.
- Decision: a new global `keys.Picker` bound to `ctrl+p` opens it (`model.openPicker`), the
  same "reachable from every Focus" shape `Edit`/`Find`/`Navigator`/`Help` already have.
  Dismissal reuses the existing generic `keys.Cancel` (esc) rather than a new binding —
  `handlePickerInput` mirrors `handleHelpInput`'s shape (owns the keyboard while open, only
  Cancel does anything yet).
- Why: low-risk and reversible (one keymap.go entry, GLOSSARY.md already calls every binding
  "provisional") — logged rather than asked. `ctrl+p` is a common "quick open" convention
  (VS Code, Sublime, etc.) and collides with nothing already bound. M2.3 (Footer/Help Modal
  coverage) and M2.2 (layout) are the tasks that make the binding visible/discoverable; this
  task only wires the state machine.
