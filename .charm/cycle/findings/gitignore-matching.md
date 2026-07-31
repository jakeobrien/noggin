# .gitignore matching for the tree walk

Question: how should M2.1's walk honour `.gitignore` — a dependency, shell-out, or
hand-rolled? And what subset does a read-only browse actually need? (Asked building M2.1.)

## The choice

- **Shell-out (`git check-ignore`) is wrong for our inputs.** It only answers within a real
  git worktree and consults that repo's whole ignore hierarchy (global excludes,
  `.git/info/exclude`, parent `.gitignore`s). noggin roots are arbitrary directories, many not
  git repos at all; a fixture tree under `t.TempDir()` is not a worktree. Also needs `git` on
  PATH at runtime.
- **A dependency adds weight to a tracked concern.** The dep graph is watched for
  security-review (PROGRESS.md notes the v2 migration's churn). `sabhiram/go-gitignore` is tiny
  but has its own `**`/anchoring gaps; go-git's `plumbing/format/gitignore` is more correct but
  drags in go-git. Neither is clearly better than a focused hand-roll for our subset.
- **Hand-rolled, chosen.** `internal/tree/gitignore.go`, ~40 significant lines. We own the
  behaviour and test it directly (`gitignore_test.go`, 18 match cases + negation + nested
  precedence).

## Subset covered (what a browse needs)

- Per-directory `.gitignore` files, accumulated root→leaf; later files and later lines win;
  `!` negation re-includes (git's last-match-wins). (`ignored`, `gitignore.go`)
- Comments (`#`), blank lines, escaped leading `\#`/`\!`, trailing-whitespace trim.
- Directory-only (trailing `/`); anchoring (a leading or embedded `/` pins the pattern to the
  `.gitignore`'s own directory, else it matches at any depth via `(?:.*/)?`).
- Wildcards `*` and `?` (do not cross `/`), `[...]` classes with `!`/`^` negation, and `**`
  (`**/` → zero-or-more dirs, trailing `/**` → everything under, bare `**` → `.*`).
- Compiled to one `regexp` per line; matched against the entry's path relative to the owning
  file's directory.

## Deliberately left out (and why it is safe)

- **Global excludes, `.git/info/exclude`, `core.excludesFile`.** Out of scope: noggin shows
  what is on disk under a root, not what a given git config would hide. `.git` itself is
  excluded by a built-in name check in the walk, independent of any `.gitignore`.
- **Re-including a file under an already-excluded directory.** git forbids it and so do we, for
  free: an ignored directory is *pruned* in `walkDir` (never descended), so its children are
  never evaluated. This is also why S6's "folders with nothing readable are still shown" holds
  only for *non-ignored* folders — an ignored dir is gone entirely.
- **Exotically escaped class internals / `[a-z]` range edge cases** beyond straight pass-through
  to `regexp`. Real `.gitignore`s in doc trees don't lean on these; add cases to
  `gitignore_test.go` if a real tree proves otherwise.

If a real root ever reveals a gap, the fix is a case in `gitignore_test.go` and a few lines in
`translateGlob`/`compilePattern`, not a dependency swap.
