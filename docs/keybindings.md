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
| `ctrl+n` | toggle sidebar | toggle sidebar | toggle sidebar | - | - | — | — |
| `r` | focus Sources [^focus] | focus Sources [^focus] | focus Sources [^focus] | - | - | — | - |
| `f` | focus Files [^focus] | focus Files [^focus] | focus Files [^focus] | - | - | — | - |
| `o` | focus DV1 | focus DV1 | focus DV1 | - | - | — | - |
| `ctrl+o` | focus DV2 [^dv2] | focus DV2 [^dv2] | focus DV2 [^dv2] | — | — | — | — |
| `ctrl+x` | close DV2 [^dv2] | close DV2 [^dv2] | close DV2 [^dv2] | — | — | — | — |
| `R` | refresh [^refresh] | refresh [^refresh] | refresh [^refresh] | - | - | — | - |
| `?` | open help | open help | open help | - | - | close help | - |
| `space` | open Picker | open Picker | open Picker | - | - | — | - |
| `ctrl+space` | reopen Picker to last query | reopen Picker to last query | reopen Picker to last query | - | - | — | - |
| `e` | edit [^edit] | edit [^edit] | edit [^edit] | - | - | — | - |
| `m` | toggle Markdown/Text [^mode] | toggle Markdown/Text [^mode] | toggle Markdown/Text [^mode] | - | - | — | - |
| `/` | find [^find] | find [^find] | find [^find] | - | - | — | - |
| `n` / `N` | next / prev match [^find] | next / prev match [^find] | next / prev match [^find] | - | - | — | - |
| **Selection & movement** | | | | | | | |
| `↑`/`k`, `↓`/`j` | move selection | move selection; clears Peek | scroll 1 line | - (`k`,`j`) | - (`k`,`j`) | — | `↑`/`↓` move Results; `k`/`j` type |
| `→`/`l`, `←`/`h` | — | expand / collapse folder (Tree only) | — | - (`l`,`h`) | - (`l`,`h`) | — | `l`,`h` type |
| `g`/`home`, `G`/`end` | — | — | top / bottom | - (`g`,`G`) | - (`g`,`G`) | — | `g`,`G` type |
| `ctrl+u` / `ctrl+d` | — | — | ½ page up / down | text-field edit | text-field edit | — | scroll Preview ½ page |
| `pgup` / `pgdn` (`ctrl+b` / `ctrl+f`) | — | — | ½ page up / down | `ctrl+b`/`ctrl+f` text-field edit | `ctrl+b`/`ctrl+f` text-field edit | — | scroll Preview full page |
| **Activation** | | | | | | | |
| `enter` | activate Source [^src] | open into DV1 [^open] | — | commit query | confirm name | — | open Result into DV1 [^pickopen] |
| `ctrl+enter` / `ctrl+l` | — | open into DV2 [^open] | — | — | — | — | open Result into DV2 [^pickopen] |
| `p` | — | peek in place [^peek] | — | - | - | — | - |
| `ctrl+p` | — | peek into DV2 [^peek] | — | — | — | — | — |
| `[` / `]` | — | — | history back / forward [^history] | - | - | — | - |
| **Stars, names, cycling** | | | | | | | |
| `s` | star Root [^starroot] | star selected row [^star] | star this viewer's document [^star] | - | - | — | - |
| `alt+s` | — | — | — | — | — | — | star selected Result [^star] |
| `c` | rename Root [^rename] | — | — | - | - | — | - |
| `tab` | — | cycle Browse Mode [^browse] | — | — | — | — | cycle mode File → Grep → Wikipedia |
| `shift+tab` | — | cycle Browse Scope [^browse] | — | — | — | — | cycle Scope [^scope] |
| `a` | — | toggle File Filter [^filter] | — | - | - | — | - |
| `alt+h` | — | — | — | — | — | — | toggle dotfiles [^hidden] |
| **Dismissal** | | | | | | | |
| `esc` | — | — | — | cancel query | discard edit | close help | close Picker |

## Footnotes

[^focus]: Reveals the Navigator first if `ctrl+n` has hidden it. If the terminal is too narrow to show the Navigator at all, focus does not move.

[^dv2]: `ctrl+o` and `ctrl+x` are no-ops when DV2 doesn't exist. `ctrl+o` never moves focus elsewhere as a side effect. `ctrl+x` moves focus to DV1 only if DV2 held it; otherwise focus stays put. Closing DV2 drops its document, history and find state, but never touches DV1's.

[^refresh]: Re-reads the Active Root's tree and both viewers' documents (whichever exist) from disk, and recomputes git markers. A viewer showing a Wikipedia article re-fetches it instead, bypassing the session cache. Clears any Peek. A committed find may be lost. Doesn't add a history entry. Failures appear in the Footer.

