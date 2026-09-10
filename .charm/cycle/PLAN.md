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
are trivial and independent — pull them forward whenever a short slice is wanted. M1.9–M1.12,
M2.9–M2.11, M3.6–M3.7, M4.4, M5.3, M7.4, M8.2a, M8.4, M8.5, M9.2 and M9.3 are replan-added
hygiene/test-gap/bugfix tasks, independent of the milestone sequencing above; pull forward
whenever a short slot is available. M1.9 and M1.11 (the `make check-pty` string-drift fixes)
are worth doing before M9.1 specifically, since M9.1's own acceptance criterion leans on that
gate. M1.11 is currently the true cursor (in-progress, not done — see PROGRESS.md): its
second clause missed one occurrence (checks.py:164, replan 2026-09-09 pass, TRIAGE T30),
reopening it ahead of M10.1.

Cross-cutting: every network call is user-initiated (search, Activate, Peek, Refresh — D119)
and every network-touching task must be testable without the network (fake/injected client,
`httptest`); `go test ./...` never reaches Wikimedia. Findings: `findings/wikimedia-api.md`.
M1.8 and M2.8 are trivial code-hygiene cleanups independent of M3; pull forward whenever a
short slot is available.

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
- **M1.6** — Propagate GLOSSARY's C6 rename: every "Roots Nav" string and code comment in
  `cmd/noggin` (pane border/footer title, comments across rootsnav.go, files.go, keymap.go,
  model.go, footer.go, layout.go, help.go, roots.go, browse_list.go, style.go,
  picker_view.go) becomes "Sources Nav"/"Sources" — done when: `grep -rn "Roots Nav"
  cmd/noggin` is empty and a rootsnav render test asserts the pane label reads "Sources" —
  effort: 1 (replan: TRIAGE T1, T2)
- **M1.7** — Guard the Document Store's article `<title>` field against embedded tab/
  newline characters corrupting a saved record (reject or escape on write) — done when: a
  docstore test round-trips an article whose title contains an embedded tab and one with an
  embedded newline without corrupting the record or adjacent fields — effort: 1 (replan:
  TRIAGE T3)
- **M1.8** — Extend M1.7's tab/newline escaping to the Document Store's path-like `<key>`
  fields (pre-existing corruption risk, not introduced by M1.7, low real-world likelihood) —
  done when: a docstore test round-trips a record whose key contains an embedded tab and one
  with an embedded newline without corrupting the record or adjacent fields, using the same
  escape mechanism as the title field, and existing on-disk stores load unchanged — effort: 1
  (replan: TRIAGE T17)
- **M1.9** — `test/pty/checks.py` (a cross-cycle regression suite, not per-cycle-scoped) still
  hard-codes pre-C6 strings that M1.3 and M1.6 deliberately renamed: the "Roots" pane label/
  footer focus name (now "Sources") and the File Picker's "Active Root" scope label (now
  "Active Source" — Grep keeps "Active Root"), across 11 assertions in `M1.2`, `M3.2`, `M3.3`,
  `M3.4`, `M5.2`, `M7.1` and `C2a/M2.1`/`C2a/M6.2` blocks (confirmed by rerunning `make
  check-pty`, not sandbox flakiness — see the replan correction of TRIAGE T18 in DECISIONS.md)
  — done when: `make check-pty` shows zero failures attributable to these two renames (any
  unrelated new failure is a separate TRIAGE item, not this task's problem) — effort: 2
  (replan: TRIAGE T18, corrected)
- **M1.10** — M1.7's and M1.8's tab/newline round-trip tests never cover a literal backslash
  in a title or key — the one character `escapeDocField`'s scheme special-cases and reverses
  first (per its doc comment) — done when: a docstore test adds a backslash case for both the
  title and key round-trips, locking in escape-order correctness — effort: 1 (replan: TRIAGE
  T2)
- **M1.11** — `test/pty/checks.py` residual string-drift M1.9's own pass missed: line 937's
  `"Roots" not in hidden_lines` assertion is vacuous since the M1.6 rename (the literal string
  never appears regardless of Navigator visibility, so the check passes unconditionally); four
  check() description strings (lines 398, 490, 551, 1428) still read "...Roots" while their own
  assertions already check "Sources" — done when: the line-937 assertion reads `"Sources" not
  in hidden_lines` and is confirmed to actually fail on a temporary Navigator-hide-toggle
  revert (then restored), and no description string in checks.py says "Roots" while its own
  assertion checks "Sources" — effort: 1 (replan: TRIAGE T1, T3)
- **M1.12** — `test/pty/checks.py:784`'s check() description string still reads "...Roots
  differs from Files and Document" though M1.6 renamed Roots Nav to Sources Nav and the
  check's own assertion already tests the post-rename footer text; M1.9's and M1.11's rename
  sweeps didn't touch this line (the assertion isn't vacuous, unlike the cases M1.11 fixed, so
  it survived both passes) — done when: the description reads "...Sources differs from Files
  and Document" and `make check-pty` still passes — effort: 1 (replan: TRIAGE T33)

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
- **M2.4** — `Client` cannot be used mid-panic from a zero-value construction (nil
  `httpClient` call, nil-map cache write), and its "safe for concurrent use" doc claim is
  backed by a test — done when: a zero-value `Client{}`'s `Fetch` no longer panics (lazy-init
  or an equivalent guard) and a `go test -race` test drives concurrent `Fetch`/`fetchAt` calls
  — effort: 2 (replan: TRIAGE T4, T5)
