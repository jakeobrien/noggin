---
title: 'Spec: keymap batch (remaps, Tab focus cycle, ctrl+w, left-arrow, view toggles)'
status: open
priority: 2
issue_type: epic
labels:
- spec
- ready-for-agent
created_at: 2026-10-15T16:20:33.602254377+00:00
updated_at: 2026-10-15T16:20:33.602254377+00:00
---

# Description

A batch of key remaps plus four behaviour changes, settled in a grilling session (2026-10). Decision record: `.charm/DECISIONS.md` **D141**.

## Remaps

- `h`/`j`/`k`/`l` no longer alias the arrow keys (LineUp/LineDown/Expand/Collapse are arrows only).
- Files Nav: `t` cycles Browse Mode (was `tab`); `a` cycles Browse Scope (was `shift+tab`).
- Picker: `alt+t` cycles mode (was `tab`); `alt+a` cycles Scope (was `shift+tab`).
- `ctrl+s` stars in every context, including the Picker (replaces `s` and `alt+s`).
- `d` focuses DV1 (was `o`); `D` focuses DV2 (was `ctrl+o`; still a no-op without DV2).
- `F` toggles the sidebar (was `ctrl+n`).
- `e` edits only from a focused Document Viewer; on the Sources Nav `e` renames the selected Root (was `c`); on the Files Nav it does nothing.
- Freed and left unbound: `o`, `ctrl+o`, `ctrl+n`, `ctrl+x`, `c`, `s`, `alt+s`, Picker `shift+tab`.
- Kept: `g`/`G`, `ctrl+u`/`ctrl+d`, `ctrl+b`/`ctrl+f`, `ctrl+l` (ActivateDV2 fallback).

## Tab focus cycle

`tab` cycles Focus forward Sources -> Files -> DV1 -> DV2 -> Sources; `shift+tab` reverses. Panes that are not visible are skipped (a hidden sidebar, or a terminal too narrow for it, removes Sources/Files; a missing DV2 is skipped). With only one focusable pane, Tab is a no-op. Never reveals the sidebar. Inert in Find, Rename, Help and the Picker.

## ctrl+w (replaces ctrl+x)

- DV1 focused, DV2 exists: DV1 takes DV2's whole viewer state (document, history, scroll, Render Mode); DV1's previous state is discarded; DV2 closes; Focus stays on DV1.
- DV1 focused, no DV2: DV1 is cleared to empty (document and history), like a fresh launch. Already empty: no-op. No Footer message.
- DV2 focused: DV2 closes; Focus moves to DV1.
- Sources/Files focused: no-op.

## Left arrow in the Files Nav (Tree)

- Expanded folder selected: collapse it (unchanged).
- File, or collapsed folder, inside a folder: collapse the parent and select it (clears Peek, as any selection change does). Descendant expansion state is kept.
- Top-level entry, or Recent/Starred mode: no-op.
- Right arrow unchanged.

## View toggles (Files Nav and Picker, same keys)

- `alt+h`: hidden files (dotfiles and anything under a dot-directory). One session-wide flag shared by Files Nav and Picker. Default: shown.
- `alt+i`: gitignored files. One session-wide flag shared by both. Default: hidden. When shown, the Tree stops applying `.gitignore` and ripgrep runs with `--no-ignore-vcs`. A fixed set of heavy folder names stays excluded while shown (declared once in `internal/tree`): `node_modules`, `.venv`, `venv`, `__pycache__`, `.mypy_cache`, `.pytest_cache`, `.tox`, `.gradle`, `.next`, `target`. `.git` is always excluded. Revealed gitignored entries are drawn muted in the Tree.
- `alt+m`: File Filter (Markdown | All), still per Root (D140). Replaces Files Nav `a` and Picker `alt+a`.
- Flags bound the Tree only; in Recent/Starred or with the Wikipedia Source active, toggling still flips the flag and the Footer says it has no effect there (same shape as today's File Filter).
- A toggle that hides the selected Tree row moves the selection to the nearest visible ancestor (refresh's rule). An open document stays open.
- Indicator: right-aligned on the Files Nav status row (Tree mode only), plain text in the row's style: `h` or a space, `i` or a space, then ` md` or `txt`, single-space separated (default `h    md`, all on `h i txt`). Replaces the ` · Markdown`/` · All` suffix. Under width pressure the indicator drops first, then the mode/scope label truncates.
- The Picker title uses the same indicator, showing only the parts that apply: none in Wikipedia mode, no filter part in Starred scope. Replaces ` · Markdown` / ` · Hidden`.

## Docs

Update `docs/keybindings.md`, the help modal, footer hints, `keymap.go` comments, tests and PTY checks.
