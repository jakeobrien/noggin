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

## D52: Three low-severity review nits accepted as-is

- Status: decided
- Cycle: C2a
- Date: 2026-08-09
- Context: three low-severity, non-blocking findings from code review, triaged together at
  cycle-end replan.
- Decision (wontfix, TRIAGE T3): `filesNav.expanded`'s unbounded growth across repeated
  refreshes (never pruned for paths that no longer exist) stays as-is — already documented as
  deliberate in `refreshRoot`'s own doc comment, harmless dead weight, not worth added
  bookkeeping this cycle.
- Decision (wontfix, TRIAGE T9): `ToggleHidden` keeps its name rather than being renamed to
  match its `Picker*`-prefixed siblings (`PickerUp`/`PickerDown`/`PickerMode`/`PickerScope`) —
  churn-only with no behavior change; revisit only if the Picker-only naming convention is
  otherwise being touched.
- Decision (wontfix, TRIAGE T11): `parseGitStatus`'s assumption that `git rev-parse
  --show-toplevel`'s output and a Root's own path resolve identically (e.g. through the same
  symlinks) stays untested — a divergence degrades silently to "no markers", still S8-
  compliant, not a crash; a symlinked-Root edge case sits outside S8's explicit out-of-scope
  list but isn't worth a dedicated test absent an actual bug report.
- Why: all three are genuinely low risk and low value relative to the cost of acting on them
  now; recorded so a fresh session doesn't re-raise them as new findings.

