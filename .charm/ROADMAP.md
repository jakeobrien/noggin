# Roadmap

<!-- Durable, cross-cycle. Intent above the cycle, plus the cycle ledger. Create lazily — a
     single-cycle project does not need it. Owned by /brainstorm (scope) and /charm-close
     (harvest); charm-replan writes Unscheduled; charm-build never reads this file.

     COARSE BY CONSTRUCTION. Entries are outcomes, not tasks: no acceptance criteria, no
     effort estimates, no milestones, no priority scores. Detail is SPEC.md's job, once an
     entry is bet on. Speculative detail here will be wrong by the time it matters.

     Cycle keys (C1, C2…) are durable: allocated here, stamped into the cycle docs' headers,
     and leading the archive dir name (archive/C3-2026-07-name/) so a key resolves to its
     archive without a lookup. Never renumber — order is table position, so resequencing is
     free. -->

## North star

<!-- A few lines: what this project is for, and what good looks like beyond the current
     cycle. The "why" above any single spec. -->

<north star>

## Cycles

<!-- Past, active and planned, in order. Statuses: planned | active | done | dropped.
     Planned rows are a key and a sentence. -->

| Key | Name     | Status  | Outcome / intent                  | Archive                        |
|-----|----------|---------|-----------------------------------|--------------------------------|
| C1  | <name>   | done    | <one line: what it delivered>     | `archive/C1-<yyyy-mm>-<name>`  |
| C2  | <name>   | active  | <one line: what it is for>        | -                              |
| C3  | <name>   | planned | <one line>                        | -                              |

## Unscheduled

<!-- Work-shaped items deferred beyond the cycle that surfaced them (routed here by
     charm-replan), plus ideas not yet bet on. DELIBERATELY LOSSY: pruned at every
     /charm-close; forgetting is the feature. Not a backlog and not a queue — no statuses,
     no durable ids. An item that survives several prunes is telling you to schedule it; one
     nobody misses should be dropped and re-raised if it still matters. One line each; cite
     like findings. Pin to a cycle key only where it genuinely belongs to one. -->

- <one line> (from C2; `src/foo.ts:42`) <!-- optionally: → C3 -->
