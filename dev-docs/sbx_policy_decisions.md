# SBX Policy Decisions

Records every agent request for a sandbox config change — network policy additions,
tool installations, or any other environmental adjustment. Declined requests must
**not** be raised again in the same session unless the user re-opens the topic.

---

2026-08-14

**Constraint:** HTTPS to `en.wikipedia.org` (and `api.wikimedia.org`) is denied by the
sandbox network policy — `403 Blocked by network policy: no matching allow rule`. The C4a
wikipedia-spike cycle exists to fetch and convert real Wikipedia articles, so this blocks
plan tasks M1.2 through M3.x entirely.

**Resolution:** `GRANTED` (2026-08-14) — verified: REST HTML 200 for all four sample
titles; Go's `net/http` works through the proxy unmodified. `api.wikimedia.org` also open.

**Workaround:** None taken. N/A once granted. While blocked, asked the user to allow-list `en.wikipedia.org` (only the
`/api/rest_v1/page/html/{title}` endpoint is needed). Deliberately did *not* substitute
saved HTML fixtures — that would invalidate the spike's premise. M1.1 (scaffolding) can
proceed offline in the meantime.

**Context:** `.charm/DECISIONS.md` D99; `.charm/cycle/PLAN.md` M1.2. The Go module proxy
(`proxy.golang.org`) *is* reachable, so adding `goquery` / `html-to-markdown` is fine.
