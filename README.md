# noggin

A keyboard-driven terminal UI for reading markdown. noggin keeps several note
sources — personal notes, shared work notes, the docs of whatever project
you're in, and Wikipedia — open and reachable in one session. It is a reader,
not an editor: when you need to write, it hands off to `$EDITOR`.

## Features

- **Multiple sources in one session** — star the folders you care about
  (**roots**) and switch between them, with Wikipedia always built in.
- **File explorer** — browse a folder tree, or jump to your **Recent** and
  **Starred** documents. Git status (`M`/`A`) is shown beside changed files.
- **Well-rendered markdown** — headings, tables, syntax-highlighted code, and
  YAML frontmatter shown as a card. Toggle to raw text with line numbers.
- **Fuzzy find and grep** — a picker for file names, file contents and
  Wikipedia article titles, with a live preview.
- **Find in document** — incremental search with next/previous match.
- **Two side-by-side viewers**, each with its own back/forward history.
- **Wikipedia reader** — search and read English Wikipedia articles converted to
  markdown, and star them like any other document.
- **Themes** — four built-in dark themes (`gruvbox-material`, `tokyonight`,
  `everforest`, `charm`); switch with `ctrl+t`.
- **Edit in your editor** — press `e` to open the current file in `$EDITOR`
  (or `vi`).

## Installation

There's no published binary yet; build from source. You need
[Go](https://go.dev/dl/) (version in `go.mod`) and `make`.

```sh
git clone <this repo>
cd noggin
make build      # builds bin/noggin
make install    # copies it to ~/.local/bin (override with PREFIX=/some/path)
```

Make sure the install directory is on your `PATH`. To run without installing:

```sh
make run
```

`go install` is not supported yet.

## Getting started

```sh
noggin            # open your starred roots
noggin <dir>      # open a folder as an ad-hoc root for this session
noggin <file>     # open a file directly; the current directory becomes a root
```

The first time, run `noggin <dir>` on a folder of notes, then press `ctrl+s`
in the Sources pane to star it — a plain `noggin` will open it from then on.

A few keys to get oriented (press `?` at any time for the full list for the
pane you're in):

| Key | Action |
|---|---|
| `r` / `f` / `d` / `D` | focus Sources, Files, first viewer, second viewer |
| `tab` / `shift+tab` | cycle focus between panes |
| `↑` `↓` `←` `→` | move; expand / collapse folders |
| `enter` | open the selected file (`ctrl+enter` opens it in the second viewer) |
| `space` | open the picker: file names, grep, or Wikipedia |
| `/`, `n`, `N` | find in the document, next / previous match |
| `ctrl+s` | star / unstar a root, document or article |
| `e` | edit in `$EDITOR` |
| `m` | toggle markdown / raw text |
| `R` | refresh from disk |
| `q` | quit |

noggin never watches the filesystem; if something changes on disk it will say
so and you press `R` to refresh.

More detail lives in [docs/README.md](docs/README.md) and
[docs/keybindings.md](docs/keybindings.md).

## Dependencies

- **Go** — to build (see `go.mod` for the required version).
- **A terminal with 24-bit colour** — 256-colour terminals work with
  approximated colours.
- **[ripgrep](https://github.com/BurntSushi/ripgrep) (`rg`)** — required for the
  picker's file and grep search. Without it noggin still launches and reads
  documents; only those two picker modes are unavailable.
- **`git`** *(optional)* — for `M`/`A` status markers. Without it no markers are
  shown.
- **`$EDITOR`** *(optional)* — used for `e`; falls back to `vi`.
- **Network access** *(optional)* — only for Wikipedia.

## Wikipedia attribution

Wikipedia article text is © Wikipedia contributors and licensed under
[CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/). noggin fetches
articles from English Wikipedia and reformats the HTML to markdown for display;
it doesn't store article content on disk, only each article's title and URL
(for stars and Recent). Every open article shows its licence, canonical URL and
a "Reformatted to markdown" note on its location line at the bottom of the
viewer.
