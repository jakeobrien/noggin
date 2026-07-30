# Findings index

- `frontmatter-rendering.md` — Glamour ignores YAML frontmatter and CommonMark mis-renders it
  as an H2; Glow pre-strips the bytes. Settles C1/S7's frontmatter behavior and flags the
  cost of ever needing frontmatter *values* at render time.
- `go-toolchain-env.md` — the sandbox's `GOROOT` is misset; build/test needs
  `GOROOT=/usr/lib/go`. Module fetching works. Charm library versions at plan time.
