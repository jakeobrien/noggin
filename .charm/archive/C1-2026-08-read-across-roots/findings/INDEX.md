# Findings index

- `frontmatter-rendering.md` — **promoted to `.charm/knowledge/frontmatter-rendering.md`** at
  close (C1).
- `bubbletea-v2-migration.md` — the verified Charm v2 surface: the `charm.land` module-path
  rename, the all-or-nothing coupling, every API break that touches our ~600 lines, and what v2
  buys (viewport search for S8, D13's hazard designed out, TTY-free test options). Settled D14;
  migration completed in M1.4. Archived with the cycle: every hazard it names now has its own
  code comment at the point of use, and the decision itself is durable via D14.
- `pty-verification.md` — **promoted to `.charm/knowledge/pty-verification.md`** at close (C1).
- `glamour-in-bubbletea.md` — `WithAutoStyle` queries the terminal at render time and deadlocks
  against Bubble Tea's input reader; Glamour also leaks literal tabs into rendered code fences.
  Both fixed in M1.3, re-verified under v2 in M1.4. Archived with the cycle: fully captured by
  D13/D16 and `document.go`'s own comments.
- `gitignore-matching.md` — **promoted to `.charm/knowledge/gitignore-matching.md`** at close
  (C1).
- `go-toolchain-env.md` — **promoted to `.charm/knowledge/go-toolchain-env.md`** at close (C1).
- `viewport-search-ansi-offsets.md` — **promoted to
  `.charm/knowledge/viewport-search-ansi-offsets.md`** at close (C1).
