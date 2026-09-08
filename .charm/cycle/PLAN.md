# Plan — C6 sources

<!-- Per-cycle. The spec decomposed into milestones and tasks. Technically aware but
     high-level: state WHAT and the done-condition, not step-by-step HOW — trust executor
     judgment. Every task MUST have an acceptance criterion, machine-checkable where possible,
     and an effort estimate: a relative 1-5 gut-scoping (1 trivial, e.g. fixing a few typos;
     5 very high, e.g. large or complicated tasks) — not a rigorous prediction, just how the
     task looks at scoping time. Durable keys: milestones M1, M2...; tasks M1.1, M1.2...
     Never renumber; append or suffix (M1.2a) when inserting. Keys are scoped to this cycle —
     durable docs referencing them qualify with the cycle key (C3/M1.2). Note dependencies
     where ordering matters. -->

Ordering note: M1 (Source model) and M2 (Wikipedia client) gate everything Wikipedia-shaped;
M2 is independent of M1 and can be built in either order or interleaved. Milestones M6, M7 and
M9 are the spec's pre-agreed cuts in cut order (Recent → Root names → background tint) and are
sequenced late so overrun costs the least-valued work; M8's location line and Picker rows fall
back to the folder base name if M7 is cut. M8.1 (half-page paging) and M3.1 (2-char threshold)
are trivial and independent — pull them forward whenever a short slice is wanted.

Cross-cutting: every network call is user-initiated (search, Activate, Peek, Refresh — D119)
and every network-touching task must be testable without the network (fake/injected client,
`httptest`); `go test ./...` never reaches Wikimedia. Findings: `findings/wikimedia-api.md`.

## M1 — Source model and Sources Nav

Delivers: S1, S5 (the mode/scope structure — lists fill in M5/M6), store groundwork for S4

- **M1.1** — Introduce the **Source** as the model's selection entity: a Source is a Root or
  a Query Source; the Wikipedia Source is a singleton, always present, never persisted, never
  starred; "Active Source" replaces the Active Root as session state, and every Root-only
  subsystem (tree walk, grep, git markers, tree staleness, Refresh) tolerates a Query Source
  being active — done when: `go test ./...` passes and a model test activates the Wikipedia
  Source with zero Roots, renders every pane, and runs Refresh without panic or error — effort:
  4
- **M1.2** — Sources Nav lists the Wikipedia Source last as a two-line row (`Wikipedia` /
  `en.wikipedia.org`, line 2 de-emphasized like a path), present with zero or many Roots; the
  "No roots…" empty state is gone; `s` on it is a no-op with a Footer message; broken/starred
  markers never apply to it — done when: rootsnav render tests assert the row with zero and
  several Roots, the Footer message on `s`, and the empty state's absence — effort: 2 (needs:
  M1.1)