- **M2.5** — `internal/wikipedia` code hygiene: rename `fetchAt`/`fetchCanonicalHTML`/
  `canonicalFromURL`'s `fetchBase`/`wikiBase` parameters so they no longer shadow the
  package-level constants of the same name; document (or guard against) `canonicalFromURL`'s
  assumption that a trailing `html` URL segment means a redirect; broaden `ServiceError`/
  `OfflineError`'s doc comments to acknowledge both `search.go` and `fetch.go` use them — done
  when: no parameter shadows a same-named package constant in those three functions, and both
  error types' doc comments mention search and fetch — effort: 2 (replan: TRIAGE T6, T7, T8)
- **M2.6** — `internal/wikipedia/live_test.go`'s temporary exact-title-fetch duplication
  (called out as provisional in the M2.1 handoff, pending M2.3) now calls `Client.Fetch` (or
  `fetchAt`) instead of its own copy, and its stale "pending M2.3" comment is removed — done
  when: the opt-in live test suite still passes (`WIKISPIKE_LIVE=1`) with the duplicated fetch
  logic gone — effort: 1 (replan: TRIAGE T10)
- **M2.7** — Two low-risk `internal/wikipedia` nits: `search.go`'s `fmt.Sprintf("%d", limit)`
  becomes `strconv.Itoa(limit)`; the descriptive-User-Agent doc comments in `search.go` and
  `search_test.go` that cite "S8" are re-pointed at "S3" (S8 is Root names, unrelated) — done
  when: both changes land and existing tests still pass — effort: 1 (replan: TRIAGE T11, T12)
- **M2.8** — `Client.NewClient` builds its zero-value fields (`&http.Client{Timeout:
  fetchTimeout}`, `make(map[string]Article)`) directly instead of going through
  `ensureInit`, duplicating M2.4's lazy-init path where the two can drift — done when:
  `NewClient` constructs via `ensureInit` (or an equivalent single init path) and M2.4's
  zero-value-construction and concurrency tests still pass — effort: 1 (replan: TRIAGE T15)
- **M2.9** — Neither of `internal/wikipedia`'s two real exported entry points is ever called
  directly by a test: `NewClient()` itself (only its zero-value bypass and `ensureInit` are
  exercised) and `(*Client).Search` (only the package-level `Search` function and
  `cmd/noggin`'s fake are) — a future regression in either's one-line delegation would go
  uncaught — done when: an `httptest`-backed test exercises `NewClient()`'s `Fetch` and a
  separate test exercises `(*Client).Search` against an `httptest` server via the existing
  overridable-base pattern — effort: 1 (replan: TRIAGE T1, T22)
- **M2.10** — M2.9's two new tests mutate package-level vars (`fetchBase`, `wikiBase`,
  `searchBase`) with a save/restore-on-defer pattern but no not-parallel-safe doc warning, the
  convention the codebase already establishes for this exact hazard class
  (`cmd/noggin/files_search.go`'s `grepArgMaxBytes` note) — done when: both tests (or their var
  declarations) carry an equivalent doc comment — effort: 1 (replan: TRIAGE T2)
