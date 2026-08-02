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

## D8: Pre-agreed cut order if C1 overruns

- Status: decided
- Cycle: C1
- Date: 2026-07-30
- Context: Twelve deliverables against a one-week appetite.
- Decision: Cut in this order, most expendable first: S12 (help modal), S9 (navigator toggle),
  S8 (find-in-document).
- Why: Deciding the sacrifice while calm is better than triaging it on day five. Everything
  above those three is what makes the dogfood week possible at all.

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
- Why: low risk and cheap to revisit \u2014 D2 confines every key literal to keymap.go, so
  retuning this one binding later is a one-file, one-line change; no plan or test depends on
  the specific key, only on the binding named `keys.Navigator` existing.

## D26: M5.3 Help Modal is real cell compositing via lipgloss v2's Canvas/Layer, not a
full-screen replacement

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
  second, centered, smaller layer (z1), and renders the compositor \u2014 the pane underneath
  stays pixel-visible around the modal's edges. The modal itself is content from
  `bubbles/help`'s `FullHelpView` (going with the grain, D1) fed `helpBindings(pane)`, a
  fuller-than-footer list that includes what `paneHints` deliberately omits (PrevPane; for
  the viewer, Cancel/FindNext/FindPrevious \u2014 see footer.go's comment). Input is trapped the
  same way `handleFindInput` traps an in-progress search: while `m.helpOpen`, only Help or
  Cancel do anything, everything else (including Quit) is swallowed. Fresh-context review
  (code-reviewer subagent) caught that a Compositor sizes its output canvas to the union of
  every layer's bounds, not the base layer's own size \u2014 an unclamped modal taller/wider than
  the terminal (e.g. the viewer's full binding list on a short terminal) would grow the
  rendered frame past what `View()`'s no-alt-screen drawing can show correctly. Fixed by
  wrapping the modal box in `lipgloss.NewStyle().MaxWidth(width).MaxHeight(height)` before
  compositing, so it is truncated rather than overflowing; `TestHelpOverlayNeverExceedsTheTerminal`
  locks this in.
- Why: true compositing is the only way to keep the pane visible around the modal, which is
  what "opens over" means; going with the grain (D1) again rather than hand-splicing strings.
  The canvas round-trip re-serializes SGR runs and trims trailing padding on unaffected rows
  \u2014 cosmetically identical on a real terminal, but byte-different, which is why
  `help_test.go`'s overlay test compares `ansi.Strip`'d, trailing-space-trimmed lines rather
  than raw bytes. This round-trip only ever runs while the modal is open (`render()` skips it
  otherwise), so it cannot affect any other milestone's byte-exact frame comparisons.

## D27: C1's replan (2026-08-02) routes the M5 review backlog into two new tasks, one
deferral and one drop

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

## D28: T28's fix distinguishes a directory argument from a file argument, not just
absolute from relative

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

## D29: C1's replan (2026-08-03) routes the M5.4/M5.5-diff review backlog into two new
tasks and one direct documentation fix

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

## D30: C1's replan (2026-08-03) routes the M5.6 validate/review follow-up (T44/T45) and one
review test-gap (T46) into one new task; T44 dropped as superseded

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
