# Document memory is one global store keyed by path, not attached to Roots

Starred and last-viewed state for every document lives in one persisted store, keyed by absolute path (or canonical URL for an Article). It isn't attached to Roots. Per-Root Starred and Recent lists are derived by path prefix (ADR-0003). No Root lifecycle event destroys document state. A document starred under a Root that is later un-starred stays starred, and becomes an Orphan if no Root covers it. Recency for Ad-hoc Roots is persisted like everything else.

## Considered Options

- **Shadow Roots**: a retained-but-unstarred Root class with a TTL and a cap, allowed to nest. It would have held recency for Ad-hoc Roots. Rejected because it adds a durable concept with time-based expiry and no visible surface, and because it breaks ADR-0003.
- **Purge document stars when a Root is un-starred** (behind a confirm dialog): rejected. Un-starring a Root means "stop opening it at launch", not "I'm done with these documents". A purge can't be undone, and it would also remove stars that existed before the Root was starred.

_Formerly D61._