- **M2.11** — M2.10's not-parallel-safe mutation-contract doc comment is duplicated
  near-verbatim across `fetch_test.go` and `search_test.go` (each naming the other test by
  name) instead of stated once and cross-referenced — done when: one copy is marked canonical
  and the other points to it (or both point to a single shared note), with no wording drift
  between them — effort: 1 (replan: TRIAGE T31)

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
- **M3.6** — GLOSSARY's `Result` entry predates the Wikipedia Result (M3.2/M3.3) and the
  Picker star it can now take (M3.4): fold in a third case, an article title+description
  (Wikipedia Picker) — done when: the entry names all three Result shapes — effort: 1 (replan:
  TRIAGE T20)
- **M3.7** — `toggleHiddenFiles` (alt+h) unconditionally requeries even in `wikipediaMode`,
  where `hideDotfiles` has no bearing on Wikipedia Results — pressing alt+h there fires a
  real, debounced network search with no visible effect — done when: alt+h in `wikipediaMode`
  still flips `m.hideDotfiles` (for later File/Grep use) but does not requery, pinned by a
  test — effort: 1 (replan: TRIAGE T21)

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
- **M4.4** — M4.3's staleness and edit tests only cover "article in DV1, file in DV2"; unlike
  refresh (which already has `TestRefreshUpdatesAFileInDV1BesideAnArticleInDV2`), no mirror
  exists for either. Both are pane-symmetric by construction (staleness loops over whichever
  docs exist; `startEdit` acts on the focused viewer regardless of DV1/DV2) but unverified —
  done when: a DV2-article/DV1-file mirror test exists for both staleness and edit, matching
  the refresh precedent — effort: 1 (replan: TRIAGE T23)

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
- **M5.3** — `browseRow.rootLabel`'s doc comment predates the article case M5.2 introduced
  (rootLabel holds the fixed `wikipediaLabel` for an article row, not a Root's base name) and
  doesn't cross-reference it — done when: the comment names all three cases (Root-covered
  file, orphan file, article) — effort: 1 (replan: TRIAGE T4)

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
- **M7.4** — Two stale doc comments in `store.go`: `storedRoot` (store.go:82) still frames
  `renameStore` as a "future (M7.2)" writer though M7.2 has landed in this same cycle;
  `escapeStoreField`/`unescapeStoreField` (store.go:231) cite D61 (document-memory-is-global,
  unrelated) to justify the two stores' separate escaping code where D21/D67 (the actual
  format decisions) are the relevant citations — done when: both comments state the
  writers/citations plainly with no stale tense or wrong D-number — effort: 1 (replan: TRIAGE
  T24, T25)

## M8 — Viewer and Picker polish

Delivers: S9, S10, S12

- **M8.1** — `pgup`/`pgdn` scroll half a page in a Document Viewer (same distance as
  `ctrl+u`/`ctrl+d`); other panes unchanged — done when: a viewer scroll test shows `pgdn`
  and `ctrl+d` land on the same offset, the Help Modal lists them together, and a Files Nav
  paging test is unchanged — effort: 1
