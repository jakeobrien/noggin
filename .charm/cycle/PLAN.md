# Plan — C4a wikipedia-spike

<!-- Per-cycle. The spec decomposed into milestones and tasks. Technically aware but
     high-level: state WHAT and the done-condition, not step-by-step HOW — trust executor
     judgment. Every task MUST have an acceptance criterion, machine-checkable where possible,
     and an effort estimate: a relative 1-5 gut-scoping (1 trivial, e.g. fixing a few typos;
     5 very high, e.g. large or complicated tasks) — not a rigorous prediction, just how the
     task looks at scoping time. Durable keys: milestones M1, M2...; tasks M1.1, M1.2...
     Never renumber; append or suffix (M1.2a) when inserting. Keys are scoped to this cycle —
     durable docs referencing them qualify with the cycle key (C3/M1.2). Note dependencies
     where ordering matters. -->

Shape: three build milestones (fetch → convert → judge/retune) then a verdict milestone.
M1 and M2 exist only to make M3 possible; M3 is the cycle. Retuning (S6a) lives inside M3
as a deliberately repeatable task, not a milestone of its own.

<!-- Reconciled 2026-09-02 (charm-replan) against SPEC.md's D104 amendment: S3/S6/S7 ->
     S3a/S6a/S7a, widening the retune budget to two costed experiments (maths, images).
     M3.4 rewritten, M3.5 added, M4.1 widened. See DECISIONS.md D104. -->

<!-- Reconciled 2026-09-02, later same day (charm-replan) draining TRIAGE T2-T7: added
     M2.4a (T3), M3.3a (T2), M3.6 (T5/T6/T7); M4.1 widened for T4 and re-dependencied on all
     three. See DECISIONS.md D105, D106. T4/T5/T6/T7 needed no decision entry (low severity,
     mechanical fixes with an obvious answer). -->

## M1 — Fetch an article

Delivers: S1 (entry point), S2, S8

- **M1.1** — Scaffold a throwaway command-line spike as its own `main`, separate from the
  noggin TUI: takes an article title argument and an output path, writes a file. Clearly
  marked experimental — done when: `make check` passes and no file under `cmd/noggin/` or
  `internal/` references the spike package (grep clean) — effort: 2
- **M1.2** — Fetch article HTML from Wikimedia's public REST API by exact title, sending a
  descriptive User-Agent per Wikimedia policy — done when: the spike fetches all four S5
  sample titles (spaces, parentheses, `?`) successfully, and the User-Agent header it sends
  is asserted by a test — effort: 2 (needs: M1.1)

## M2 — Convert it to markdown worth reading

Delivers: S1 (conversion), S3a (chrome removal only — the image placeholder half of
S3a is M3.5's costed experiment, not this milestone), S4

- **M2.1** — Convert the fetched HTML to markdown, tables included — done when: running the
  spike on a sample article writes a `.md` file with headings, paragraphs and at least one
  markdown table — effort: 2 (needs: M1.2)
- **M2.2** — Strip everything that is not the article before conversion: navboxes, infoboxes,
  edit-section links, citation superscripts, reference lists, metadata tables, images.
  Content tables survive — done when: the four converted samples contain no `[edit]`, no
  reference-list section, no image markup, and no infobox/navbox table, while
  `Major Arcana`'s content tables are still present — effort: 3 (needs: M2.1)
- **M2.3** — Strip link markup, keeping the link's text as ordinary prose — done when: the
  four converted samples contain no markdown link or URL syntax, and previously-linked
  phrases still read as prose in context — effort: 2 (needs: M2.1)
- **M2.4** — Leave no holes where chrome was cut: no orphaned captions, empty table rows,
  dangling brackets, or runs of blank lines — done when: no converted sample contains three
  or more consecutive blank lines, an empty list item, or an unmatched `[`/`(` left by a
  removal, checked mechanically over all four — effort: 3 (needs: M2.2, M2.3)
- **M2.5** — Consolidate the per-task opt-in live tests (M1.2, M2.2, M2.3, M2.4 each added
  their own, TRIAGE T1) into one shared per-title live fixture covering fetch+strip+links+
  convert, so `WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v` no longer trips the
  REST API's 429 when run as a whole — done when: a single live test (or a shared fetch
  cached across subtests) exercises all four S5 titles once each, the old per-task Live
  tests are removed or reduced to assertions against its shared result, and a whole-suite
  `-run Live` run passes without a 429 — effort: 2 (needs: M2.2, M2.3, M2.4)
- **M2.4a** — *Fix-up on M2.4 (TRIAGE T3, charm-replan 2026-09-02):* `cleanupProse`'s
  bracket-balance handling only catches literal empty pairs (`reEmptyParens`), not a real
  dangling `)`/`(` left after a removal — `tmp/wikispike-tests-1/list_languages.md` has one
  (386 `(` vs 388 `)`). Give `cleanupProse` a real paren/bracket-depth pass, or a targeted
  fix for the specific removal that orphans it — done when: all four current converted
  samples have balanced `(`/`)` and `[`/`]` counts, checked mechanically, and M2.4's
  PROGRESS.md note no longer needs qualifying against this gap — effort: 2 (needs: M2.4)

