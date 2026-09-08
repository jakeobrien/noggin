# Spec — C6 sources

<!-- Per-cycle. Functional, non-technical, concrete. This is the contract of intent: what the
     deliverables ARE, not how they are built. Keep it product-level so upfront errors don't
     cascade into implementation. Durable keys (S1, S2...) — never renumber. Mid-cycle
     changes go through /brainstorm amend: supersede or suffix keys, renew the agreed date;
     charm-replan then reconciles the plan. Keys are scoped to this cycle — durable docs
     referencing them qualify with the cycle key (C3/S2). -->

Status: agreed 2026-09-07

## Framing

Wikipedia enters noggin not as a new mode but by widening what sits in the upper Navigator.
A **Source** is anything noggin can serve documents from. Two classes exist after this cycle:

- **Root** — a folder. Plural, user-created, browsable as a tree, greppable. Unchanged in
  kind; gains an editable name.
- **Query Source** — a remote service reached only by search. Exactly one of each kind,
  always present, cannot be created, removed or starred. **Wikipedia** is the first; the
  dictionary/thesaurus (roadmap) is the intended second.

Both classes share the deeper rule: noggin never stores a Source's content, only pointers to
it (paths for Roots; canonical URL + title for Wikipedia) plus star and last-viewed state.
The difference is reachability — a Root is bounded and crawlable, so it has a Tree and a
grep; a Query Source is not, so its documents are reached only by search, and afterwards via
Starred or Recent.

Every deliverable below that mentions Wikipedia must hold, reworded only by name, for a
second Query Source. Where a behaviour would need restating for the dictionary, the model is
wrong, not the dictionary.

## Deliverables

- **S1** — **Sources Nav.** The upper Navigator pane lists every Source: the Ad-hoc Root
  first (if any), then persisted Roots, then Query Sources — the **Wikipedia Source** as a
  two-line row (`Wikipedia` / `en.wikipedia.org`, the second line de-emphasized like a Root's
  path). Wikipedia is always listed, even with zero Roots (the "No roots…" empty state no
  longer occurs). It can be selected and made the **Active Source**, but `s` on it is a no-op
  with a Footer message, and nothing can add, remove or hide it. Everything that today says
  "Active Root" reads as "Active Source" wherever a Query Source can be the one active.

- **S2** — **Wikipedia search is a Picker mode.** The Picker's mode key cycles
  `File → Grep → Wikipedia`; the mode is reachable from any Active Source. Opening the Picker
  while Wikipedia is the Active Source lands on Wikipedia mode. Wikipedia mode has no Scope
  (the scope key is a no-op and reads `-`). The Prompt matches **article titles** (Wikipedia's
  own title/prefix search; not full-text), and shows a bounded Results list in Wikipedia's
  order — each row the article title plus, when Wikipedia supplies one, its short description,
  with a star marker if the article is Starred. There is **no Preview**: the Preview pane shows
  a short placeholder instead of article content. A search that fails (offline, service error)
  says so in the Picker's status line and leaves the last Results in place. The Picker's star
  key works on a Wikipedia Result exactly as on a file Result.

- **S3** — **Opening an article.** Activating a Wikipedia Result, a Starred or Recent article
  row, or Peeking one, fetches the article, converts it to markdown at the quality C4a certified
  (D112, including its known maths limitation), and renders it in the targeted Document Viewer
  — the Targeting modifier, Peek semantics, Document History and Find all apply as for a file.
  Redirects are followed: the document noggin opens (and records) is the **canonical article**,
  English Wikipedia only. While a fetch is in flight the **UI stays responsive** and shows a
  visible in-progress indicator; a failed fetch (offline, article gone, timeout) surfaces as a
  Footer message and leaves the viewer's current document untouched. Fetched articles are
  cached **for the session only**; `R` (Refresh) re-fetches any open Wikipedia article; the
  staleness indicator never fires for one. `e` on a Wikipedia article is refused with a Footer
  message. noggin identifies itself to Wikimedia per their API etiquette (descriptive
  User-Agent).

- **S4** — **Articles are documents.** A Wikipedia article can be Starred and un-starred from
  every place a file can (Files Nav row, focused Document Viewer, Picker), with the same
  visible markers (viewer title row, list rows, Picker rows). Its recency is recorded on
  Activate exactly as a file's is (Peek does not count). What is persisted per article is a
  pointer — canonical URL and title — plus starred and last-viewed; never content. The
  existing store bounds (count cap, starred never evicted) cover articles and files together.

- **S5** — **Files Nav with Wikipedia active.** Browse Modes are `Starred` and `Recent` only
  (no `Tree`); the mode key cycles between the two, and the first activation of Wikipedia
  lands on `Starred`. Scopes: `Starred` — `Active Source` | `All Starred` (no `Orphans`; that is
  a Root concept and stays reachable from any Root); `Recent` — `Active Source` | `All Sources`.
  Rows show the article title. Peek works on article rows (S3 rules). Per-Source browse memory
  applies to Wikipedia as to any Root.

- **S6** — **Cross-Source lists and searches.** With a Root active, Wikipedia articles appear
  in the Files Nav's `All Starred` and `All Sources` (Recent; the scope formerly labelled
  `All Roots`) lists, labelled `Wikipedia` where files are labelled by Root name. In the Picker:
  File mode's `Starred` scope includes starred articles (matched on title, labelled
  `Wikipedia`) — this is the **only** File search that spans both classes; File mode's
  `All Roots` scope is files only. **Grep never includes Wikipedia**, in any scope. With
  Wikipedia active, File and Grep modes skip the `Active Root` scope and offer only
  `All Roots` | `Starred` (no empty dead-end scope).

