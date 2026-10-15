# noggin key commands by mode

Derived from `cmd/noggin/keymap.go` and the key dispatch in `model.go` (`Update`),
`picker.go` (`handlePickerInput`), `files.go` and `rootsnav.go`.

## Modes (columns)

Exactly one mode owns the keyboard at a time. Precedence, highest first:
**Find input → Rename input → Help modal → Picker → focused pane**.

| Mode | What it is |
| --- | --- |
| **Sources** | Sources Nav pane has Focus (Roots + the Wikipedia row) |
| **Files** | Files Nav pane has Focus (`Tree` / `Recent` / `Starred` Browse Modes) |
| **Viewer** | DV1 or DV2 has Focus; keys act on that viewer |
| **Find** | A `/` query is being typed (modal input over the focused viewer) |
| **Rename** | A Root's name is being edited inline in the Sources Nav |
| **Help** | The `?` Help Modal is open |
| **Picker** | The `space` fuzzy-finder is open (File / Grep / Wikipedia) |

Legend: **—** = does nothing in that mode · **types** = the key is entered as text
· **global** = same action wherever the key is live (see the *Global* rows).

## Key commands (rows)

| Key | Sources | Files | Viewer | Find | Rename | Help | Picker |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **Global — available in Sources, Files and Viewer; blocked in the three modal modes** | | | | | | | |
| `q` / `ctrl+c` | quit | quit | quit | - | - | — | `q` -; `ctrl+c` — |
| `F` | toggle sidebar | toggle sidebar | toggle sidebar | - | - | — | - |
| `r` | focus Sources [^focus] | focus Sources [^focus] | focus Sources [^focus] | - | - | — | - |
| `f` | focus Files [^focus] | focus Files [^focus] | focus Files [^focus] | - | - | — | - |
| `d` | focus DV1 | focus DV1 | focus DV1 | - | - | — | - |
| `D` | focus DV2 [^dv2] | focus DV2 [^dv2] | focus DV2 [^dv2] | - | - | — | - |
| `tab` / `shift+tab` | focus next / prev pane [^cycle] | focus next / prev pane [^cycle] | focus next / prev pane [^cycle] | — | — | — | — |
| `ctrl+w` | — | — | close viewer [^close] | — | — | — | — |
| `R` | refresh [^refresh] | refresh [^refresh] | refresh [^refresh] | - | - | — | - |
| `?` | open help | open help | open help | - | - | close help | - |
| `space` | open Picker | open Picker | open Picker | - | - | — | - |
| `ctrl+space` | reopen Picker to last query | reopen Picker to last query | reopen Picker to last query | - | - | — | - |
| `m` | toggle Markdown/Text [^mode] | toggle Markdown/Text [^mode] | toggle Markdown/Text [^mode] | - | - | — | - |
| `/` | find [^find] | find [^find] | find [^find] | - | - | — | - |
| `n` / `N` | next / prev match [^find] | next / prev match [^find] | next / prev match [^find] | - | - | — | - |
| **Selection & movement** | | | | | | | |
| `↑`, `↓` | move selection | move selection; clears Peek | scroll 1 line | text-field edit (`↑`,`↓` ignored) | text-field edit (`↑`,`↓` ignored) | — | move Results |
| `→`, `←` | — | expand / collapse folder (Tree only) [^left] | — | text-field edit | text-field edit | — | text-field edit |
| `g`/`home`, `G`/`end` | — | — | top / bottom | - (`g`,`G`) | - (`g`,`G`) | — | `g`,`G` type |
| `ctrl+u` / `ctrl+d` | — | — | ½ page up / down | text-field edit | text-field edit | — | scroll Preview ½ page |
| `pgup` / `pgdn` (`ctrl+b` / `ctrl+f`) | — | — | ½ page up / down | `ctrl+b`/`ctrl+f` text-field edit | `ctrl+b`/`ctrl+f` text-field edit | — | scroll Preview full page |
| **Activation** | | | | | | | |
| `enter` | activate Source [^src] | open into DV1 [^open] | — | commit query | confirm name | — | open Result into DV1 [^pickopen] |
| `ctrl+enter` / `ctrl+l` | — | open into DV2 [^open] | — | — | — | — | open Result into DV2 [^pickopen] |
| `p` | — | peek in place [^peek] | — | - | - | — | - |
| `ctrl+p` | — | peek into DV2 [^peek] | — | — | — | — | — |
| `[` / `]` | — | — | history back / forward [^history] | - | - | — | - |
| **Stars, names, editing, cycling, view toggles** | | | | | | | |
| `ctrl+s` | star Root [^starroot] | star selected row [^star] | star this viewer's document [^star] | — | — | — | star selected Result [^star] |
| `e` | rename Root [^rename] | — | edit [^edit] | - | - | — | - |
| `t` | — | cycle Browse Mode [^browse] | — | - | - | — | - |
| `a` | — | cycle Browse Scope [^browse] | — | - | - | — | - |
| `alt+t` | — | — | — | — | — | — | cycle mode File → Grep → Wikipedia |
| `alt+a` | — | — | — | — | — | — | cycle Scope [^scope] |
| `alt+h` | — | toggle hidden files [^view] | — | — | — | — | toggle hidden files [^view] |
| `alt+i` | — | toggle gitignored files [^view] | — | — | — | — | toggle gitignored files [^view] |
| `alt+m` | — | toggle File Filter [^filter] | — | — | — | — | toggle File Filter [^pickfilter] |
| **Dismissal** | | | | | | | |
| `esc` | — | — | — | cancel query | discard edit | close help | close Picker |

