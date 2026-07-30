# Progress — C1 read-across-roots

## Cursor

Current task: M1.1

## Handoff

Planning is done; no code exists yet. The repo holds only `NOTES.md` (raw brainstorm input,
superseded by `SPEC.md`) and `.charm/`.

Next step: M1.1 — scaffold the Go module (module path `noggin`-ish, `cmd/noggin`) and a
minimal Bubble Tea program.

Gotchas for the first build session:
- `GOROOT` is misset in this sandbox (`/home/agent/.local/go`, absent). Export
  `GOROOT=/usr/lib/go` (go 1.26.0). Module downloads through the proxy work.
  See `findings/go-toolchain-env.md`.
- Bubble Tea `@latest` resolves to v1.3.10 (v2 is pre-release); Glamour v1.0.0 is out. Take
  latest stable and go with the grain (D1).
- M1.3 renders a document from a bare path argument; M3.1 replaces that ad-hoc argument
  handling with the real S1 launch contract (D11).

No blockers.

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
