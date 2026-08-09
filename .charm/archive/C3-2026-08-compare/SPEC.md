# Spec — C3 compare

<!-- Per-cycle. Functional, non-technical, concrete. This is the contract of intent: what the
     deliverables ARE, not how they are built. Keep it product-level so upfront errors don't
     cascade into implementation. Durable keys (S1, S2...) — never renumber. Mid-cycle
     changes go through /brainstorm amend: supersede or suffix keys, renew the agreed date;
     charm-replan then reconciles the plan. Keys are scoped to this cycle — durable docs
     referencing them qualify with the cycle key (C3/S2). -->

Status: agreed 2026-08-08

## Deliverables

- **S1** — A second Document Viewer (DV2) is available beside the always-present first
  (DV1), in a fixed 50/50 left-right split of the Documents region. DV1 alone takes the full
  width whenever DV2 is absent. DV2 does not exist at launch; it is created on demand (S3),
  never by a separate explicit "open a second viewer" action.
- **S2** — DV1 and DV2 are fully independent Document Viewers: each has its own Active
  Document, Peek state, Find/Committed-Find state and scroll position. An action taken in or
  targeted at one never affects the other's state.
- **S3** — A `ctrl+` targeting modifier, applied consistently: `ctrl+enter` and `ctrl+p`
  from the Files Nav, and `ctrl+enter` from the Picker, route that load/peek to DV2 instead
  of DV1 — creating DV2 first if it does not already exist. The unmodified keys always target
  DV1. Loading into DV2 moves Focus to DV2 (mirroring DV1's existing Activate behavior);
  peeking into DV2 never moves Focus (mirroring DV1's existing Peek behavior exactly).
- **S4** — DV2 can be closed with `ctrl+x`, reachable from any pane's Focus (Roots Nav, Files
  Nav, DV1, DV2) but not while a modal (Picker, Help Modal) owns the keyboard. No-op if DV2
  does not exist. If DV2 held Focus, Focus moves to DV1 afterward; from any other Focus,
  Focus is unchanged.
- **S5** — Four direct focus-jump keys, reachable from any non-modal Focus: `r` (Roots Nav),
  `f` (Files Nav), `o` (DV1), `ctrl+o` (DV2 — no-op if DV2 does not exist, since a bare focus
  request has no document to create). The `ctrl+` prefix mirrors S3's targeting modifier for
  consistency, even though the two are different mechanisms — jumping Focus has no load/peek
  side effect on either viewer's content. Tab/Shift+Tab pane-cycling is removed as redundant
  once these four cover every pane.
- **S6** — Once a Document Viewer (DV1 or DV2) holds Focus, its own local commands (Find,
  Edit, scrolling) act on that viewer's Active Document only, never the other viewer's.
- **S7** — Refresh (`R`) re-reads the Active Root's tree and both DV1's and DV2's Active
  Documents (whichever exist) from disk in one action, regardless of which pane holds Focus.
- **S8** — DV2's existence, its Active Document, and its Peek/Find state are session-only:
  always absent at launch, never persisted, never restored.
- **S9** — The staleness indicator (S10/C2a) moves off the Footer's hint line into its own
  dedicated row, styled with a highlight color and inverse text, shown only while something
  is stale (occupies no space otherwise). It remains one combined, deliberately unspecific
  signal — now spanning the Active Root's tree, DV1's Active Document and DV2's Active
  Document (previously just tree and the one document) — with the same wording and the same
  "press R to refresh" resolution as today.

## Out of scope

- Read-while-editing / concurrent tmux-pane editing (Unscheduled, from C1/D3) — this cycle
  makes it meaningful but does not deliver it.
- More than two Document Viewers; a resizable or ratio-adjustable split; stacked or
  horizontal split orientation — exactly two, exactly 50/50 left-right, fixed.
- Visual "1"/"2" labeling of the viewers beyond the existing focus-highlighting convention.
- Wiki-link/URL jump targeting (deferred with wiki-links themselves to C4).
- "Follow" mode and any targeting for it — Follow itself remains Unscheduled (from C1).
- Per-source (per-viewer / per-tree) staleness signal granularity — one combined signal only
  (see S9). Considered and consciously deferred; logged to Unscheduled, not dropped.
- Changing `m.message`'s general transient-notice mechanism (errors, confirmations) — it
  already fits "transient" cleanly and is left exactly as it is today.

## Acceptance (cycle-level)

- From a fresh single-viewer session, a user can load or peek a second document into a
  fully independent DV2, side by side with DV1, using the `ctrl+` targeting modifier from
  either the Files Nav or the Picker.
- Once focused, each viewer's Find/Edit/scroll act only on that viewer's own Active Document.
- DV2 closes with `ctrl+x` from any non-modal Focus; Focus lands on DV1 if DV2 was the one
  closed.
- Every pane (Roots Nav, Files Nav, DV1, DV2) is reachable by a direct key (`r`/`f`/`o`/
  `ctrl+o`) from anywhere non-modal; Tab no longer cycles panes.
- Refresh updates the tree and both open documents' content in one action.
- The staleness signal is visibly legible (styled, never truncated away by hint text) at
  ordinary terminal widths, and reflects all three sources (tree, DV1, DV2).

## Open questions

<!-- Unresolved items blocking or shaping the spec. Move to DECISIONS.md once material. -->
<!-- none currently — the shift+enter/ctrl+enter disambiguation question was resolved into
     DECISIONS.md (D53, superseded by D55). -->
