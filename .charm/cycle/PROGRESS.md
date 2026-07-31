# Progress — C1 read-across-roots

## Cursor

Current task: M1.1

## Handoff

Planning is done; no code exists yet. The repo holds only `NOTES.md` (raw brainstorm input,
superseded by `SPEC.md`) and `.charm/`.

Next step: M1.1 — scaffold the Go module (module path `noggin`-ish, `cmd/noggin`) and a
minimal Bubble Tea program.

Gotchas for the first build session:
- **Resolved 2026-07-31**: the sandbox's `GOROOT` used to point at a toolchain that wasn't
  installed, breaking every `go` call; the container has since been repaired and re-validated
  fresh — `go` (go1.26.5), gcc/cgo, and `staticcheck`/`goimports`/`gopls` all now work with the
  environment exactly as inherited, no `GOROOT`-unset needed. The Makefile's defensive
  `GOROOT` neutralisation (`unexport GOROOT` or similar) is kept anyway for portability to a
  host where it isn't set correctly, not because this container needs it now. Module
  downloads through the proxy work. See `findings/go-toolchain-env.md`.
- Bubble Tea `@latest` resolves to v1.3.10 (v2 is pre-release); Glamour v1.0.0 is out. Take
  latest stable and go with the grain (D1).
- M1.3 renders a document from a bare path argument; M3.1 replaces that ad-hoc argument
  handling with the real S1 launch contract (D11).

No blockers. The sandbox image fix is done — the container now runs with a full go1.26.5
toolchain, gcc/cgo, and the gopls/staticcheck/goimports tool belt (re-validated 2026-07-31).
The finding's recipe recommendation to drop `GOROOT` from the ENV block was not literally
followed (`GOROOT` is still set, just now pointing at a real directory) — harmless in
practice, and not this cycle's concern either way.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | todo   | -        | -         |      |
| M1.2 | todo   | -        | -         |      |
| M1.3 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M2.2 | todo   | -        | -         |      |
| M2.3 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M3.2 | todo   | -        | -         |      |
| M3.3 | todo   | -        | -         |      |
| M3.4 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
