# Charm v2: what it is, what it costs us, what it buys us

Checked 2026-07-31 against the live module proxy from inside the sandbox, not just docs.
Every API shape below was confirmed with `go doc` against the real modules in the module
cache; doc-only claims are marked as such.

## The versions are stable, and the import path moved

| Module | Latest | Was (C1 plan time) |
|---|---|---|
| `charm.land/bubbletea/v2` | v2.0.8 | `github.com/charmbracelet/bubbletea` v1.3.10 |
| `charm.land/bubbles/v2` | v2.1.1 | `github.com/charmbracelet/bubbles` v1.0.0 |
| `charm.land/lipgloss/v2` | v2.0.5 | `.../lipgloss` v1.1.1-0.20250404203927-… |
| `charm.land/glamour/v2` | v2.0.1 | `github.com/charmbracelet/glamour` v1.0.0 |

**The v2 modules are not importable under `github.com/charmbracelet/...`.** Charm renamed the
module path to the `charm.land` vanity domain. The old path still has git tags, so
`go list -m -versions github.com/charmbracelet/bubbletea/v2` lists up to v2.0.8 and looks
fine — but fetching it fails:

```
go: github.com/charmbracelet/bubbletea/v2@v2.0.8 requires …: parsing go.mod:
    module declares its path as: charm.land/bubbletea/v2
            but was required as: github.com/charmbracelet/bubbletea/v2
```

`go get charm.land/{bubbletea,bubbles,lipgloss,glamour}/v2@latest` resolves cleanly in this
sandbox (verified). New indirect deps of note: `charmbracelet/ultraviolet` (the new cell-based
renderer, a `v0.0.0-2026…` pseudo-version) and `charmbracelet/x/termios`.

## All-or-nothing

`bubbles/v2` imports `charm.land/bubbletea/v2` and `charm.land/lipgloss/v2` directly, and every
component's `Update` switches on v2 message types. There is no shim: bubbles v2 + lipgloss v2
require bubbletea v2 (bubbles' own `UPGRADE_GUIDE_V2.md`). Glamour v2 is the exception — it
depends on no bubbletea at all — but it does pull lipgloss v2, so a partial migration means two
lipgloss majors in the graph. Migrate the four together or not at all.

## What actually changes in our ~600 lines

- **`View() string` → `View() tea.View`.** The headline break. `tea.NewView(s)` wraps a string;
  the struct's other fields (`AltScreen`, `MouseMode`, `Cursor`, `WindowTitle`,
  `BackgroundColor`, `KeyboardEnhancements`…) replace the old `tea.With*` program options and
  imperative commands, which are **removed**. `Init() Cmd` and `Update(Msg) (Model, Cmd)` are
  unchanged. We pass no program options today (`main.go:31`), so only the return type bites.
  Gotcha (doc-derived): those fields are re-derived every render, so conditionally setting
  `AltScreen` in one code path flickers the terminal.
- **`tea.KeyMsg` is now an interface**, implemented by `tea.KeyPressMsg` and
  `tea.KeyReleaseMsg`. `KeyPressMsg` is a `tea.Key` with `Code rune` (replaces `Type`),
  `Text string` (replaces `Runes []rune`) and `Mod KeyMod` (replaces `Alt bool`). The
  `tea.KeyRunes` / `tea.KeyCtrlC` / `tea.KeyType` constants are gone. This deletes
  `keymap.go`'s whole `namedKeys map[string]tea.KeyType` table and its `keyMsg` helper
  (`keymap.go:80-107`) — synthetic keys in tests become struct literals,
  `tea.KeyPressMsg{Code: 'k', Text: "k"}` or `tea.KeyPressMsg{Code: tea.KeyTab}`.
  `viewer.go:131`'s `scroll(msg tea.KeyMsg)` must take the concrete press type.
  Note: `msg.String()` now returns `"space"`, not `" "`.
- **`key.Matches` is generic**: `Matches[Key fmt.Stringer](k Key, b ...Binding) bool`. Call
  sites are source-compatible; only code storing `tea.KeyMsg` as a concrete struct changes.
- **bubbles `viewport`**: `New(80, 24)` → `New(WithWidth(80), WithHeight(24))`, and `Width`,
  `Height`, `YOffset` stop being public fields in favour of getters/setters. `SetContent`,
  `PageUp/Down`, `HalfPageUp/Down`, `ScrollUp/Down`, `GotoTop/Bottom` keep their names, so
  `scrollActions` survives as a table. `HighPerformanceRendering` is gone.
