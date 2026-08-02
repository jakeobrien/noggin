# Knowledge index

<!-- Durable, cross-cycle. Promoted findings: knowledge that outlived the cycle that learned
     it. Same format and rules as cycle/findings/ — one topic per file, condensed, cited
     (filepath:line, URLs), statusless reference material, never raw dumps. Read via this
     index, one topic at a time.

     Promotion happens at /charm-close, not at write time (hindsight is cheaper than
     foresight), and only if both hold: (1) it will still be true and useful next cycle, and
     (2) it has no better home in the codebase — how the system works belongs in docs/,
     README or a comment, where proximity to the code keeps it honest. What lands here is
     what has nowhere else to live: external/environment constraints, third-party quirks,
     and dead ends ("we tried X, here is why it does not work").

     Every file carries `As of: C<n>` — durable knowledge without a freshness stamp is a
     trap. A file that contradicts the code is a TRIAGE.md item, not a silent edit. Keep
     this tier small; nothing archives it automatically. -->

- `frontmatter-rendering.md` — Glamour ignores YAML frontmatter and CommonMark mis-renders it
  as an H2; cited by `document.go` at the frontmatter-stripping code. (as of C1)
- `gitignore-matching.md` — why the tree walk hand-rolls `.gitignore` instead of a dependency
  or shell-out, the subset covered, and what's deliberately left out; cited by
  `internal/tree/gitignore.go`. (as of C1)
- `go-toolchain-env.md` — this sandbox's Go toolchain constraints, resolved 2026-07-31; cited
  by the Makefile. Re-check before trusting if the environment changes again. (as of C1)
- `pty-verification.md` — the recipe for driving the real binary from a pty under Charm v2
  (startup handshake, cell-diffing screen emulation, border-scroll height check); cited by
  the Makefile and `viewer_test.go`. The basis for every `make check-pty` addition. (as of C1)
- `viewport-search-ansi-offsets.md` — bubbles v2's `viewport.SetHighlights` drifts off-line
  against ANSI-styled content; why `find.go` computes match lines itself instead. Re-check
  before ever reaching for `SetHighlights` again. (as of C1)
