# Progress — C1 read-across-roots

## Cursor

Current task: M1.2

## Handoff

M1.1 is done: `noggin` module scaffolded (module path `noggin`, `go 1.26.5`), `cmd/noggin`
holds a minimal Bubble Tea program (`model` with no state yet) that renders a placeholder
string and quits cleanly on `q` / `ctrl+c`. `Makefile` provides `build` (`go build -o bin/
./...`), `vet`, `fmt` (`gofmt -l .`, fails on non-empty output), `test` and `check` (all
four), plus `unexport GOROOT` for host portability; `bin/` is gitignored.

Verified: `make check` is clean (build/vet/fmt/test all pass, no test files yet). Quit
behavior verified via a pty harness (python3 `pty` module) driving the built `bin/noggin`
binary directly — both `q` and `ctrl+c` exit with code 0 and the output stream contains
Bubble Tea's terminal-restore sequences (`\x1b[?2004l` bracketed-paste off, `\x1b[?25h`
cursor shown, mouse-tracking off), confirming the terminal is restored, not just that the
process exits.

Next step: M1.2 — lay out the frame (Navigator column with two stacked panes, Document
Viewer, Footer), focus movement between panes, resize reflow, all bindings through one
keymap definition (D2). `model.go` currently has no layout/dimensions handling at all —
M1.2 will need to start tracking `tea.WindowSizeMsg` and introduce the keymap file the
acceptance criterion's `rg` check depends on.

Gotchas still relevant:
- Bubble Tea `@latest` resolved to v1.3.10 (v2 is pre-release); confirmed live via `go get`
  during M1.1. Glamour v1.0.0 is out, not yet added (needed at M1.3). Go with the grain (D1).
- M1.3 renders a document from a bare path argument; M3.1 replaces that ad-hoc argument
  handling with the real S1 launch contract (D11).
- Sandbox Go toolchain is fully working as of 2026-07-31 (go1.26.5, gcc/cgo,
  gopls/staticcheck/goimports all present); see `findings/go-toolchain-env.md`. The
  Makefile's `unexport GOROOT` is defensive-only, not compensating for a live break.

No blockers.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
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