- **M1.3** — Relabel "Active Root" → "Active Source" in every scope label, Footer hint and
  Help entry where a Query Source can be the active one (Files Nav Recent/Starred scopes, File
  Picker scope); Grep's scope keeps "Active Root" — done when: `grep -rn "Active Root"
  cmd/noggin` matches only Grep-scope labels and Root-only code comments, and existing tests
  are updated rather than deleted — effort: 1 (needs: M1.1)
- **M1.4** — Files Nav with Wikipedia active: Browse Modes are `Starred` and `Recent` only,
  first activation lands on `Starred`; scopes are `Starred`: Active Source | All Starred and
  `Recent`: Active Source | All Sources (the Recent scope formerly `All Roots` is relabelled
  `All Sources` for every Active Source); per-Source browse memory covers Wikipedia — done
  when: browse tests cover the mode cycle, each scope cycle and the mode/scope row labels from
  Wikipedia-active, plus memory across a Wikipedia → Root → Wikipedia switch; lists may be
  empty — effort: 3 (needs: M1.1)
- **M1.5** — The Document Store admits **article records**: a pointer (canonical URL + title)
  plus starred and last-viewed, alongside file records, with the store's count cap and
  starred-never-evicted rule applying to both together; existing on-disk stores load unchanged
  — done when: docstore tests cover a mixed file+article round-trip, a legacy-format file
  loading byte-for-byte equivalent records, and eviction across a mixed store — effort: 2

## M2 — Wikipedia client

Delivers: the fetch/search/convert capability S2 and S3 build on (`internal/` package)

- **M2.1** — Move the spike's HTML→markdown pipeline (math, images, chrome strip, links,
  convert, cleanup) into an internal package importable by `cmd/noggin`; `cmd/wikispike`
  becomes a thin CLI over it (D120) — done when: the spike's existing tests pass against the
  moved code and regenerating one C4a sample through `wikispike` before and after the move
  yields identical markdown (recorded in the task's commit message) — effort: 2
- **M2.2** — Title search: a query yields a bounded list of `{title, description}` in
  Wikipedia's order via the title-search endpoint, with noggin's descriptive User-Agent and
  errors that distinguish offline from service failure — done when: `httptest`-backed tests
  cover results with and without descriptions, an empty result, an HTTP error, a connection
  failure and the User-Agent header — effort: 2 (needs: M2.1)
- **M2.3** — Canonical fetch: fetching a title follows redirects and reports the **canonical
  URL and title** from the final response, converts to markdown, bounds the request with a
  timeout, and caches the result for the session keyed by canonical URL with an explicit
  bypass for Refresh — done when: `httptest` tests cover a redirect resolving to the canonical
  title/URL, a 404, a timeout, a cache hit (no second request) and a bypass (second request)
  — effort: 3 (needs: M2.1)

## M3 — Wikipedia Picker

Delivers: S2, S13, S6 (Picker scope rules)

- **M3.1** — Minimum query length is 2 in every Picker mode; the debounce becomes per-mode so
  Wikipedia mode can wait longer than local modes — done when: the below-threshold status
  text and tests reflect 2, a 2-character File search runs, and a test shows Wikipedia mode's
  debounce exceeds File mode's — effort: 1
- **M3.2** — Wikipedia mode joins the mode cycle `File → Grep → Wikipedia`, reachable from any
  Active Source; opening the Picker with Wikipedia active lands on Wikipedia mode; the scope
  key is a no-op and the scope reads `-` — done when: picker tests cover the three-way cycle
  from both kinds of Active Source, the landing mode, and the `-` scope label — effort: 2
  (needs: M1.1)
- **M3.3** — Wikipedia search runs through the existing generation-counter debounce as an
  async cmd against an injectable client; Results rows show title + description (when
  present); a failed search sets the Picker status line and keeps the previous Results; the
  Preview pane shows a short placeholder — done when: model tests with a fake client cover
  results rendering, stale-generation discard, failure-keeps-rows, and the placeholder —
  effort: 4 (needs: M2.2, M3.2)
- **M3.4** — The Picker star key on a Wikipedia Result toggles the article's record in the
  Document Store and shows the star marker on the row, persisting like a file star — done
  when: a picker star test stars and un-stars an article, asserts the marker and the
  persisted record — effort: 2 (needs: M1.5, M3.3)
- **M3.5** — Cross-class Picker rules: with Wikipedia active, File and Grep modes skip
  `Active Root` and offer only `All Roots` | `Starred`; File mode's `Starred` scope includes
  starred articles matched on title and labelled `Wikipedia`; File `All Roots` and Grep in
  every scope exclude articles — done when: tests assert each of the four rules with a store
  holding both a starred file and a starred article — effort: 3 (needs: M1.5, M3.2)

## M4 — Opening an article

Delivers: S3

- **M4.1** — A `document` carries its source identity (file path, or article URL + title) so
  the viewer title row, Document History and Find work on an article; the article's history
  and store key is its canonical URL — done when: viewer/history tests open an article
  document, Find within it, and walk back/forward between it and a file — effort: 3 (needs:
  M1.5)
- **M4.2** — Async fetch-and-open: Activating a Wikipedia Result or Peeking it issues a fetch
  cmd; the UI stays responsive and shows an in-progress indicator; on completion the targeted
  viewer (Targeting modifier, Peek semantics) renders the canonical article; on failure a
  Footer message appears and the viewer's document is untouched; a completion whose target
  moved on (Peek cleared, Picker re-opened) is dropped — done when: model tests with a fake
  client cover Activate into DV1 and DV2, Peek, a failure, the indicator while in flight, and
  a stale completion — effort: 5 (needs: M2.3, M3.3, M4.1)
- **M4.3** — `R` re-fetches every open article bypassing the session cache; the staleness
  indicator never fires for an article; `e` on an article is refused with a Footer message —
  done when: refresh, staleness and edit tests each cover an article in DV1 beside a file in
  DV2 — effort: 2 (needs: M4.2)

## M5 — Articles are documents: stars and cross-Source lists

Delivers: S4 (star parts), S5 (Starred rows), S6 (Files Nav lists)

- **M5.1** — Star and un-star an article from the focused Document Viewer and from a Files
  Nav row, with the markers on the viewer title row and list rows — done when: tests toggle
  from each place and assert the marker and the store record — effort: 2 (needs: M4.1)
- **M5.2** — Files Nav Starred lists include articles: with Wikipedia active, `Active Source`
  shows starred articles by title and `All Starred` shows files and articles; with a Root
  active, `All Starred` includes articles labelled `Wikipedia` and `Orphans` excludes them;
  Activate and Peek on an article row fetch per M4.2 — done when: browse_starred tests cover
  each scope from both kinds of Active Source and an Activate from a row — effort: 3 (needs:
  M1.4, M1.5, M4.2)

## M6 — Recent for articles

Delivers: S7 and the Recent parts of S4, S5, S6 — *pre-agreed cut 1*

- **M6.1** — Activating an article records last-viewed (Peek does not); Recent with Wikipedia
  active (`Active Source` | `All Sources`) and with a Root active (`All Sources`, labelled
  `Wikipedia`) lists articles by recency under the same per-Source cap as files, omitting
  entries the store can no longer resolve — done when: recency/browse tests cover recording on
  Activate, non-recording on Peek, ordering, the cap and the label — effort: 3 (needs: M4.2,
  M1.4)

## M7 — Root names

Delivers: S8 — *pre-agreed cut 2*

- **M7.1** — A persisted Root carries a name, defaulting to its folder's base name, persisted
  in the roots store across sessions; existing stores load with default names — done when:
  store round-trip tests cover named and unnamed Roots and a legacy-format store — effort: 2
- **M7.2** — Inline rename in the Sources Nav: a rename key edits the selected Root's name on
  its row, Enter confirms, Esc cancels, empty reverts to the default; the Ad-hoc Root is
  refused with a Footer message until starred, and starring gives it the default name; the
  Wikipedia Source is refused — done when: rootsnav tests cover each path and the key appears
  in Footer hints and the Help Modal — effort: 3 (needs: M7.1)
- **M7.3** — One Root-label helper feeds every place a Root's base name shows today: Sources
  Nav line 1, Picker result labels, cross-Source Files Nav rows — done when: tests assert a
  renamed Root's label in all three, and no `filepath.Base(<root path>)` label computation
  remains outside the helper — effort: 2 (needs: M7.1)

## M8 — Viewer and Picker polish

Delivers: S9, S10, S12

- **M8.1** — `pgup`/`pgdn` scroll half a page in a Document Viewer (same distance as
  `ctrl+u`/`ctrl+d`); other panes unchanged — done when: a viewer scroll test shows `pgdn`
  and `ctrl+d` land on the same offset, the Help Modal lists them together, and a Files Nav
  paging test is unchanged — effort: 1
- **M8.2** — Location line: each Document Viewer's bottom row shows, de-emphasized, `<Root
  name> · <path relative to the Root>` for a file under a Root, the full path for an orphan,
  the URL for an article; reflects a Peeked document while the Peek lasts; absent with no
  document; content loses one row — done when: render tests cover the four states plus Peek,
  and the content height assertion drops by one — effort: 3 (needs: M4.1, M7.3)
- **M8.3** — De-emphasized secondary text: the Sources Nav path line and Picker File-row paths
  render dim with matched path characters emphasized against the dim, the Root label at normal
  weight; the Root's parent-path fragment is dropped from Picker File rows (D117) — done when:
  render tests assert the styles, the fragment's absence, and C1/D22's disambiguation tests are
  retired with a note — effort: 2 (needs: M7.3)

## M9 — Focus by background

Delivers: S11 — *pre-agreed cut 3*

- **M9.1** — Unfocused panes render on a subtly lighter background tint; the focused pane
  keeps the terminal background; the border-colour cue stays — done when: frame tests assert
  the tint on every unfocused pane and its absence on the focused one across all four focus
  targets, and a `make check-pty` check confirms the tint reaches terminal cells — effort: 3

## M10 — Product documentation

Delivers: README coverage of every user-visible change in S1–S13

- **M10.1** — Update `README.md` for Sources (Sources Nav, Wikipedia Source, Wikipedia Picker,
  articles in Starred/Recent, Root names, location line, half-page paging, 2-char threshold,
  background focus), with every key and label verified against the implementation — done
  when: `grep -n "Roots Nav\|No roots\|at least 3" README.md` is empty and each new binding in
  the README matches `keymap.go` — effort: 2 (needs: M1–M9 as built)
