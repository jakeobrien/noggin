# Vision — C3 compare

<!-- Per-cycle. The raw, high-level goals for THIS cycle. Not formalized, possibly un-vetted;
     captured from the human during brainstorming. May contain contradictions and open ends —
     SPEC.md is where they get resolved. Free-form. Intent that reaches beyond this cycle
     belongs in ROADMAP.md, not here. The cycle key (C<n>) comes from the roadmap ledger. -->

## Goals

More than one document on screen at once, with keyboard targeting (the roadmap's C3
one-liner). Concretely:

- A second, optional Document Viewer (DV2) beside the always-present first one (DV1), fixed
  50/50 left/right split. Fully independent: own Active Document, own Peek, own Find, own
  scroll — not a mirror or a preview of DV1.
- DV2 is created on demand by *targeting* it from Files Nav or the Picker with a modifier,
  not by a separate "open a second viewer" action.
- A single, consistent targeting scheme: `shift+` on the relevant key routes the action to
  DV2 instead of DV1 (`shift+enter` to load, `shift+p` to peek, from Files Nav; `shift+enter`
  to load from the Picker).
- Direct focus-jump keys reachable from anywhere: `r`/`f`/`d`/`shift+d` for Roots Nav/Files
  Nav/DV1/DV2. This replaces Tab/Shift+Tab pane-cycling, which becomes redundant and is
  removed.
- Close DV2 with `ctrl+x`, global (outside modals).
- `NOTES.md`'s original "Targetting" section got the mechanism right in spirit but wrong in
  two ways this cycle corrects: it proposed a silent no-op when the second viewer doesn't
  exist (we want create-on-demand instead, for the load/peek actions — a bare focus request
  is the one place the no-op survives, since there's no document to create there), and it
  used `ctrl+` where we're standardizing on `shift+` throughout.

Adjacent fix riding along: the staleness indicator (S10/C2a) is currently a Footer suffix
that, per D53, essentially never fits at normal terminal widths — invisible when it should
be, and it fights with hint text for the same line when it does fit. Moving it to its own
dedicated, highlighted row above the Footer (present only while stale) fixes this outright,
and matters more now that there's a second document's disk-state to report on too.

## Non-goals / boundaries

- Not resizable/multiple/stacked viewers — exactly two, exactly one fixed 50/50 left-right
  split.
- Not "read while editing" (tmux-pane-beside-a-live-session) — D3 flagged this cycle as the
  trigger condition for it, but it's explicitly not being picked up here.
- Not per-source (per-viewer / tree) staleness granularity — staying with one combined,
  deliberately unspecific signal (D49's existing rationale), just relocated and restyled.
  Considered and consciously deferred, not an oversight (see effort/impact note below).
- Not visual "1"/"2" labeling of the viewers — existing focus highlighting is enough.
- Not touching `m.message`'s general transient-notice mechanism (errors, confirmations) —
  it already fits "transient" cleanly (event-scoped, cleared on next keypress) and has a
  different problem shape than staleness (an ongoing condition) — left exactly as-is.

## Notes

- Effort/impact considered for the staleness relocation: a unified single row (cheap, no
  change to the collapsed-boolean model) vs. per-pane per-source indicators (more locally
  correct, but re-litigates a deliberate C2a decision and multiplies layout surface) vs. a
  hybrid (source-aware wording in one shared row). Went with the unified row — see
  DECISIONS.md for the reasoning once logged.
- Open risk, not a blocker: `shift+enter` (and equally `ctrl+enter`, which this cycle isn't
  using) depends on terminal keyboard-protocol support to disambiguate from plain `enter`.
  To be verified at build time; fallback binding TBD if it doesn't hold up. Logging as an
  open question.
