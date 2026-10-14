# Rendering YAML/TOML frontmatter: options survey

As of: pre-C7 (2026-10-14)

Question: instead of hiding frontmatter (D5), what are the ways noggin could *display* it (or
use its values), what does each cost, which parser library would back it, and how does each
interact with Find, the Grep hand-off (D34) and tags (D10)? Research only — no repo code was
changed. Behaviour marked **[ran]** was executed in throwaway programs under `/tmp/fmres`
(copies of `theme.go`/`theme_markdown.go` + glamour v2.0.1 + goldmark v1.8.2); **[inferred]**
means reasoned from source, not executed. Supersedes the "revisit" pointer in
`knowledge/frontmatter-rendering.md` (see Q1g: its library path is stale).

## Summary

Recommendation-neutral. "Deps" = new `go.mod` requirements. "Find" = does Find / the D34 Grep
landing (`picker.go:1060`) see the frontmatter text. "Tags" = how much it shortens a later
tags implementation (D10).

| # | Option | Effort [inferred] | Deps | Visual result | Find / Grep landing | Tags synergy |
|---|---|---|---|---|---|---|
| a | Strip (status quo, D5) | 0 | none | nothing shown | rg line numbers offset; in-frontmatter hits degrade (`picker.go:1060-1075`) | none; tags need their own parser |
| b | Prepend ```` ```yaml ```` fence | S: splitter fix + ~10 lines; drop `stripFrontmatter` call in `resolveGrepLanding` | none | Chroma-highlighted raw YAML in Theme colours; long lines wrap, continuation loses the block indent [ran] | **Best**: text is verbatim, so occurrence ordinals line up with the raw file | none (just display) |
| c | Rewrite to GFM table / definition list | M: YAML parse + escaping (+~50 lines) | yaml lib | Table spans full width (Glamour 2.0.1); DefList = term line + `🠶 value` line [ran] | Values searchable, but quotes/brackets/escapes differ from raw → ordinals can drift | Parsed values on hand |
| d | lipgloss card, concatenated *into* rendered content | M–L: parse + card + muted variant + tests (+~100 lines) | yaml lib | Bordered key/value box in Theme colours, width-exact [ran] | Searchable *if* inside `vp.SetContent`; invisible if a separate widget | Parsed values on hand |
| e | Collapsible (toggle key) | +S–M on top of b/c/d | none | Hidden by default, key shows | Toggle forces `render()` → drops Find state like a resize | none |
| f | Use values (`title` …) in chrome | S on top of a parse | yaml lib | Pane title / location line change | none | Parse-at-load is shared with tags |
| g | goldmark frontmatter ext + `ansi.NewRenderer` | M: ~15 lines plumbing [ran] + 2 deps + a guard for the ext's failure modes | `go.abhg.dev/goldmark/frontmatter` (+ toml, yaml.v3) | Same as b–d; only gain is data returned from the same parse | n/a (display chosen separately) | Poor: parses whole markdown (~600× slower than split+yaml [ran]) |

Observation across rows: **b–f do not need goldmark/Glamour changes.** Every call site funnels
through `renderMarkdownStyled` (`document.go:181`, called from `viewer.go:471,488,509,523`), so
one splitter before `RenderBytes` (`document.go:203`) plus a render of the pieces is enough.

## Q1 — Display options

### 1a. Keep stripping (status quo)

- `stripFrontmatter` (`cmd/noggin/document.go:93-111`) is a regex `(?m)^---[ \t]*\r?\n` matched
  twice; first match must be at offset 0. No parsing. Tests: `viewer_test.go:51`, `:97`,
  `picker_test.go:722`, `test/pty/checks.py:227` (all assert frontmatter is *absent*).
- Un-stripped, Glamour does render it wrong: opening `---` → rule, YAML → paragraph, closing
  `---` → setext H2 **[ran]** (SGR `1;4;38;2;234;105;98` = bold+underline in `Markdown.H2`
  `#ea6962`, `theme.go`), and inline markdown inside is interpreted (`*c*` lost its asterisks).
  The prior finding's "spurious H2" claim is confirmed (was "not executed").
