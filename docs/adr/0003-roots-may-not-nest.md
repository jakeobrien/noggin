# Roots may not nest

A Root may be neither a parent nor a child of another Root. A path given inside an existing Root activates that covering Root. A path that is a parent of an existing Root is rejected. This is a sharp edge on purpose: `noggin ~` is refused while `~/notes` is starred. In return, "which Root is this document in" always has exactly one answer, and search scope and the per-Root Starred and Recent lists depend on that answer.

## Consequences

- The Document Store (ADR-0005) keeps no Root attribution and gets per-Root lists by path prefix. That only works because Roots don't nest. Allowing nesting would mean redesigning how documents belong to Roots.
- "Shadow Roots" (retained, unstarred Roots allowed to nest) were rejected partly because they would break this rule.

_Formerly D6._
