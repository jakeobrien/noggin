# Ignore rules: the Tree's matcher and ripgrep's

The Files Nav Tree and the Pickers apply `.gitignore` through different engines (ADR-0004), so they can disagree at the edges. This note covers what each one does.

## The Tree: a hand-rolled subset (`internal/tree/gitignore.go`)

A Root needn't be a git repository, so neither `git check-ignore` (only works inside a worktree, and consults the whole repo's ignore hierarchy) nor git config applies. The matcher covers what a browse needs:

- per-directory `.gitignore` files, accumulated root → leaf; later files and later lines win; `!` re-includes (last match wins);
- comments, blank lines, escaped leading `\#` / `\!`, trailing-whitespace trim;
- directory-only patterns (trailing `/`); anchoring (a leading or embedded `/` pins to the file's own directory, otherwise any depth);
- `*` and `?` (don't cross `/`), `[...]` with `!`/`^` negation, and `**` (`**/`, trailing `/**`, bare `**`).

Each line compiles to one `regexp`, matched against the path relative to the owning `.gitignore`'s directory.

Deliberately left out:

- **Global excludes, `.git/info/exclude`, `core.excludesFile`.** noggin shows what's on disk under a Root, not what one git config would hide. `.git` itself is excluded by name.
- **Re-including a file under an ignored directory.** git forbids it, and noggin gets that for free: ignored directories are pruned, never descended.
- **Exotic `[...]` escapes and range edge cases** beyond passing them through to `regexp`.

If a real Root reveals a gap, add a case to `gitignore_test.go` and a few lines in `translateGlob`/`compilePattern`.

## The Pickers: ripgrep (`cmd/noggin/files_search.go`)

- **`--no-require-git`** is required: by default `rg` only honours `.gitignore` inside a git repository, so a plain-folder Root's `.gitignore` would be silently ignored.
- **`rg` doesn't apply ignore rules to files named on its command line.** Grep therefore searches the explicit file list the File Picker produced, rather than letting `rg` walk again.
- **Showing ignored files** runs `rg` with `--no-ignore-vcs`; the Tree stops applying `.gitignore`. A fixed set of heavy folder names (declared once in `internal/tree`) stays excluded either way.
