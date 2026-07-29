# Decisions

<!-- Durable, cross-cycle. One entry per decision. Statuses: open | decided | superseded.
     Decisions are judgments, not a backlog. Via the escalation matrix: medium risk/value ->
     do + log here (decided); high + non-blocking -> skip + log the open JUDGMENT QUESTION
     here — deferred work-shaped items go to cycle/TRIAGE.md instead, and if replan defers
     them past this cycle, to ROADMAP.md Unscheduled. Blockage is tracked in PROGRESS.md
     (blocked rows reference D keys); no blocking field here. Keep entries short; the "why"
     matters most.

     Every entry stamps its `Cycle:` — bare M/S keys inside an entry resolve against that
     cycle, permanently. A reference to another cycle must qualify it (C2/M1.2), since key
     namespaces reset each cycle.

     Retention is a hindsight judgment made at /charm-close, not at write time: entries that
     no longer constrain future work are retired (moved to the closing cycle's archive), and
     the enduring ADRs stay. Never delete an entry here; retire it, or mark it superseded and
     link the replacement. -->

## D1: <title>

- Status: open | decided | superseded (see D<n>)
- Cycle: C<n>
- Date: <yyyy-mm-dd>
- Context: <one or two lines: what forced the decision>
- Decision: <what was decided, or the open question if status is open>
- Why: <rationale / tradeoffs; alternatives rejected and why, if notable>
