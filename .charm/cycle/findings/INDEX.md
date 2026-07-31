# Findings index

- `frontmatter-rendering.md` — Glamour ignores YAML frontmatter and CommonMark mis-renders it
  as an H2; Glow pre-strips the bytes. Settles C1/S7's frontmatter behavior and flags the
  cost of ever needing frontmatter *values* at render time.
- `pty-verification.md` — driving the real binary from a pty needs the OSC 11 + cursor-position
  startup queries answered or it hangs; `go test` never sees colour. The recipe for every later
  end-to-end check.
- `go-toolchain-env.md` — **resolved 2026-07-31**: the sandbox's `GOROOT` was misset,
  breaking `go`; re-validated fresh and the container now has a working go1.26.5 toolchain,
  gcc/cgo, and gopls/staticcheck/goimports, no workaround needed. Kept for history and Charm
  library versions at plan time.
