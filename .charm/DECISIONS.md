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