- **lipgloss**: `lipgloss.Color("240")` still works — it is now a *function* returning
  `image/color.Color`. `NewStyle`, `Border`, `RoundedBorder`, `JoinVertical`, `JoinHorizontal`
  are unchanged in shape. Two real hazards: (1) the `*Renderer` type and the global colour
  profile are **removed**, and (2) `Style.Width()`/`Height()` now **include the border** in the
  measured size (lipgloss PR #451, flagged as breaking) — `layout.go`'s `paneBorderWidth` /
  `rect.content()` arithmetic must be re-verified, not assumed.
  `HasDarkBackground()` becomes `HasDarkBackground(in, out term.File) bool` — explicit I/O, no
  error return. `AdaptiveColor` moved to `lipgloss/v2/compat`.
  Also: lipgloss v2 converts `\t` to 4 spaces at render time by default
  (`Style.TabWidth`), which may make our `expandTabs` redundant — **unverified, test it**.
- **glamour**: `WithStandardStyle(style)` and `WithWordWrap(n)` both still exist (verified via
  `go doc`; a secondary source claimed `WithStandardStyle` was removed — it is not).
  `WithAutoStyle()` **is** removed and the default style is hardcoded `dark`. Glamour v2 no
  longer queries the terminal at all: "Querying the terminal is now part of Lip Gloss v2"
  (glamour PR #410). `WithColorProfile` is gone; downsampling moves to lipgloss/bubbletea.
- **`tea.ExecProcess(cmd, fn)` is unchanged.** M4.1's editor handoff is unaffected.

## What it buys us

1. **`viewport` gained search.** `SetHighlights([][]int)`, `HighlightNext()`,
   `HighlightPrevious()`, `ClearHighlights()` and `EnsureVisible(line, colstart, colend)` are
   all in v2.1.1 (verified). That is most of S8 / M5.1 — including the character-exact
   highlighting the spec calls desirable-but-not-required — as library behaviour instead of
   hand-rolled offset mapping over wrapped, styled output.
2. **The D13 hazard is designed out.** Bubble Tea v2 parses OSC 10/11/12 inline in its own
   input stream and offers `tea.RequestBackgroundColor()` → `tea.BackgroundColorMsg.IsDark()`;
   lipgloss v2 was made pure so that Bubble Tea solely owns terminal I/O. The specific v1
   deadlock — a library's OSC 11 query racing Bubble Tea's reader — cannot recur through this
   path. It can still recur for any third-party code doing its own raw query, so D13's rule
   stands as a rule; only its cause is gone. Known rough edges (doc-derived): background-colour
   replies are unreliable under tmux (bubbletea#1160) and mis-parsed on WSL2 (#1440) — keep the
   before-`Run()` `HasDarkBackground` approach with a fallback rather than depending on the
   message arriving.
3. **Deterministic tests.** `tea.WithWindowSize(w, h)` and `tea.WithColorProfile(p)` set size
   and colour without a real TTY — directly against the pain in `findings/pty-verification.md`,
   where `go test` never sees colour and the pty harness must answer startup queries by hand.
4. `SoftWrap`, `LeftGutterFunc`, `StyleLineFunc` and horizontal scrolling in `viewport`, plus
   the ultraviolet cell-diffing renderer.

## What it costs / risks

- `bubbles/v2` has **no tree component** (packages: cursor, filepicker, help, key, list,
  paginator, progress, spinner, stopwatch, table, textarea, textinput, viewport — verified).
  M2.2 is hand-rolled either way; v2 changes nothing there.
- v2.0.0 → v2.0.8 in a few months is active churn on a young major, and `ultraviolet` is a
  pre-1.0 pseudo-version dependency.
- The pty recipe in `findings/pty-verification.md` was derived against v1's startup query set.
  v2's differ; the harness needs re-verification, not just re-running.

## Sources

- `github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md`, release notes for
  `bubbletea` v2.0.0, and PR #1080 (input parser).
- `github.com/charmbracelet/lipgloss/blob/main/UPGRADE_GUIDE_V2.md`, PR #451 (border in
  measured size), PR #438 (`HasDarkBackground` signature).
- `github.com/charmbracelet/glamour/blob/main/UPGRADE_GUIDE_V2.md`, PR #410 (auto-style
  removal), PR #409 (wrapping via cellbuf).
- `github.com/charmbracelet/bubbles/blob/main/UPGRADE_GUIDE_V2.md` ("Companion upgrades
  required").
- Local verification: `go get charm.land/{bubbletea,bubbles,lipgloss,glamour}/v2@latest`;
  `go doc` on `tea.Model`, `tea.View`, `tea.Key`, `tea.KeyPressMsg`, `viewport.Model`,
  `glamour.TermRendererOption`, `lipgloss.HasDarkBackground`, `lipgloss.Color`.
