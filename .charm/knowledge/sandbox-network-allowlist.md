# Wikimedia domains needed explicit sandbox allow-listing

As of: C4a

Found while planning C4a/M1.2 (2026-08-14). `https://en.wikipedia.org/api/rest_v1/...` and
`https://api.wikimedia.org/...` both returned `Blocked by network policy: no matching allow
rule` from this sandbox by default. The Go module proxy was unaffected — only the two
Wikimedia hosts needed listing.

The user allow-listed `en.wikipedia.org`; verified after: REST HTML returns 200 for titles
containing spaces, parentheses and `?`, and Go's `net/http` reaches it through the sandbox
proxy with no TLS workaround needed.

Re-check before assuming this still holds if the sandbox environment resets or is rebuilt —
same caveat as `go-toolchain-env.md`'s own re-check note. Relevant to any future cycle that
resumes live Wikipedia network calls (e.g. C6, if it builds a real Wikipedia collection).