## Footnotes

[^focus]: Reveals the Navigator first if `F` has hidden it. If the terminal is too narrow to show the Navigator at all, focus does not move.

[^dv2]: `D` is a no-op when DV2 doesn't exist and never moves focus elsewhere as a side effect. Neither `d` nor `D` reveals a hidden sidebar.

[^cycle]: `tab` cycles focus forward Sources → Files → DV1 → DV2 → Sources; `shift+tab` reverses. Panes that aren't visible are skipped: a hidden sidebar, or a terminal too narrow to show it, drops Sources and Files; a missing DV2 is skipped. A no-op with a single visible pane. Never reveals the sidebar. Inert in Find, Rename, Help and the Picker.

[^close]: With DV2 focused: closes DV2 (dropping its document, history and find state) and focus moves to DV1. With DV1 focused and DV2 open: DV1 takes over DV2's whole state (document, history, scroll, Render Mode), DV1's previous state is discarded, and DV2 closes; focus stays on DV1. With DV1 focused and no DV2: DV1 is cleared to empty (document and history), like a fresh launch; already empty is a no-op, with no Footer message. A no-op from Sources and Files.

[^refresh]: Re-reads the Active Root's tree and both viewers' documents (whichever exist) from disk, and recomputes git markers. A viewer showing a Wikipedia article re-fetches it instead, bypassing the session cache. Clears any Peek. A committed find may be lost. Doesn't add a history entry. Failures appear in the Footer.

[^edit]: Hands the focused viewer's document to `$EDITOR` (fallback `vi`) and re-reads it on return. Acts on the focused viewer only (not reachable from Sources or Files, where `e` is Rename or does nothing). Footer message if no document is open, or for a Wikipedia article ("cannot edit a Wikipedia article"). Silent no-op while Peeking.

[^mode]: Flips the targeted viewer's Render Mode between Markdown and Text (sticky for that viewer, session-only). Acts on DV2 only when DV2 has focus; otherwise DV1. Keeps a find (query and current match, re-run in the new mode) or, with no find, the proportional scroll position. Footer message on a Text file, which is always shown as text (also while one is Peeked).

[^find]: `/` starts an incremental search over the focused viewer's document (Footer message if none is open; silent no-op while Peeking). Once committed with `enter`, `n`/`N` step through matches (wrapping) from any pane, but only act when matches exist and the viewer isn't Peeking.

[^src]: Makes the selected Source active and moves focus to the Files Nav. A Root whose path is gone is refused with a Footer message. The Wikipedia row switches the Files Nav to `Recent`/`Starred` only (no `Tree`) and makes the Picker open in Wikipedia mode.

[^open]: Loads the selected file into the viewer (DV1 for `enter`, DV2 for `ctrl+enter`/`ctrl+l`, creating DV2 if needed) and moves focus there. Folders and an empty pane are no-ops. In `Recent`/`Starred` it opens the selected flat row: a broken starred document is refused, and a Wikipedia article is fetched asynchronously. `ctrl+l` is the fallback for terminals that send `ctrl+enter` as plain `enter`. Opening records the document in `Recent`.

[^peek]: Shows the selected document read-only in the viewer (DV1 for `p`, DV2 for `ctrl+p`, creating DV2 if needed) without making it the Active Document and **without moving focus**. A second press clears it. No-op on a folder or on the document already active there. A selection change, focus change, opening the Picker, or a refresh also clears Peek. A Peeked document can't be searched or edited. Doesn't count toward `Recent`.

[^history]: Walks that viewer's own back/forward Document History (DV1 and DV2 are independent). Restores each document's last scroll position. History is session-only.

[^starroot]: Stars or un-stars the selected Root, persisting it to the `roots` file; starred Roots are what a plain `noggin` opens next time. Refused with a Footer message on the Wikipedia row.

[^star]: Persists to the `documents` file, tracked by absolute path (or canonical URL for an article), independent of the Root's own star. `ctrl+s` is the same chord in the Picker as everywhere else; as a control chord, it is never typed into the Prompt.

