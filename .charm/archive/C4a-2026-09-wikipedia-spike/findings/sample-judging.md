# M3.2 — sample judging verdict (A1–A5)

Recorded 2026-09-02, from the user's read of all four S5 samples in noggin's Document
Viewer at real pane widths (scratch root outside the repo, per M3.1).

## Per-article verdict

| Article                                   | A1 (readable) | A2 (headings) | A3 (no gaps/artifacts) | A4 (tables) |
|--------------------------------------------|:---:|:---:|:---:|:---:|
| `Why is there anything at all?`            | pass | pass | pass | pass |
| `E (mathematical constant)`                | pass | pass | **fail** | pass |
| `List of programming languages by type`    | pass | pass | pass | pass |
| `Major Arcana`                             | pass | pass | pass | pass |

**A3 failure detail (`E (mathematical constant)`):** stripped maths formulae leave visible
gaps in the prose — no other artifacts (no stray brackets, no orphaned captions).

## A5 — known-limitation call

The one A3 shortfall is exactly the case A5 names: content noggin deliberately does not
render (maths), confined to one article, with the other three clearing A1–A3. Per SPEC.md's
own worked example, this is **go with a known limitation, not a no-go**, on the raw
(pre-M3.4/M3.5) conversion output.

## Note on the earlier "image gaps" mention

D104 / SPEC.md's amendment narrative (`DECISIONS.md:1256`) describes the swiss-cheese
motivation as "missing maths **and images**" from an earlier, less structured read. This
per-article A1–A4 pass recorded here shows no image-driven A3 failure in this specific
sample set — none of the four apparently lost enough to a stripped image to register as a
visible gap at pass/fail granularity. Recorded for the discrepancy's sake; it does not
change M3.5's scope, since S6a/D104 commits to the image-placeholder experiment as a costed
exploration regardless of whether this exact sample tripped it.

## Consequence for M3.3

No A1–A3 shortfall exists in this sample beyond the one above, and that one is already
carved into its own experiment (M3.4). M3.3 (generic retune) therefore has nothing left to
retune under its own mandate — see PROGRESS.md's Status table.
