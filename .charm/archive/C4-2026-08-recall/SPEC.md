# Spec — C4 recall

<!-- Per-cycle. Functional, non-technical, concrete. This is the contract of intent: what the
     deliverables ARE, not how they are built. Keep it product-level so upfront errors don't
     cascade into implementation. Durable keys (S1, S2...) — never renumber. Mid-cycle
     changes go through /brainstorm amend: supersede or suffix keys, renew the agreed date;
     charm-replan then reconciles the plan. Keys are scoped to this cycle — durable docs
     referencing them qualify with the cycle key (C3/S2). -->

Status: agreed 2026-08-10

## Deliverables

- **S1** — A persisted **document store** in a data location (not the config directory),
  holding one record per document keyed by its absolute path, carrying whether the document is
  Starred and when it was last viewed. Records are **not attached to Roots**: a Root's Starred
  or Recent documents are derived by path, which is unambiguous because Roots may not nest
  (C1/D6). The starred-Roots store moves out of the config directory into the same data
  location. This is a **clean break** — no migration; an existing config-directory
  starred-Roots file is ignored and Roots are re-starred by hand, once.

- **S2** — **Starring a document**: a document can be starred and un-starred from the Files Nav
  (acting on the selected file), from a focused Document Viewer (acting on *that* viewer's
  Active Document, so DV1 and DV2 star different documents), and from the Picker (acting on the
  selected Result). Starring is persisted and survives restart. Nothing about the Root lifecycle
  destroys document stars: un-starring a Root leaves its documents' stars and timestamps intact.
  Un-starring a document while it is displayed in a Files Nav Starred list removes that row
  immediately; un-starring from the Picker leaves the Result in place with its star marker
  cleared, since removing the row under the cursor mid-search is jumpier than it is worth.

- **S3** — **Star is visible**: a star marker on Files Nav rows in every browse mode *except*
  Starred (where membership already says it), a star indicator to the left of the document
  title in each Document Viewer's title row, and a star marker on Picker rows (File and Grep)
  for a starred Result — so starring is legible even with the Navigator hidden, and while
  picking. A **broken starred document** (its path no longer exists) is still listed and still
  marked, but cannot be activated — the same treatment a Broken Root gets. Broken stars are
  therefore only visible in a Starred list, never in a Tree.

- **S4** — **Recency is recorded** when a document is Activated (in DV1 or DV2) or opened via
  the `$EDITOR` handoff. Peek does not count. Timestamps are persisted for every document
  regardless of whether its Root is starred, ad-hoc, or no longer in the session, and are never
  deleted as a side effect of a Root being un-starred. The store as a whole is bounded, evicting
  the oldest timestamps first and never evicting a starred record; eviction clears a timestamp
  and leaves the star. Ranking is **recency only** — last-viewed timestamp, no visit counts.

- **S5** — **Browse modes in the Files Nav**: `Tree`, `Recent`, `Starred`. One key cycles the
  mode and a second key cycles the scope; both are available only while the Files Nav has Focus,
  and there is **no modal picker** for either. Scope values differ per mode:
  - `Tree` — no scope. The scope key is a no-op and the scope reads as `-`.
  - `Recent` — `Active Root` | `All Roots`.
  - `Starred` — `Active Root` | `All Starred` | `Orphans`.

  `All Starred` is the whole store, including documents under no current Root; `Orphans` is
  exactly those documents under no current Root. The current mode and scope are shown on their
  own row in the Files Nav, directly below its title row (which keeps the label and the S10
  badge) — costing the file list one row of height in every mode.

- **S6** — **Flat list rows and their bounds**: a row shows the document's file name
  left-aligned and, right-aligned, the base name of the Root it belongs to — or, when it belongs
  to no session Root, the left-truncated parent path (the same treatment the Roots Nav already
  gives same-named Roots). `Recent` shows at most 100 documents per Root; `Recent`/`All Roots`
  merges each Root's list sorted by timestamp with no further cap. A recency record whose file
  no longer exists is silently omitted (unlike a broken star, which is shown — a star carries
  user intent, a timestamp does not). Every mode/scope combination with nothing to show
  displays a brief message saying so.

- **S7** — **Browse memory**: the mode is remembered per Root, and the scope is remembered per
  (Root, mode) pair, so returning to a Root and to a mode restores what you last had. This
  memory is **session-only** — every session starts on `Tree`.