[^edit]: Hands the focused viewer's document to `$EDITOR` (fallback `vi`) and re-reads it on return. Acts on DV2 only when DV2 has focus; otherwise DV1. Footer message if no document is open, or for a Wikipedia article ("cannot edit a Wikipedia article"). Silent no-op while Peeking.

[^mode]: Flips the targeted viewer's Render Mode between Markdown and Text (sticky for that viewer, session-only). Acts on DV2 only when DV2 has focus; otherwise DV1. Keeps a find (query and current match, re-run in the new mode) or, with no find, the proportional scroll position. Footer message on a Text file, which is always shown as text (also while one is Peeked).

[^find]: `/` starts an incremental search over the focused viewer's document (Footer message if none is open; silent no-op while Peeking). Once committed with `enter`, `n`/`N` step through matches (wrapping) from any pane, but only act when matches exist and the viewer isn't Peeking.

[^src]: Makes the selected Source active and moves focus to the Files Nav. A Root whose path is gone is refused with a Footer message. The Wikipedia row switches the Files Nav to `Recent`/`Starred` only (no `Tree`) and makes the Picker open in Wikipedia mode.

[^open]: Loads the selected file into the viewer (DV1 for `enter`, DV2 for `ctrl+enter`/`ctrl+l`, creating DV2 if needed) and moves focus there. Folders and an empty pane are no-ops. In `Recent`/`Starred` it opens the selected flat row: a broken starred document is refused, and a Wikipedia article is fetched asynchronously. `ctrl+l` is the fallback for terminals that send `ctrl+enter` as plain `enter`. Opening records the document in `Recent`.

[^peek]: Shows the selected document read-only in the viewer (DV1 for `p`, DV2 for `ctrl+p`, creating DV2 if needed) without making it the Active Document and **without moving focus**. A second press clears it. No-op on a folder or on the document already active there. A selection change, focus change, opening the Picker, or a refresh also clears Peek. A Peeked document can't be searched or edited. Doesn't count toward `Recent`.

[^history]: Walks that viewer's own back/forward Document History (DV1 and DV2 are independent). Restores each document's last scroll position. History is session-only.

[^starroot]: Stars or un-stars the selected Root, persisting it to the `roots` file; starred Roots are what a plain `noggin` opens next time. Refused with a Footer message on the Wikipedia row.

[^star]: Persists to the `documents` file, tracked by absolute path (or canonical URL for an article), independent of the Root's own star. In the Picker, `alt+s` is used instead because plain `s` would be typed into the Prompt.

[^rename]: Edits the Root's name inline (`enter` confirms, `esc` cancels, an empty name reverts to the folder's base name). Refused with a Footer message on the Wikipedia row and on an unstarred ad-hoc Root.

[^browse]: `tab` cycles `Tree` → `Recent` → `Starred` (with Wikipedia active, `Tree` drops out and it toggles `Recent` ↔ `Starred`). `shift+tab` cycles the current mode's Scope: `Recent` = Active Source ↔ All Sources; `Starred` = Active Source → All Starred → Orphans (no `Orphans` with Wikipedia active). No-op in `Tree`. Mode and scope are remembered per Source for the session.

[^filter]: Flips the Active Root's File Filter between `Markdown` and `All` (Markdown files plus non-binary Text files). Per Root, session-only; a Root starts on `Markdown`, except the ad-hoc Root of `noggin <file>` on a Text file, which starts on `All`. Bounds the Tree only (`Recent`/`Starred` ignore it) and is named on the Tree's status row. A selection the toggle hides moves to its nearest surviving ancestor. With Wikipedia active it does nothing but show a Footer message.

[^scope]: File and Grep cycle Active Root → All Roots → Starred (with Wikipedia as Active Source, only All Roots ↔ Starred). No-op in Wikipedia mode. The Prompt text is kept across mode and scope switches.

[^hidden]: Hides/shows dotfiles and files under dot-directories in File and Grep results at once; the title gains ` · Hidden` while on. Session-only. No effect in Wikipedia mode. `.git/` is always excluded.

[^pickopen]: Activates the selected Result into DV1 (`enter`) or DV2 (`ctrl+enter`/`ctrl+l`) and closes the Picker, without changing the Active Root or the Files Nav's expansion/selection. A Grep Result lands on the matched line (ripgrep's exact line and column when the viewer is in Text mode), with `n`/`N` working from it; a Wikipedia Result fetches the article asynchronously.

## Notes

- **Typing modes are exclusive.** While Find, Rename or the Picker's Prompt is active, every key not listed for that mode is entered as text — `q` never quits and `j` never scrolls. This is why the Picker has its own `alt+s` / `alt+h` and arrow-only movement.
- **Mouse**: wheel scrolling is also supported, aimed at the pane under the cursor rather than the focused one. It isn't a key command, so isn't tabulated.
- **Where `?` shows these**: the Help Modal lists a per-pane subset (`helpBindings` in `help.go`); the Footer shows a shorter one. Neither lists `alt+s`, which appears only in the Picker's footer hint.
