# Sources keep pointers, never content

noggin stores pointers, stars and timestamps for every Source. It never stores document content. For a Root that costs nothing, since the files are on disk. For a Query Source it means every view of an Article is a fetch, so Articles are never grepped and have no Picker Preview. A Wikipedia Picker shows a placeholder instead. With a Query Source active, the File and Grep Pickers skip the Active Root scope instead of offering an empty one. Every network call stays user-initiated: search, Activate, Peek and Refresh. Lists that span Sources (All Starred, All Sources) can still include Articles, because they read the Document Store and fetch nothing.

## Consequences

- Grepping starred Articles, or previewing them, would mean caching content. That reverses this decision and needs its own answers on cache size, staleness and offline behaviour.
- A Query Source is a typed Source with conditional UX (no Tree, no grep), not a separate layer.

_Formerly D114 (architecture), D119._