- **S8** — **`Starred` as a Picker scope**: a third value on the Picker's existing scope axis
  (`Active Root` | `All Roots` | `Starred`), available in **both** the File and the Grep Picker
  so that switching mode never has to silently rewrite the scope. `Starred` scope means the
  whole store — every starred document, including orphans — with no sub-scopes; it is the
  surface that keeps orphaned stars reachable.

- **S9** — **Document history**: each Document Viewer keeps its own back/forward history of the
  documents opened in it, traversable with two keys. Browser semantics: activating a new
  document after going back discards the forward entries. The same document may appear more than
  once, but never in adjacent entries. The `$EDITOR` handoff and Refresh do not create entries.
  History may walk into documents under other Roots and **never changes the Active Root**.
  Going back or forward restores that document's most recent scroll position, remembered per
  document rather than per history entry. All of this is **session-only** for both viewers,
  including DV2's, which dies with DV2 (C3/S8).

- **S10** — **Pane focus badges**: each pane shows the key that focuses it, right-aligned on its
  existing title row (`Files` left, `[f]` right; DV2's reads `[^o]`), shown whether or not the
  pane holds Focus, and dropped entirely when the pane is too narrow to fit title plus badge.
  The four focus-key hints (`r`/`f`/`o`/`ctrl+o`) are removed from the Footer; every other
  Footer element — Refresh, Help, Quit, pane-local hints, transient messages — is unchanged.

## Out of scope

- **Wiki-links and URL following.** They require a notion of "which link is current" — the
  document cursor C1/D4 deliberately refused. Deferred to Unscheduled as one piece with the
  other cursor-dependent features (Select mode + copy), to be revisited together.
- **Shadow Roots** — a retained-but-unstarred Root class with a TTL, to hold recency for ad-hoc
  and un-starred Roots. Considered and rejected: making the store path-keyed and global (S1)
  delivers the same outcome without a new concept, a time-based behavior or a carve-out in D6's
  no-nesting rule. Not a deferral; there is nothing left to build.
- **Any cleanup dialog when a Root is un-starred.** Un-starring a Root is deliberately
  non-destructive to document state (S2/S4). Bulk un-starring is done from the Picker's
  `Starred` scope, one document at a time.
- **Frecency** — visit counts, decay, or any ranking other than last-viewed timestamp.
- **Migration** of the existing config-directory starred-Roots file (S1 is a clean break).
- **Path normalization** — two paths reaching the same file (symlinks, `/tmp` vs `/private/tmp`)
  star and record as two documents. Accepted, not mitigated.
- **Configurable caps** — the per-Root 100 and the store bound are fixed.
- **Tags, and the `Results` browse mode** `NOTES.md` pairs with them — both are C5.
- **"Follow" mode** (auto-Peek on selection change) — remains Unscheduled, from C1.
- **Persisted history** — history and its scroll memory are session-only, deliberately.

## Acceptance (cycle-level)

- A document starred in one session is still starred in the next, and its star survives
  un-starring the Root it lives under, opening it from an ad-hoc Root, or leaving that Root out
  of a later session entirely.
- A document starred while no Root covers it is still reachable and un-starrable from the
  Picker's `Starred` scope, and appears in the Files Nav's `Orphans` scope.
- From the Files Nav, two keys reach every mode/scope combination with no modal, the title row
  always says which one is showing, and returning to a Root and mode restores the last
  combination used there in that session.
- The Recent list orders documents by when they were last Activated or edited, ignores Peeks,
  and does not lose entries when Roots are un-starred.
- Starring is legible from the Document Viewer alone, with the Navigator hidden, and a
  starred Result is visibly marked in both Picker modes.
- Each Document Viewer walks its own back/forward history independently, lands at the scroll
  position last used in each document, and never changes the Active Root while doing so.
- Every pane's focus key is readable off its own title row, and the Footer no longer carries
  those four hints.

## Pre-agreed cuts

<!-- Ordered. If the cycle spills, these go first, in this order, without re-litigation. -->

1. Starring from the Picker (S2's third surface) — Files Nav and Document Viewer suffice.
2. Per-document scroll memory in history (S9) — history still works, landing at the top.
3. The `Orphans` scope (S5) — `All Starred` already makes orphans visible, just not isolated.
4. `Starred` scope on the Grep Picker (S8) — File Picker only, accepting the mode-switch wrinkle.

## Open questions

<!-- Unresolved items blocking or shaping the spec. Move to DECISIONS.md once material. -->

- None blocking. Two judgment calls are logged rather than left open: the rejection of shadow
  Roots and the deferral of wiki-links (see DECISIONS.md).