- Same wrongness on GitHub's own REST renderer: `POST https://api.github.com/markdown` with
  frontmatter returned `<hr>` + `<h2>title: Hi tags: [a, b]</h2>` **[ran]**.

### 1b. Fenced ```yaml block

- Prepending ```` ```yaml\n<fm>``` ```` renders verbatim (asterisks, pipes, quotes preserved)
  **[ran]**. Glamour hands the fence to Chroma (`glamour@v2.0.1/ansi/codeblock.go:80-110`,
  `quick.Highlight`); `elements.go:333-346` builds it from the node's lines + language.
- **Theme colours apply.** `cfg.CodeBlock.Chroma = chromaStyle(t)` (`theme_markdown.go:90-95`)
  maps YAML keys (`NameTag`, `theme_markdown.go:123`) to `Syntax.Tag` and punctuation/strings
  likewise. Observed SGR: keys `38;2;239;119;112` (= `#ef7770` Syntax.Tag), `:` `168;153;132`
  (Punctuation), strings `169;182;101` (Syntax.String) **[ran]**.
- Wrapping: a 160-char `description:` at width 50 wraps (max line width 48, no overflow), but
  continuation lines lose the code block's 2-col margin, so the wrap looks ragged **[ran]**.
- **Muted pane gotcha (pre-existing, affects b only).** Chroma style "charm" is registered once
  per process and never replaced (`ansi/codeblock.go:84-86`, comment "Don't register the style
  if it's already registered"). Rendering focused-then-muted left the muted block's token colours
  at the *focused* values (`239;119;112`, `168;153;132`, `169;182;101` still present); muted-first
  produced muted colours everywhere **[ran]**. `render()` renders focused first
  (`viewer.go:509` then `:523`), so an unfocused pane's *existing* code blocks are probably
  already un-muted; option b would add one more such block. Option c/d avoid it (no Chroma).
- Cost: needs the same robust splitter as any option that shows frontmatter (see Q4); no new dep.

### 1c. Key/value table or definition list via markdown rewrite

- Glamour enables GFM (tables) and `extension.DefinitionList` (`glamour@v2.0.1/glamour.go:70-74`);
  both have ansi renderers (`ansi/renderer.go:65-73`).
- **Table [ran]:** `| Key | Value |` header + one row per key renders as a full-width two-column
  table with `│`/`─` rules; wraps long values inside the cell at narrow widths (30 cols OK).
  GFM tables need a header row, so a literal "Key | Value" header is unavoidable (or empty cells).
  **Escaping is mandatory:** an unescaped `a|b *c*` value rendered as just `a` — silent data loss
  **[ran]**; `\|` and `\*` escapes preserved it. `[[wiki]]` links are untouched in v2.0.1.
- **DefinitionList [ran]:** `term\n: value` renders `term` then `🠶 value` on the next line (noggin's
  `DefinitionDescription.BlockPrefix` is `"\n🠶 "`, `theme_markdown.go:87`; element behaviour at
  `ansi/elements.go:433-462`). Compact, no escaping problems for keys, but values still go
  through inline-markdown parsing. There is no distinct term/description colour in noggin's
  stylesheet (both default to the body colour; `DefinitionTerm` unset).
- A bold-key bullet list also works **[ran]** and is the least fragile; it has no header row.
- Lists/maps need flattening (`tags: [a, b]` → `a, b`); nested maps need a policy. All
  decisions are noggin's, since the YAML must be parsed first.

### 1d. Custom lipgloss "metadata card" outside Glamour

- Built in the experiment from yaml.v3 `Node` (order-preserving), `lipgloss.Border(Rounded)`,
  `Theme.Border`, key style `Markdown.H5`+bold, value style `Theme.Fg`, `MarginLeft(2)`, outer
  `Width(w-2)`: at widths 60 and 30 the max rendered line width equalled the requested width
  and long values wrapped inside the box **[ran]**. It aligns with Glamour's 2-col
  document margin (`theme_markdown.go:36-40`).
- Concatenate `card + "\n" + glamourOutput` and pass to the single `vp.SetContent`
  (`viewer.go:522`): the card is then just rendered lines, so Find (`findMatches` strips ANSI
  and counts newlines, `find.go:186-190`) and `StyleLineFunc` reverse-video (`find.go:148-160`)
  work with **no change**. A *separate* widget above the viewport would be invisible to Find and
  would need a row budget in `docContentHeight` (`location.go:97`) — avoid.
- **Muted pane:** `renderMarkdownStyled(..., muted)` is called twice per render
  (`viewer.go:509,523`) and Peek (`:471,488`). The card must take the same flag and use
  `Theme.UnfocusedText/UnfocusedSubtle` (`unfocused_style.go:32-48` shows the existing
  `mutedText` helper). The Glamour stylesheet is muted by deriving it (`mutedMarkdownStyleConfig`,
  `theme_markdown.go:150`); the card is lipgloss-built so it needs its own muted style — no
  ANSI post-processing (`unfocused_style.go` header rules it out).
- No-TTY: `markdownColoured()` (`document.go:122`) gates colour for `go test`; the card would
  need the same gate or `fixtureStructure`-style plain assertions (`viewer_test.go`) will see SGR.
  [inferred]
- Best visual control, most code. Looks like Obsidian's "Properties" panel (see Q3).

### 1e. Collapsible / toggle

- Orthogonal modifier on b/c/d. State: a `showMeta bool` on `viewer` (per-viewer, like Find),
  default false; `document` could carry parsed frontmatter so toggling needs no re-parse.
- Toggling must re-run `render()` (`viewer.go:500`), which resets `findState` and
  `StyleLineFunc` (`:503-504`) and re-sets scroll content — i.e. a Committed Find is dropped,
  and scroll offset is affected as lines are added/removed above the viewport top. [inferred]
- Keymap: all bindings live in `keymap.go` (D2). Letter keys already used: `c e f n o p q r s N R`
  `/ ? [ ]` plus nav keys (grep of `WithKeys`, `keymap.go`); `m`, `i`, `y` are unbound; `t`
  is reserved for Tags in NOTES (`.charm/NOTES.md:252`). Needs a `Keymap` field, Help Modal
  entry (`help.go`), footer hint, and a per-viewer test (DV1/DV2 independence per GLOSSARY
  "Document Viewer"). Not persisted across sessions unless added to a store.
- Obsidian precedent: three display modes, Visible/Hidden/Source (Q3).

### 1f. Use frontmatter values to influence the viewer

- Pane title is `p.title() + ": " + [★ ] + v.label()` (`model.go:2170-2190`), `label()` is
  `filepath.Base(path)` for files and `d.title` for articles (`document.go:59-67`). So there is
  precedent for a document-supplied title: articles already use one. A frontmatter `title` could
  populate the same slot for files — but then the pane shows a different name than the Files Nav
  row and Picker row (which show file names), and `document.title` today means "article's own
  display title" (`document.go:20-29`). Open question for the user.
- Other candidate hooks: Location line (`location.go:32`), a tag count/badge, `draft: true` badge,
  `aliases` for the File Picker. Each needs the parsed value on `document` at load
  (`loadDocument`, `document.go:70`) rather than at render time, so Refresh and the `$EDITOR`
  handoff re-parse naturally (they already re-read bytes).
- Obsidian assigns meaning to `tags`, `aliases`, `cssclasses` as defaults
  (`obsidianmd/obsidian-help en/Editing and formatting/Properties.md:29`).

### 1g. goldmark frontmatter extension + `ansi.NewRenderer`

- **Re-check of the prior claim: still true in v2.0.1 *and* glamour `main`.** `TermRenderer.md`
  is unexported (`glamour@v2.0.1/glamour.go:35`); `NewTermRenderer` hard-wires GFM,
  DefinitionList, `WithAutoHeadingID` (`:70-77`) then installs the ansi renderer (`:88-95`).
  The only option that extends goldmark is in-package `WithEmoji` (`:208-213`,
  `emoji.New().Extend(tr.md)`). No hook. glamour `main` (f71347d, 2026-10-08) is identical in
  these lines; its only new option is `WithHyperlinkMode`. `go list -m -versions` shows v2.0.1 as
  latest. Issue #205 "Frontmatter support" is still open, last activity 2022-11-23
  (https://github.com/charmbracelet/glamour/issues/205).
- **What bypassing costs [ran]:** `ansi.Options` is public (`ansi/renderer.go:17-25`) and
  `ansi.NewRenderer` is public (`:33`). The re-implementation is: `goldmark.New(GFM,
  DefinitionList, frontmatter ext, WithAutoHeadingID)`, `ansi.NewRenderer(ansi.Options{WordWrap,
  Styles, ChromaFormatter})`, `md.SetRenderer(renderer.NewRenderer(WithNodeRenderers(
  util.Prioritized(ar, 1000))))`, then `md.Convert(src, w, parser.WithContext(ctx))` and
  `frontmatter.Get(ctx).Decode(&node)` — about 15 lines, and the rendered body was identical
  minus frontmatter. noggin only uses `WithStyles`, `WithChromaFormatter`, `WithWordWrap`
  (`document.go:195-198`), each of which just sets one `ansi.Options` field
  (`glamour.go:146-152,173-179,226-231`), so nothing else needs re-creating. Side effect: the
  default word-wrap of 80 (`glamour.go:25,80`) that applies when noggin passes width ≤ 0
  (contradicting the "renders unwrapped" comment at `document.go:173`) must be copied or changed.
- **Gain is small.** The extension returns *data*; display still has to be chosen (b/c/d). You
  can get the same data from a hand-written splitter + yaml decode with no goldmark involvement.
  The only thing a goldmark-level approach adds is putting a *node* in the AST for
  a custom renderer — not needed for b/c/d.
- **Library traps found [ran]:** (1) `goldmark-meta`'s `WithTable()` AST (it synthesises
  `ast.String` cells) **panics** in Glamour v2.0.1: `Warning: unhandled element String` then nil
  deref at `ansi/table.go:186`. (2) Glamour prints that warning with `fmt.Println`
  (`ansi/elements.go:478`) — stdout, inside a Bubble Tea program, would corrupt the frame; any
  unknown node kind from an extension does this. (3) `goldmark-frontmatter` v0.3.0's failure
  modes are in Q4 (unclosed block swallows the document; false positive on line 2).

## Q2 — Parsing libraries

Versions and dates from `go list -m -json <mod>@latest` and the GitHub API (2026-10-14).

| Library | Latest / date | Formats | Key order | Deps (non-std pkgs in a hello binary) [ran] | License | Status |
|---|---|---|---|---|---|---|
| `go.abhg.dev/goldmark/frontmatter` (**formerly** `github.com/abhinav/goldmark-frontmatter`) | v0.3.0, 2025-11-17; repo pushed 2026-10-12 | YAML `-`, TOML `+`, custom `Format{Delim,Unmarshal}`; ≥3 delimiter chars, closer must match count (`parse.go:177-193`, `README.md`) | Via `Decode(&yaml.Node)` yes; `map` no [ran] | 11 (BurntSushi/toml, yaml.v3, goldmark…) | BSD-3 | Active, 33★ |
| `github.com/yuin/goldmark-meta` | v1.1.0, 2022-02-12; pushed 2026-08-31 | YAML only (`meta.go:99-103` `isSeparator`) | **Yes**: `GetItems` → `yaml.MapSlice` (`meta.go:61`) | 10 (yaml.v2) | MIT | Stable/quiet |
| `github.com/adrg/frontmatter` | v0.2.0, 2020-11-04; pushed 2026-07-30 | YAML `---`, `---yaml`, TOML `+++`/`---toml`, JSON `;;;`/`{}` (`format.go:57-68`) | Via yaml.v2 `MapSlice` if you pass one | 3 (yaml.v2, toml v0.3.1) | MIT | No release in 6 yr; not goldmark-based; returns the rest of the bytes |
| `gopkg.in/yaml.v3` | v3.0.1, 2022-05-27 | YAML | **Yes**: `yaml.Node` keeps document order (`yaml.go:344-346`) | 1 | MIT + Apache-2.0 | **Archived/unmaintained** (README banner; repo archived 2025-04-01) |
| `go.yaml.in/yaml/v3` | v3.0.5, 2026-07-26 | YAML (yaml.v3 API) | Yes, `Node` | ~1 | same | Maintained continuation under the YAML org (https://github.com/yaml/go-yaml, pushed 2026-10-13) |
| `github.com/goccy/go-yaml` | v1.19.2, 2026-01-08 | YAML | Yes: `yaml.MapSlice` + `UseOrderedMap()` (`yaml.go:70-77,243`) | 9 | MIT | Active |
| `github.com/BurntSushi/toml` | v1.6.0, 2025-12-18 | TOML | Maps unordered; `MetaData.Keys()` gives order | 2 | MIT | Active |

Notes:
- **Prior finding's import path is stale.** `github.com/abhinav/goldmark-frontmatter`'s `go.mod`
  declares `module go.abhg.dev/goldmark/frontmatter`; the README installs
  `go get go.abhg.dev/goldmark/frontmatter@latest`. (A `go get` by the old path was not tried.)
  Update `knowledge/frontmatter-rendering.md` if this option is ever pursued.
- **Ordering matters for display** and rules out `map[string]any` decoding. Verified: decoding
  into `map` printed three different key orders over three iterations; `yaml.Node` returned
  `title tags author draft link note` (document order) **[ran]**. For TOML via BurntSushi, order
  needs `MetaData.Keys()`; an ordered TOML display is harder than YAML.
- **Weight:** the repo already has goldmark (indirect, `go.mod`) and Chroma. yaml.v3 alone adds
  one package and ~0.6 MB to a hello-world binary; the goldmark-frontmatter route adds ~1.1 MB
  (toml + yaml) **[ran, crude: `-s -w` hello binary, arm64]**. YAML parsing of a frontmatter block
  of a few lines is microseconds (Q5).
- **Hand-rolled flat `key: value` parser (no dep) [inferred]:** workable for display of simple
  scalars, ~30 lines. Breaks on: block scalars (`|`, `>`), multi-line flow lists, block lists
  (`tags:\n  - a`, the form Obsidian documents as *required* for `tags`,
  `Tags.md:12`), nested maps, quoted keys/values with `:` or `#`, anchors, comments. Tags
  alone need flow + block lists, so a hand-rolled parser would likely grow into half a YAML
  parser. A pragmatic split: hand-roll the **splitter** (fences, BOM, CRLF — Q4) and let a
  YAML lib parse the block.
- **TOML**: only goldmark-frontmatter and adrg handle `+++`; with a splitter you can pick the
  decoder by fence. Hugo/Hugo-style notes are the only realistic source of TOML in a notes
  vault; Obsidian is YAML-only (JSON is accepted and re-saved as YAML, `Properties.md:259-273`).

## Q3 — Prior art

| Tool | Behaviour | Source |
|---|---|---|
| Glow | Strips: `utils.RemoveFrontmatter`, regex `(?m)^---\r?\n(\s*\r?\n)?` twice; called at `main.go:280`, `ui/ui.go:177,272`. No title extraction. | `glow@8b0600f` (2026-10-02) `utils/utils.go:19-35` |
| Frogmouth (Textual) | "Front matter is now ignored when viewing a file" — markdown-it-py `front_matter_plugin`, so parsed *and dropped* | `frogmouth/ChangeLog.md` "Unreleased"; `frogmouth/widgets/viewer.py:12,150` |
| mdcat | No handling: `Options` = TASKLISTS\|STRIKETHROUGH\|TABLES only (no metadata-block flag although pulldown-cmark offers `ENABLE_YAML_STYLE_METADATA_BLOCKS`), so it renders as rule + setext heading (inferred) | `mdcat/src/lib.rs:82-85`; pulldown-cmark `src/lib.rs:742` |
| Obsidian | Visible (default): a *Properties* panel of typed rows at the top of the note; Hidden; Source (plain YAML). Typed: text, list, number, checkbox, date, date&time, tags. Nested properties and markdown in values **not supported**; name/value order is irrelevant to Obsidian. | `obsidianmd/obsidian-help` `en/Editing and formatting/Properties.md:27,65-83,146` |
| Obsidian tags | `tags` property must be a YAML **list**; also inline `#tag`; nested via `/` (`#inbox/to-read`) | `.../Tags.md:12-38` |
| Hugo | Formats by delimiter: YAML `---`, TOML `+++`, JSON `{…}`, plus org-mode `#+`; leading ASCII whitespace/blank lines skipped before the opener; not displayed (metadata for templates) | `hugoDocs content/en/content-management/front-matter.md:17-20`; `hugo parser/pageparser/pagelexer_intro.go:16-37,123-157` |
| Jekyll | "must be the first thing in the file", valid YAML between triple-dashed lines; empty block still counts; **warns that a UTF-8 BOM breaks it**; not displayed | `jekyll/jekyll docs/_docs/front-matter.md:5-9,26-30,38-42` |
| GitHub | github.com's blob view is *widely documented* (2016 blog post) as rendering frontmatter as a table. **Not verified**: github.blog needs sandbox approval, and `docs.github.com` search for front matter returned only Pages/Docs-authoring pages, none describing the rendering. Executed evidence: the `/markdown` REST renderer does *not* do it (hr + H2, Q1a). | docs.github.com search API (queries "yaml front matter", "frontmatter markdown preview"); `POST api.github.com/markdown` [ran] |

Pattern: terminal viewers (Glow, Frogmouth, mdcat-by-accident) hide it; GUI editors
(Obsidian, GitHub) show it as a structured panel/table; none of the terminal ones render it.
Hugo's "leading whitespace skipped" and Obsidian's "JSON also accepted" are the two permissive
definitions; Jekyll's is the strict one (first thing in file, no BOM).

## Q4 — Spec edge cases

There is no standard; definitions differ (Q3). Table: results of running the **current regex**
(`document.go:93,106`) and `goldmark-frontmatter` v0.3.0 (`&frontmatter.Extender{}` +
`ansi.NewRenderer`) over a corpus **[ran]**. "ok" = block hidden/consumed and body kept; "leaks"
= YAML shown as body text; "eats body" = document content lost.

| Input | Current regex | goldmark-frontmatter ext |
|---|---|---|
| LF / CRLF normal | ok / ok | ok / ok |
| Empty block `---\n---` | ok | ok |
| UTF-8 BOM + normal | **leaks** (`^---` matches at offset 3, `fences[0][0] != 0`) | **leaks** (no BOM handling; goldmark has none) |
| `...` closing fence (YAML doc-end marker) | leaks (not recognised) | **eats body** (`...` isn't a closer → block runs to EOF) |
| TOML `+++` | leaks | ok |
| Bare JSON `{…}` (Hugo) | leaks | leaks |
| JSON inside `---` fences (Obsidian) | ok (stripped) | ok |
| Trailing space on fence (`--- `) | ok (`[ \t]*`) | leaks (exact delimiter run) |
| 4+ dash fences | leaks (needs exactly 3) | ok (≥3, count must match) |
| Blank line before opener | leaks | ok (see next row) |
| Closer at EOF with no `\n` | leaks (regex needs `\n`) | ok |
| Doc starts with `---` rule, no 2nd `---` | ok (not stripped) | **eats body** |
| Doc starts with `---` rule, later `---` rule | strips up to the later rule | strips up to the later rule |
| `# T` then `---\nfoo: bar\n---` on line 2 | ok (not stripped) | **false positive**: `foo: bar` consumed (off-by-one `lineno > 1`, `parse.go:59`) |
| `text` then `---` (setext) | ok | ok (`CanInterruptParagraph` false, `parse.go:125`) |

Additional definitional points:
- Obsidian: tags/properties in YAML only; JSON read then re-saved as YAML; bare JSON not listed
  (`Properties.md:259-273`). Hugo accepts bare `{…}` JSON and `#+` org-mode (pagelexer_intro.go
  above). Jekyll: YAML only, BOM fatal.
- YAML itself permits `...` as document end; none of Jekyll/Hugo/Obsidian docs mention it
  as a frontmatter closer (searched the three sources above; Hugo's lexer only looks for
  its own opener repeated, `pagelexer_intro.go:146`). Rare in practice; whether to support is a
  user decision.
- Whatever noggin does, a hardened splitter (strip BOM, accept `---`/`+++`, optional trailing
  whitespace, CRLF, require a closer else treat as no frontmatter, closer at EOF) is needed
  by every option except (a) — and (a) would also benefit from the BOM and `+++` fixes, since
  today those leak YAML into the render.

## Q5 — Interactions

**Find / Grep hand-off (D34, `.charm/DECISIONS.md:~380-400`).**
- Today `resolveGrepLanding` (`picker.go:1060-1075`) re-strips, subtracts the stripped line
  count, and returns `ok=false` for hits inside frontmatter, degrading to "first body match".
  Test: `picker_test.go:722-770`.
- If frontmatter is shown **verbatim (b)**, the function simplifies: remove the strip step;
  every raw occurrence has a rendered counterpart (ordinal mapping stays exact). Wrapping in
  Glamour doesn't change ordinals since Find matches against ANSI-stripped text
  (`find.go:186-190`).
- **Reformatted (c/d)**: raw `author: "Lucas, George"` becomes `author   Lucas, George`; a query
  containing the quotes or `key:` finds nothing in rendered space, and a query that matches
  text only in the rendered (e.g. a `Key` header, or a card border) shifts ordinals by one for
  every later raw match. D34's accepted degradation already tolerates this, but the in-frontmatter
  degrade branch is then *needed in a different way* (the offset is the rendered card/table,
  not zero). [inferred]
- Find over a card placed in the same viewport content works unchanged (Q1d).
- ripgrep Preview pane shows the **raw file** (D30, `picker.go:594-597`): frontmatter is always
  visible there regardless of option. No change needed.
- Toggle (e): any toggle that adds/removes lines invalidates line-addressed state, which
  `render()` already resets (`viewer.go:500-505`).

**Tags (D10, `.charm/ROADMAP.md:70-71`, `NOTES.md:333-337`).**
- Tag parsing must run over *many files* (Tag Browser = "all tags for the scope",
  `NOTES.md:126`), where parsing whole markdown through goldmark is the wrong tool: benchmark on a
  121 KB document — goldmark parse with the frontmatter ext **4.0 ms/op**; split + `yaml.Unmarshal`
  into `struct{Tags []string}` **6.7 µs/op** (~600×) **[ran, arm64 sandbox, single run]**.
  So option g is the *worst* synergy for tags; the best is: splitter + yaml decode, shared with
  b–d, fed by reading only the file's head.
- Whichever display is chosen, parsing at `loadDocument` time (not at render) gives tags and
  titles the same data.
- If tag notation stays open: Obsidian-style `tags:` list is the only frontmatter form with
  documented semantics (YAML list; nested with `/`, `Tags.md:12-38`), which conveniently
  matches NOTES' `projects/work` nesting (`NOTES.md:25`).
- The display option does not decide D10; but showing `tags` in a card (d) would be the natural
  place to later make tags selectable (a cursor, which C1/D4 refused).

## Open questions for the user

1. Is the goal *seeing* frontmatter (read-only reference), *using* it (title, tags, badges), or
   both? (Decides b–e vs f vs a shared parse.)
2. Should frontmatter be visible by default? If hidden-by-default, is a per-viewer toggle
   (e) wanted, and should it persist per session/Root like Browse Mode?
3. Raw fidelity vs. tidy look: is Chroma-highlighted verbatim YAML (b) good enough, or is a
   structured card (d) wanted despite ~2–3× the code and a muted-style variant?
4. Should a frontmatter `title` rename the pane/Location line for files, given Files Nav and
   Pickers will still show filenames? (Articles already do this.)
5. Which formats must be recognised: YAML only, or TOML `+++` and bare JSON too? Obsidian is
   the stated tag notation candidate (`NOTES.md:336`), which implies YAML only.
6. Is adding a YAML dependency acceptable? yaml.v3 is archived; `go.yaml.in/yaml/v3` is the
   maintained drop-in; `goccy/go-yaml` is the other active choice.
7. Should the splitter fixes (BOM, `+++`, `...`, 4+ dashes) land as part of any of this,
   independently, even if the answer stays "hide"? They are the only *bugs* found (BOM/`+++`
   leak YAML into the render today).
8. If the YAML fails to parse (invalid frontmatter), what is shown: raw block, an error
   marker, or nothing? (Obsidian shows source/a warning; no terminal precedent.)
9. New domain term: "Frontmatter" is absent from `GLOSSARY.md`. If shown/used, it needs an
   entry (and perhaps "Properties" vs "Metadata block").

## Sources

Repo (noggin, at 2026-10-14):
- `cmd/noggin/document.go:31-67,70,91-111,181-205`; `viewer.go:466-531,590-605`;
  `find.go`; `picker.go:595,1040-1090`; `theme_markdown.go:87-130,150`; `theme.go`;
  `unfocused_style.go:32-48`; `model.go:2156-2190`; `location.go:32,97`; `keymap.go`.
- `.charm/DECISIONS.md` D5, D10, D34; `.charm/ROADMAP.md:70-76`; `.charm/NOTES.md:25,126,252,333-337`;
  `.charm/GLOSSARY.md`; `.charm/knowledge/frontmatter-rendering.md`.
- Tests that assert frontmatter is absent: `viewer_test.go:13-18,51-71,97-107,147`,
  `picker_test.go:722-770`, `test/pty/checks.py:227`, `testdata/fixture.md:1-5`.

Module cache (`/home/agent/go/pkg/mod`):
- `charm.land/glamour/v2@v2.0.1/glamour.go:25,35,68-97,146-152,173-179,208-213,226-231`;
  `ansi/renderer.go:17-37,65-73`; `ansi/elements.go:333-346,433-462,478`;
  `ansi/codeblock.go:80-110`; `ansi/table.go:186`.
- `go.abhg.dev/goldmark/frontmatter@v0.3.0/{README.md,CHANGELOG.md,extend.go,parse.go:59,125,177-193,format.go,data.go}`.
- `github.com/yuin/goldmark-meta@v1.1.0/meta.go:61,95-165,200-224`.
- `github.com/adrg/frontmatter@v0.2.0/{format.go:57-68,parser.go}`.
- `gopkg.in/yaml.v3@v3.0.1/{README.md,yaml.go:344}`; `github.com/goccy/go-yaml@v1.19.2/yaml.go:70-77,243`.

Network (fetched 2026-10-14; `help.obsidian.md`, `gohugo.io`, `jekyllrb.com`, `github.blog`,
`web.archive.org`, `spec.commonmark.org`, `yaml.org` returned "Approval required"/403, so first-party
doc *sources* on GitHub were used instead):
- https://github.com/charmbracelet/glamour (main @ f71347d, 2026-10-08) `glamour.go:70-77`; issue
  https://github.com/charmbracelet/glamour/issues/205
- https://github.com/charmbracelet/glow (main @ 8b0600f, 2026-10-02) `utils/utils.go`, `main.go`, `ui/ui.go`
- https://github.com/Textualize/frogmouth `ChangeLog.md`, `frogmouth/widgets/viewer.py`
- https://github.com/swsnr/mdcat (@ 5012892) `src/lib.rs`; pulldown-cmark `src/lib.rs`
- https://github.com/obsidianmd/obsidian-help `en/Editing and formatting/{Properties,Tags}.md`
- https://github.com/gohugoio/hugoDocs `content/en/content-management/front-matter.md`;
  https://github.com/gohugoio/hugo `parser/pageparser/pagelexer_intro.go`
- https://github.com/jekyll/jekyll `docs/_docs/front-matter.md`
- https://docs.github.com/api/search/v1 (queries for front matter); https://api.github.com/markdown
- GitHub repo metadata (archived/pushed/license) via `api.github.com/repos/<owner>/<repo>`.
- Go module proxy (`go list -m -json <mod>@latest`).

Throwaway code (not in repo, will not persist): `/tmp/fmres/{main,g,edge}.go`,
`/tmp/fmres/bench_test.go`, `/tmp/dep/*` (dependency sizing).