## M3 — Read the sample in noggin and judge it

Delivers: S5, S6a

- **M3.1** — Convert exactly the four S5 sample articles into a scratch root outside the repo
  — done when: the scratch root holds four `.md` files, one per sample title, each non-empty
  and produced by the spike — effort: 1 (needs: M2.4)
- **M3.2** — User reads all four in noggin's own Document Viewer at real pane widths and
  judges them against A1–A4 — done when: a per-article A1–A4 verdict from the user is
  recorded in the findings file, including which failures are content noggin deliberately
  does not render (A5) — effort: 2 (needs: M3.1)
- **M3.3** — Retune what gets stripped and re-judge. Repeatable; bounded by the cycle and by
  "no new capability to buy quality" — done when: either every A1–A3 shortfall the user
  raised is fixed or recorded as retune-exhausted with the reason, and no new capability was
  added to achieve it — effort: 3 (needs: M3.2)
- **M3.4** — *S6a's maths exception (D104):* try a Unicode textual approximation of
  `E (mathematical constant)`'s formulae first (fractions, sub/superscripts, roots,
  sums/integrals, Greek letters and similar) — a new third-party dependency is permitted for
  this specifically. If that is not a small-ish lift, fall back to preserving each formula's
  raw source text so the surrounding sentence still reads (already permitted, unchanged) —
  done when: either the Unicode approximation ships with no mid-sentence gaps from dropped
  formulae, or raw-text preservation ships with no mid-sentence gaps and findings record why
  Unicode was abandoned, or the whole attempt is abandoned and recorded as not-low-effort —
  effort: 3 (needs: M3.3)
- **M3.5** — *S6a's image exception (D104):* replace a stripped image with a plain-text
  placeholder derived from its caption, falling back to alt-text, falling back to filename —
  no clickable link (the image's link is stripped along with every other link, M2.3). Try the
  near-zero-cost path first — stop deleting the image node in M2.2's chrome strip and let the
  existing render pipeline carry the placeholder through — before writing new extraction
  logic — done when: each of the four S5 samples that has an image now shows a placeholder
  in its place with no dangling brackets/link markup and no stray caption divorced from where
  the image was, or the attempt is abandoned as not-low-effort and recorded — effort: 3
  (needs: M3.3)

- **M3.3a** — *Fix-up on M2.2 (TRIAGE T2, charm-replan 2026-09-02):* `stripChrome` doesn't
  match Wikipedia's portal-bar (`div.portal-bar`) or sister-project side-box
  (`div.sistersitebox`) chrome, both navigational, not article content — discovered via
  M3.5, previously masked because their icon images vanished without a trace, now visible as
  leaked list items / `[Image: icon]` placeholders (`why_anything.md`'s
  Philosophy/Physics/Astronomy/Religion/Science line; `e_math.md`/`major_arcana.md`'s
  Wikimedia Commons/Wikiquote lines) — done when: `stripChrome` also removes both selectors
  and none of the four current converted samples shows a leaked portal-bar or
  sister-project line — effort: 2 (needs: M2.2, M3.5)
- **M3.6** — *Housekeeping pass (TRIAGE T5/T6/T7, charm-replan 2026-09-02):* low-severity
  code-review findings bundled into one pass — (T5) fix the stale doc comment on
  `TestStripLinksPreservesSurroundingProse` (`links_test.go:33-36`, still names the old
  identifier); (T6) factor the duplicated goquery parse/re-serialize boilerplate
  (`strip.go`, `links.go`, `math.go`, `images.go`) into one shared helper, and give the
  `mw:File`/`mw:Extension/math` `typeof` prefixes a shared constant instead of parallel
  string literals; (T7) fix the "placeheld" typo in `images.go`'s error message to
  "placeholder" — done when: all three are addressed and `make check`/`go test ./...` stay
  clean — effort: 2 (needs: M3.4, M3.5)

## M4 — Record the verdict

Delivers: S7a

- **M4.1** — Findings record: what converts well, what converts badly, what could not be
  fixed within the retune budget, and the selector/cleanup list as it ended up — plus, for
  each of S6a's two experiments (maths M3.4, images M3.5): what was tried and in what order,
  effort actually spent, any dependency added, resulting quality by eye, and a recommendation
  on whether it is worth carrying into a future real Wikipedia feature. Also (TRIAGE T4):
  cite/regenerate the current scratch-root samples the findings are based on and remove or
  replace any stale root-level sample copies predating the M2.2/M3.4/M3.5 fix-ups, so a later
  reader doesn't diff against them by accident — done when: a cited findings file exists,
  indexed in `findings/INDEX.md`, covering all of the above, with enough detail that a later
  cycle need not re-run the experiment, and no stale sample copy remains uncited or
  unreplaced — effort: 3 (needs: M2.4a, M3.3, M3.3a, M3.4, M3.5, M3.6)
- **M4.2** — Log the explicit go / no-go / retune-exhausted verdict — done when:
  `DECISIONS.md` carries a `decided` entry stating one of the three verdicts, referencing
  M4.1's findings and the roadmap entries it gates — effort: 1 (needs: M4.1)