[^rename]: `e` on the Sources Nav edits the Root's name inline (`enter` confirms, `esc` cancels, an empty name reverts to the folder's base name). Refused with a Footer message on the Wikipedia row and on an unstarred ad-hoc Root.

[^browse]: `t` cycles Browse Mode `Tree` → `Recent` → `Starred` (with Wikipedia active, `Tree` drops out and it toggles `Recent` ↔ `Starred`). `a` cycles the current mode's Scope: `Recent` = Active Source ↔ All Sources; `Starred` = Active Source → All Starred → Orphans (no `Orphans` with Wikipedia active). No-op in `Tree`. Mode and scope are remembered per Source for the session.

[^left]: `←` on an expanded folder collapses it. On a file or collapsed folder inside a folder it collapses the parent and selects it (clears Peek; descendants keep their expansion state). A no-op on a top-level entry and in `Recent`/`Starred`. `→` expands the selected folder.

[^filter]: Flips the Active Root's File Filter between `Markdown` and `All` (Markdown files plus non-binary Text files). Per Root, session-only; a Root starts on `Markdown`, except the ad-hoc Root of `noggin <file>` on a Text file, which starts on `All`. Bounds the Tree only (`Recent`/`Starred` ignore it) and shows as ` md` / `txt` in the Tree status row's indicator. A selection the toggle hides moves to its nearest surviving ancestor. With Wikipedia active it does nothing but show a Footer message.

[^pickfilter]: Flips the Active Root's File Filter between `Markdown` and `All`, the same per-Root state as the Files Nav's `alt+m` (the Tree is re-walked too). File and Grep's Active Root and All Roots scopes use the Active Root's filter for every Root searched; `All` means Markdown plus non-binary files. The `Starred` scope and Wikipedia mode ignore it. With Wikipedia as the Active Source it toggles the most recently set value (initially `Markdown`). The title's indicator shows it (` md` / `txt`), e.g. `Grep · All Roots` with `txt` at the right.

[^scope]: `alt+a` cycles. File and Grep cycle Active Root → All Roots → Starred (with Wikipedia as Active Source, only All Roots ↔ Starred). No-op in Wikipedia mode. The Prompt text is kept across mode and scope switches.

[^view]: Same chords in the Files Nav and the Picker, and one session-wide flag each, shared by both (reset only on relaunch). Toggling in either re-walks the Tree / re-searches the Picker.
    - `alt+h` hides/shows hidden files (dotfiles and anything under a dot-directory). Default: shown.
    - `alt+i` shows/hides gitignored files. Default: hidden. While shown, the Tree stops applying `.gitignore` (revealed entries are drawn muted) and ripgrep runs with `--no-ignore-vcs`. These folder names stay excluded while shown: `node_modules`, `.venv`, `venv`, `__pycache__`, `.mypy_cache`, `.pytest_cache`, `.tox`, `.gradle`, `.next`, `target`.
    - `.git/` is always excluded.
    - In `Recent`/`Starred`, or with Wikipedia active, the Files Nav toggles still flip the flag and the Footer says they have no effect there. A selection a toggle hides moves to its nearest visible ancestor.
    - Indicator: right-aligned on the Files Nav status row (Tree only) and on the Picker's Prompt title: `h` or a space, `i` or a space, then ` md` or `txt`, single-space separated. Default is `h    md`; everything on is `h i txt`. Under width pressure the indicator drops first, then the label truncates. The Picker shows only the applicable slots: none in Wikipedia mode, only `h` in the `Starred` scope.

[^pickopen]: Activates the selected Result into DV1 (`enter`) or DV2 (`ctrl+enter`/`ctrl+l`) and closes the Picker, without changing the Active Root or the Files Nav's expansion/selection. A Grep Result lands on the matched line (ripgrep's exact line and column when the viewer is in Text mode), with `n`/`N` working from it; a Wikipedia Result fetches the article asynchronously.

## Notes

- **Typing modes are exclusive.** While Find, Rename or the Picker's Prompt is active, every key not listed for that mode is entered as text — `q` never quits and `g` never jumps to the top. This is why the Picker's own bindings are all control or `alt+` chords (`ctrl+s`, `alt+t`, `alt+a`, `alt+h`, `alt+i`, `alt+m`) and its movement is arrow-only.
- **Mouse**: wheel scrolling is also supported, aimed at the pane under the cursor rather than the focused one. It isn't a key command, so isn't tabulated.
- **Where `?` shows these**: the Help Modal lists a per-pane subset (`helpBindings` in `help.go`); the Footer shows a shorter one. The Help Modal lists `tab`/`shift+tab` and `alt+i` in every pane, and Edit only in the viewer; the Footer's Sources and Files hints no longer include Edit, and the Picker's footer lists `alt+i` after star.