- **S7** — **Recent for articles is Recent for files.** Same persistence, same per-Source cap,
  same recency-only ordering, same omission of entries that can no longer be resolved. No
  Wikipedia-specific time cap or session-scoping.

- **S8** — **Root names.** A persisted Root has a **name**, defaulting to its folder's base
  name. A rename key in the Sources Nav edits the selected Root's name inline on its row:
  Enter confirms, Esc cancels, an empty name reverts to the default. The name persists across
  sessions. The Ad-hoc Root cannot be renamed until it is starred (Footer message); starring
  gives it the default name. The name is used everywhere the Root's base name is shown today:
  Sources Nav line 1, Picker result labels, Files Nav rows in cross-Source scopes, and the S9
  location line. Uniqueness is not enforced — the Sources Nav path line disambiguates.

- **S9** — **Document Viewer location line.** The bottom row of each Document Viewer shows,
  de-emphasized, where its document lives: `<Root name> · <path relative to the Root>` for a
  file under a Root; the full path for an orphan; the article's **URL** for a Wikipedia
  article. It reflects a Peeked document while the Peek lasts and is absent when the viewer
  has no document. Costs the content one row.

- **S10** — **De-emphasized secondary text.** The Sources Nav's path line (line 2) and, in
  Picker File results, the file's path are rendered de-emphasized (reduced contrast), with the
  matched characters of the path emphasized against it; the `<Root name>` label stays at
  normal weight to anchor the row. The Root's parent-path fragment is
  **dropped** from Picker File rows: a row is `<Root name>` + path relative to the Root. The
  same-named-Root disambiguation this fragment provided (C1/D22) is accepted as lost; S8's
  names cover it.

- **S11** — **Focus by background.** Unfocused panes render on a subtly lighter background
  tint; the focused pane keeps the terminal's own background (chosen over darkening the
  focused pane, which is a no-op on already-black terminals). The existing border-colour cue
  stays. This is an evaluation of background-based focus ahead of the roadmap's style pass, not
  a final palette.

- **S12** — **Half-page paging.** In a Document Viewer, `pgup`/`pgdn` scroll half a page (the
  same distance as `ctrl+u`/`ctrl+d`). Other panes are unchanged.

- **S13** — **Picker query threshold.** The minimum query length drops from 3 to 2 characters
  in every Picker mode (revises C2/D31). Searching stays debounced so no search runs per
  keystroke; Wikipedia mode waits somewhat longer than local modes since each search is a
  network call.

## Out of scope

- The dictionary / thesaurus Query Source — shapes S1–S7, is not built (roadmap: Unscheduled).
- Following links inside articles (links render as plain text; roadmap: Unscheduled).
- Hiding or disabling the Wikipedia Source; non-English Wikipedias.
- Persisting article content, offline reading of starred articles, refreshing stars in bulk.
- Grep over Wikipedia articles in any form; Wikipedia in the File Picker's `All Roots` scope.
- Launching noggin with a Wikipedia URL or title as its path argument.
- The `$EDITOR` handoff for articles.
- The style pass — dark/light mode, palette, style sheet (roadmap: Unscheduled). S10/S11 are
  provisional choices for evaluation within today's 256-colour styling.
- Half-page `pgup`/`pgdn` outside the Document Viewer; names for Ad-hoc Roots.
- Rendered-markdown Preview for any Result (D33 stays open).
- Everything the withdrawn `C5 organize` held: tags, outline, file management.

## Acceptance (cycle-level)

- From `noggin .`, with no Roots starred, Wikipedia is listed in the Sources Nav; typing two
  characters in the Picker's Wikipedia mode yields article titles; Activating one renders the
  article in DV1 while the UI stays responsive; `ctrl+enter` on another puts it in DV2 beside
  a local markdown file — side by side, one keystroke each.
- Star an article, quit, relaunch: it is in Wikipedia's `Starred` list, in a Root's
  `All Starred` list labelled `Wikipedia`, and in the File Picker's `Starred` scope; Activating
  it from any of those re-fetches and renders it; un-starring from any of those removes the
  star everywhere.
- An unstarred article viewed earlier is reachable from `Recent` after relaunch.
- With the network unavailable: Starred and Recent article lists still display; a search or
  fetch reports failure in the status line / Footer; the open document is unaffected.
- A Grep in any scope, and a File search in `All Roots`, return no Wikipedia results; a File
  search in `Starred` matches a starred article by title.
- Renaming a Root changes its label in the Sources Nav, Picker rows, cross-Source Files Nav
  rows and the location line, and survives relaunch; the Picker row no longer shows the Root's
  parent path.
- The location line, path de-emphasis, background focus tint, half-page paging and 2-character
  threshold are each observable as described.
- Reading S1–S7 with "dictionary" substituted for "Wikipedia" requires no rewording beyond the
  name — reviewed at spec agreement, re-checked at validate.

## Pre-agreed cuts

If the cycle overruns, cut in this order, each becoming an Unscheduled roadmap line:

1. Wikipedia in `Recent` (S7 and the Recent parts of S4/S5/S6): articles reachable only via
   search and Starred. (A Picker Preview for articles was the agreed first cut; S2 already
   omits it.)
2. Root names (S8; S9/S10 then fall back to the folder base name).
3. Background focus tint (S11).

## Open questions

- None at agreement. OQ1–OQ3 (background direction, Root label weight, result-row description)
  were resolved into S11, S10 and S2 respectively on 2026-09-07.
