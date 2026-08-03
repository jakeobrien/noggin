# Spec — C2 find

<!-- Per-cycle. Functional, non-technical, concrete. This is the contract of intent: what the
     deliverables ARE, not how they are built. Keep it product-level so upfront errors don't
     cascade into implementation. Durable keys (S1, S2...) — never renumber. Mid-cycle
     changes go through /brainstorm amend: supersede or suffix keys, renew the agreed date;
     charm-replan then reconciles the plan. Keys are scoped to this cycle — durable docs
     referencing them qualify with the cycle key (C3/S2). -->

Status: agreed 2026-08-02

<!-- Pre-agreed cut: if the cycle runs long, S5 (match highlighting) is the first thing to
     drop. See D32. -->

## Deliverables

- **S1 — The Picker.** A modal fuzzy-finding UI, openable from any Focus (including while the
  Navigator is hidden), composed of a Prompt, a Results list and a Preview pane. It always
  opens the same way: File Picker mode, Active Root scope, empty Prompt. It shows which mode
  and which Scope it is in. Dismissing it changes nothing — same Active Document, same Active
  Root, same scroll positions, whatever was previewed.

- **S2 — Always-typing interaction.** Every printable keystroke goes to the Prompt; there is no
  insert/normal mode split. Moving the Result selection, scrolling the Preview, switching mode
  and switching Scope all happen on keys that do not type. The Footer shows the Picker's
  actions, and the Help Modal covers them.

- **S3 — File Picker.** Fuzzy-matches the query against the **file name only** (not the path)
  across the Readable files in Scope. Results are ordered by match strength. Each row shows the
  file's Root and its path relative to that Root, so same-named files are distinguishable.

- **S4 — Grep Picker.** Searches the **literal** query text — no regex — across the Readable
  files in Scope, with smart case (case-insensitive unless the query contains uppercase). One
  Result row per matching line, showing the file, the line number and the line itself.

- **S5 — Match highlighting.** File Picker rows highlight the matched characters within the
  file name. Grep Picker rows highlight the matched text within the line.

- **S6 — Scope.** Every Picker searches either the Active Root or All Roots. All Roots silently
  skips Broken Roots — no error row.

- **S7 — Switch mode and Scope in place.** Mode (File ↔ Grep) and Scope (Active Root ↔ All
  Roots) are switched inside the open Picker; that is the only place either is chosen. The
  typed Prompt carries over across both kinds of switch — nothing is cleared. There is no
  intermediate menu modal for choosing a search.

- **S8 — No memory.** The Picker retains nothing: not the mode, not the Scope, not the query,
  neither across opens within a session nor across sessions.

- **S9 — Preview pane.** Shows the selected Result's file as **raw, unrendered text**. The File
  Picker previews from the top of the file; the Grep Picker opens scrolled to the matched line
  with that line highlighted. The user can scroll the Preview. The Preview pane never changes
  the Active Document.

- **S10 — Activate.** Activating a Result makes its file the Active Document. It does not change
  the Active Root, and does not move, expand or re-select the Files Nav tree — the Active
  Document may belong to a Root that is not active, which C1 already permits by design.
  Activating a Grep Result lands the Document Viewer on the matched line.

- **S11 — Grep hands off to Find.** After a Grep Result is activated, the document's Find state
  holds the grep query: the match is indicated as a Committed Find, and next/previous
  traversal moves between that query's other matches in the document.

- **S12 — Query threshold, result cap, empty states.** Neither Picker searches until at least
  3 characters are typed; below that the Picker says so. Results are capped, and when the cap
  is reached the Picker says the list is truncated rather than implying completeness. A query
  that matches nothing shows "no results".

- **S13 — The allowlist holds.** Both Pickers see only C1's Readable files (`.md`, `.markdown`,
  `.txt`, `README`, `LICENSE`, `CHANGELOG`). Grep reads no other file. D9 (general text/source
  viewing) is not answered by this cycle.

- **S14 — ripgrep is a requirement.** `ripgrep` supplies both the file list and the text search,
  so its ignore rules (`.gitignore`, `.ignore`, hidden-file defaults) govern what the Pickers
  see. When `ripgrep` is absent the session still launches fully usable for reading, the
  Pickers are unavailable, and reaching for one says why. The requirement is documented for
  users.

## Out of scope

Deliberately later (each recorded on the roadmap's Unscheduled list):

- Regex or fuzzy grep queries; C2 is literal only.
- A rendered-markdown Preview pane instead of raw text.
- Revealing the activated file in the Files Nav tree (expanding ancestors, selecting it).
- Matching the File Picker query against path components, not just the file name.
- Any Picker memory: last mode, last Scope, last query, repeat-last-search.

Out for other reasons:

- **Replacing C1's `/` Find.** Find stays exactly as it is. `NOTES.md`'s "Document Search" (a
  Picker over the Active Document) is considered already served by Find and is dropped, not
  deferred.
- **A Search Picker menu** — a modal for choosing which search to open. Eliminated by S7, not
  deferred.
- **Sending Results to the Files Nav** ("browse-by-results"), **Scope = documents opened this
  session**, and **hidden/ignored visibility toggles** — all already on Unscheduled from C1.
- **Answering D9.** Widening the allowlist is its own decision and its own cycle's work.
- **Tree/Picker ignore-rule consistency.** noggin's own tree walk and ripgrep will disagree at
  the edges; a file may be visible in one and not the other. Accepted for this cycle.

## Acceptance (cycle-level)

- From a cold launch with several Roots, a document can be found by name and opened without
  touching the Navigator: open the Picker, type 3+ characters of a file name, see it ranked
  with the matched characters highlighted, preview it, activate it, read it.
- The same for text: switch the open Picker to Grep, keep the typed query, get one row per
  matching line across the Scope, activate one, and land on that line with the match
  highlighted and next/previous traversal working from it.
- Scope and mode can both be changed inside one open Picker without retyping the query.
- Dismissing a Picker demonstrably leaves the session exactly as it was found.
- A session with no `ripgrep` installed still launches and reads documents; the Pickers report
  their unavailability instead of failing silently or crashing.
- No file outside the Readable allowlist is ever surfaced by either Picker.
- Typing in either Picker stays responsive on a realistically large Root — the Prompt never
  feels blocked.

## Open questions

None blocking. Carried open in `DECISIONS.md`: D9 (does noggin ever become a general
text-file viewer — deliberately not answered here, see S13) and D33 (does the raw Preview
pane ever become rendered markdown, and how would grep's line addressing survive it).