- **M8.2** — Location line: each Document Viewer's bottom row shows, de-emphasized, `<Root
  name> · <path relative to the Root>` for a file under a Root, the full path for an orphan,
  `CC BY-SA 4.0 · <URL> · Reformatted to markdown` for an article (S9a); reflects a Peeked
  document while the Peek lasts; absent with no document; content loses one row — done when:
  render tests cover the four states plus Peek, the article line's exact three-segment text,
  and the content height assertion drops by one — effort: 3 (needs: M4.1, M7.3)
- **M8.2a** — `docContentHeight`/`peekContentHeight` (location.go:79-93) floor at
  `max(boxHeight-1, 0)` instead of suppressing the location line, so at a pathologically short
  terminal (`boxHeight` 0) `composeViewerContent` still appends the location row, overflowing
  the pane's height budget and corrupting the frame (DV1's bottom border disappears, panes
  misalign) — follow `computeLayout`'s own `staleRow` precedent (never overflow) by gating
  location-line visibility on `boxHeight >= 1`; while in there, fold the two near-identical
  functions into one shared helper — done when: a `sized(t, m, 80, 4)`-then-`openDocument`
  regression test confirms no overflow and no misaligned border, and the two functions share
  one implementation — effort: 2 (replan: TRIAGE T26, T28)
- **M8.3** — De-emphasized secondary text: the Sources Nav path line and Picker File-row paths
  render dim with matched path characters emphasized against the dim, the Root label at normal
  weight; the Root's parent-path fragment is dropped from Picker File rows (D117) — done when:
  render tests assert the styles, the fragment's absence, and C1/D22's disambiguation tests are
  retired with a note — effort: 2 (needs: M7.3)
- **M8.4** — `scrollActions`' doc comment (viewer.go:601, carried through M8.1's own edit)
  cites "the help modal (S12)", but this cycle's SPEC.md now assigns S12 to M8.1's own
  half-page-paging item — the same bare, undisambiguated "(S12)" recurs at footer.go:168,
  keymap.go:15 and files_search.go:18/283 (pre-existing, all now equally ambiguous against the
  current SPEC.md); `picker_view.go:192` already shows this project's disambiguation
  convention ("S12/C2") — done when: each Help-Modal reference either adopts the same
  "S12/C\<n\>" form or drops the stale S-number where it no longer resolves unambiguously —
  effort: 1 (replan: TRIAGE T27)
- **M8.5** — M8.4 fixed only the 5 bare `(S12)` sites its own AC named; the same now-ambiguous
  bare `S12` recurs at ~20 more sites across footer.go, keymap.go, files_search.go, picker.go,
  picker_view.go, picker_view_test.go, help.go, help_test.go, footer_test.go and
  files_search_test.go (a mix of C1's help-modal sense and C2's query-threshold/result-cap
  sense) — done when: every remaining bare `S12` citation in `cmd/noggin` is disambiguated to
  `S12/C1` or `S12/C2` per M8.4's own convention (matching whichever sense the citation's
  context actually means), and `grep -rn "S12" cmd/noggin | grep -v "S12/C"` is empty — effort:
  2 (replan: TRIAGE T32)

## M9 — Focus by background

Delivers: S11 — *pre-agreed cut 3*

- **M9.1** — Unfocused panes render on a subtly lighter background tint; the focused pane
  keeps the terminal background; the border-colour cue stays — done when: frame tests assert
  the tint on every unfocused pane and its absence on the focused one across all four focus
  targets, and a `make check-pty` check confirms the tint reaches terminal cells — effort: 3
- **M9.2** — `tintPaneBackground`'s SGR-7 (reverse-video) detector (style.go:120-137) matches
  raw `;`-split parameter tokens by bare string equality (`"7"`, `"27"`), so a numeric
  parameter of exactly 7 or 27 belonging to an extended-color spec (`48;5;7m`, `38;5;27m`, or
  any `38;2;`/`48;2;` truecolor component equal to 7 or 27) is misread as reverse set/clear and
  the run silently loses its S11 tint — not hit by any current palette constant, but a latent
  correctness gap for any future style pick in that range (e.g. glamour-rendered markdown) —
  parse SGR params positionally (skip the parameter(s) following a `38`/`48` token) instead of
  testing every token independently — done when: `tintPaneBackground("\x1b[48;5;7mrow\x1b[m")`
  (and the truecolor equivalent) render tinted, and the existing reverse-video tests still pass
  — effort: 2 (replan: TRIAGE T29)
- **M9.3** — M9.2's own doc comment (style.go:120-137) claims a colour-index/truecolor
  component of 7, 27 **or 0** must never be misread as reverse set/clear or a reset, but its
  regression test (`TestTintPaneBackgroundNotFooledByExtendedColorParams`) only covers
  components 7 and 27 across its four subtests — none exercises 0, so the comment's reset
  ('0') claim is untested — done when: a subtest covering a 256-colour or truecolor component
  value of 0 (e.g. `\x1b[48;2;0;100;200m`) is added and asserts the run still gets tinted —
  effort: 1 (replan: TRIAGE T34)

## M10 — Product documentation

Delivers: README coverage of every user-visible change in S1–S13

- **M10.1** — Update `README.md` for Sources (Sources Nav, Wikipedia Source, Wikipedia Picker,
  articles in Starred/Recent, Root names, location line, half-page paging, 2-char threshold,
  background focus), with every key and label verified against the implementation — done
  when: `grep -n "Roots Nav\|No roots\|at least 3" README.md` is empty and each new binding in
  the README matches `keymap.go` — effort: 2 (needs: M1–M9 as built)
