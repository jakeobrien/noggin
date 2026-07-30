# Go toolchain in this sandbox

Question: is the Go toolchain usable here, and can modules be fetched? (Asked while planning
C1, before any code existed.)

## Findings

- `go` resolves to `/usr/bin/go` → `/usr/lib/go-1.26/bin/go`, but the environment sets
  `GOROOT=/home/agent/.local/go`, which does not exist — every invocation fails with
  `go: cannot find GOROOT directory`.
- `GOROOT=/usr/lib/go go version` → `go version go1.26.0 linux/arm64`. Overriding per command
  (or exporting once per shell) is sufficient; nothing else is broken.
- Module fetching works through the sandbox proxy: `go get
  github.com/charmbracelet/bubbletea@latest` succeeded in a scratch module.
  `GOMODCACHE=/home/agent/go/pkg/mod`, `GOTOOLCHAIN=auto`.
- Version landscape at plan time: `bubbletea@latest` = v1.3.10 (v2 exists only as
  pre-release, so `@latest` does not pick it); `glamour` has a v1.0.0; `bubbles` v0.21.1.

## Consequence

Build and test commands need `GOROOT=/usr/lib/go`. Worth putting in whatever
Makefile/justfile the scaffold gets so it is not rediscovered each session.
