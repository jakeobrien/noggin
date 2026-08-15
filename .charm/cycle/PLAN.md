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
M1 and M2 exist only to make M3 possible; M3 is the cycle. Retuning (S6) lives inside M3
as a deliberately repeatable task, not a milestone of its own.

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

Delivers: S1 (conversion), S3, S4

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

## M3 — Read the sample in noggin and judge it

Delivers: S5, S6

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
- **M3.4** — *Optional, S6's single exception:* a low-effort fallback that keeps a formula's
  text or drops it cleanly so the surrounding sentence still reads — done when: either
  `E (mathematical constant)` has no mid-sentence gaps from dropped formulae, or the attempt
  is abandoned and recorded as not-low-effort — effort: 2 (needs: M3.3)

## M4 — Record the verdict

Delivers: S7

- **M4.1** — Findings record: what converts well, what converts badly, what could not be
  fixed within the retune budget, and the selector/cleanup list as it ended up — done when:
  a cited findings file exists, indexed in `findings/INDEX.md`, covering all three, with
  enough detail that a later cycle need not re-run the experiment — effort: 2 (needs: M3.3)
- **M4.2** — Log the explicit go / no-go / retune-exhausted verdict — done when:
  `DECISIONS.md` carries a `decided` entry stating one of the three verdicts, referencing
  M4.1's findings and the roadmap entries it gates — effort: 1 (needs: M4.1)
