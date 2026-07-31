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
