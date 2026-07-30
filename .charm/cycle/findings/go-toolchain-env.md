# Go toolchain in this sandbox

Question: is the Go toolchain usable here, and how should build commands interface with it?
(Asked while planning C1, before any code existed.)

## What is actually installed

- `go` on PATH resolves to `/usr/bin/go` → `/usr/lib/go-1.26/bin/go`, the **apt** toolchain
  from the base image (`dpkg -l golang` → `2:1.26~1`, `go version go1.26.0 linux/arm64`).
- The **upstream tarball toolchain the go recipe installs is absent**:
  `/home/agent/.local/go` does not exist, `/home/agent/.local/bin` contains only `fd`, and
  `gopls` / `staticcheck` / `goimports` are nowhere on the filesystem.
- `build-essential` is not installed (`dpkg -l` → `un`, no `gcc`/`cc`), so the go recipe's
  root layer did not run either. This container is the **base** image; only the go recipe's
  `ENV` block is in effect (`/proc/1/environ`: `GOROOT=/home/agent/.local/go`,
  `GOBIN=/home/agent/.local/bin`, `GOTOOLCHAIN=auto`, PATH prefixed with
  `/home/agent/.local/go/bin`).
- Net effect: `GOROOT` points at a directory that does not exist, so **every** `go`
  invocation fails with `go: cannot find GOROOT directory: /home/agent/.local/go` — the apt
  toolchain is broken by an env var meant for a toolchain that isn't here.

## How to interface with it

- Preferred: **unset `GOROOT`**. `env -u GOROOT go version` → `go1.26.0`; `go env GOROOT` →
  `/usr/lib/go-1.26`. A Go binary infers its own root from its location, so unsetting works
  for whichever toolchain is on PATH — the apt one now, the tarball one after an image fix.
- `GOROOT=/usr/lib/go` also works but hard-codes the apt path; it would be wrong once the
  upstream toolchain is provisioned.
- `GOBIN=/home/agent/.local/bin` is on PATH, so `go install` tools land somewhere usable.
- Module fetching works through the proxy (`go get github.com/charmbracelet/bubbletea@latest`
  succeeded). `GOMODCACHE=/home/agent/go/pkg/mod`, `GOTOOLCHAIN=auto`.
- No C compiler: cgo builds will fail. Irrelevant for pure-Go Bubble Tea / Glamour.
- No `staticcheck`, `goimports` or `gopls`: verification must lean on `go vet`, `gofmt -l`
  and `go test`, or install the tools on demand (`go install` works).
- Version landscape at plan time: `bubbletea@latest` = v1.3.10 (v2 is pre-release only, so
  `@latest` does not pick it); `glamour` has v1.0.0; `bubbles` v0.21.1.

## Consequence

Every build/test command in this cycle runs with `GOROOT` unset — put that in the
Makefile/justfile the scaffold gets (`unexport GOROOT` in make, or a `GOROOT=` prefix) so it
is not rediscovered each session, and so the project still builds on a host with a normal Go
install.

Image-side fixes (owner: the sandbox recipe, not this cycle) are recorded in the handoff and
below.

## Recommended recipe changes

1. **Drop `GOROOT` from the go recipe's `ENV`.** It buys nothing — the tarball's `go` binary
   infers its root from its own path — and it is exactly what converts "toolchain missing" or
   "toolchain moved" into "no Go at all", including breaking the apt fallback. Keep `GOPATH`,
   `GOBIN`, `GOTOOLCHAIN`, and the PATH prefix.
2. **Confirm the container is actually launched from `sandbox-pi-go`.** Evidence above says
   this VM runs base with go-recipe env applied — a tag/config mismatch on the host side.
3. `ARG TARGETARCH=arm64` — drop the default. BuildKit always supplies it; the default only
   silently installs an arm64 toolchain into an amd64 image if it ever isn't set.
4. Verify the tarball against the published `go${GO_VERSION}.linux-${TARGETARCH}.tar.gz.sha256`
   before extracting — the current pipe trusts the download outright.
5. `go install ...@latest` for the tool belt makes the image non-reproducible; pin versions if
   that matters, or accept it deliberately.
