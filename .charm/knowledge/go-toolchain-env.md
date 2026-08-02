# Go toolchain in this sandbox

As of: C1

Resolved 2026-07-31; kept for history and as an early-warning reference if the environment
regresses. Do not re-derive the diagnosis below without re-checking — it already changed once.

## Current state (validated fresh, 2026-07-31)

- `/home/agent/.local/go` is a genuine tarball install (`go1.26.5`); `go env GOROOT` resolves
  to it and `go version` runs correctly as inherited — no `GOROOT`-unsetting workaround
  needed in this container.
- `build-essential`/`gcc` is installed and functional; cgo builds work.
- `gopls`, `staticcheck`, `goimports` are all present and runnable from `$GOBIN`.
- A full `go build ./...` / `go vet ./...` / `gofmt -l .` / `go test ./...` loop and module
  fetch through the proxy both pass untouched.
- The Makefile's `unexport GOROOT` is kept anyway as a **defensive default for host
  portability** (cheap insurance for a host where `GOROOT` is set but wrong — see below),
  not because anything is currently broken here.

## Original problem, for context

Earlier in the container's life, `GOROOT` was set (via the go recipe's `ENV` block) to a
tarball path that didn't exist, breaking every `go` invocation even though a working apt
`go` was on PATH — a Go binary infers its own root from its own location, so an incorrect
inherited `GOROOT` is strictly worse than none. That state is gone; this note is what
resolving it looked like, in case it recurs after an image change.

## Recipe-side recommendations (unverified from inside the container, still open)

If anyone owns the sandbox image/recipe:
1. Drop `GOROOT` from the go recipe's `ENV` entirely (the tarball binary infers its own root).
2. `ARG TARGETARCH=arm64` — drop the default; BuildKit always supplies it, and the default
   only silently installs the wrong-arch toolchain if it's ever unset.
3. Verify the tarball against the published `.sha256` before extracting.
4. `go install ...@latest` for the tool belt makes the image non-reproducible; pin versions
   if that matters.

None of these are blocking — the environment works as-is — but they're the gap between
"works now" and "won't silently break again."
