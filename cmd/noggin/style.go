package main

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// dimColor is the muted foreground every pane's empty-state message shares
// with the footer's own resting text (Files Nav, Sources Nav, Document Viewer,
// footer) — one declaration so retuning it is a one-file edit (T25), rather
// than the same lipgloss.Color("245") literal independently repeated in four
// files. A var, not a const: lipgloss.Color returns the color.Color
// interface, which cannot be a constant expression.
var dimColor = lipgloss.Color("245")

// staleRowStyle is S9/C3/M5.1's dedicated staleness row: reverse video
// (Reverse swaps the specified Foreground into the row's background) in the
// same highlight color the Help Modal's border already uses (214,
// modalStyle's own "a color no pane border uses" pick) — reused here rather
// than adding a second, arbitrary "pay attention" color.
var staleRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Reverse(true)

// locationLineStyle is S9/S9a's own de-emphasized bottom row (M8.2): the
// same dimColor every pane's empty-state/footer text already uses, plain
// (no Reverse) unlike staleRowStyle — S9 never calls it out for attention
// the way S10's staleness row does, only asks that it read as secondary.
var locationLineStyle = lipgloss.NewStyle().Foreground(dimColor)

// unfocusedTintCode is S11's subtly-lighter background tint for an
// unfocused pane (rootsPane, filesPane, dv1Pane, dv2Pane), as a raw ANSI256
// code string: a dark grey a shade above a typical terminal's own black,
// chosen over darkening the focused pane (S11's own "a no-op on an
// already-black terminal" reasoning). Kept as a string (not just the
// lipgloss.Color below) because unfocusedTintSGR needs it spliced directly
// into a hand-built escape sequence — see tintPaneBackground's own doc
// comment for why that can't go through a plain Style.Render.
const unfocusedTintCode = "236"

// unfocusedTintSGR is unfocusedTintCode's own extended-color background
// escape sequence ("48;5;<n>", the same form ansi.Style.BackgroundColor
// emits for an ExtendedColor) — built directly rather than via
// lipgloss.NewStyle().Background(...).String(), which would wrap an empty
// string in its own immediate reset (ansi.Style.Styled("") returns just
// "": Style.String() alone, with no trailing content to attach to, is not
// a usable splice-in prefix).
const unfocusedTintSGR = "\x1b[48;5;" + unfocusedTintCode + "m"

// tintPaneBackground applies S11's unfocused-pane tint to an already-
// rendered pane's title or content — one SGR-tracking pass over the
// string, not a single wrapping Style.Render. lipgloss always closes a
// nested Style.Render with a bare full reset (ansi.Style.Styled's
// ResetStyle suffix — the same mechanism lipgloss-v2-row-rendering-
// gotchas.md documents cancelling a reverse-video row's bar mid-line): a
// pane's content is many such independently-closed runs (the Sources/Files
// Nav's own per-row Style.Render calls, a viewport's glamour-rendered
// markdown spans, Peek's inverted title), so a background set only once at
// the very start would be cancelled by the first run's own closing reset
// and never reach anything after it.
//
// Reverse-video runs (a selected row, S5; Peek's or an in-flight fetch's
// inverted title, S7/C6-M4.2) are deliberately left untouched: Reverse
// swaps whatever foreground/background are *already active* at render
// time, so injecting an explicit background while it is in force would
// make the tint itself the swapped-in text colour instead of the
// terminal's own default pair — changing what S5/S7 already look like
// rather than just tinting the pane around them, which S11 never asks for.
//
// The tricky part is the row boundary itself (the "\n" between one row's
// own Style.Render output and the next's): renderBox hands this function's
// own output to a *second*, outer Style.Render (the box that adds the
// border) — and lipgloss's own word-wrap step re-asserts whatever SGR
// state was active right before a line break at the start of the next
// line, to keep a genuinely soft-wrapped, ANSI-styled paragraph looking
// continuous across the break. Confirmed live against the real binary
// (`make check-pty`): leaving the tint merely "active" across a "\n" —
// even without writing anything *new* there — let that outer step
// re-inject it at the start of the following row, corrupting a reverse-
// video row (a selected Sources/Files Nav entry) that happened to follow a
// tinted one. So this function explicitly clears the tint (SGR 49, default
// background — selective, unlike a full reset, so it never has to guess
// whether some other attribute needs to survive the boundary) immediately
// before every "\n" where it was active, and re-decides whether to
// (re-)inject it fresh once it has seen what actually opens the next row.
func tintPaneBackground(s string) string {
	if s == "" {
		return s
	}
	const clearTintSGR = "\x1b[49m"
	var (
		b             strings.Builder
		reverseActive bool // is SGR 7 (reverse video) currently in force?
		tintActive    bool // is the tint the current, explicitly-set background?
		runes         = []rune(s)
	)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\x1b' || i+1 >= len(runes) || runes[i+1] != '[' {
			if runes[i] == '\n' {
				if tintActive {
					b.WriteString(clearTintSGR)
					tintActive = false
				}
				b.WriteRune('\n')
				continue
			}
			// A plain character reached while reverse is in force is *inside*
			// a reverse-video run (its own opening escape already declined
			// to activate the tint, below) — must stay untouched for exactly
			// the same reason as that opening escape itself.
			if !reverseActive && !tintActive {
				b.WriteString(unfocusedTintSGR)
				tintActive = true
			}
			b.WriteRune(runes[i])
			continue
		}
		j := i + 2
		for j < len(runes) && runes[j] != 'm' {
			j++
		}
		if j >= len(runes) {
			// Malformed/truncated escape — lipgloss never emits one; copy
			// whatever remains verbatim rather than risk corrupting it.
			b.WriteString(string(runes[i:]))
			return b.String()
		}
		setsReverse, clearsReverse, isReset := false, false, false
		for _, tok := range strings.Split(string(runes[i+2:j]), ";") {
			switch tok {
			case "7":
				setsReverse = true
			case "27":
				clearsReverse = true
			case "", "0":
				isReset = true
			}
		}
		willBeReverse := reverseActive
		if setsReverse {
			willBeReverse = true
		}
		if clearsReverse || isReset {
			willBeReverse = false
		}
		switch {
		case willBeReverse:
			// Reverse is about to govern colour (or already does); an
			// explicit background would become its swapped-in text colour
			// instead of the terminal's own default pair, so no tint may be
			// active once this escape takes effect — clearing a still-active
			// one from earlier in the same run, not just skipping a fresh
			// injection right here.
			if tintActive {
				b.WriteString(clearTintSGR)
			}
			tintActive = false
		case isReset:
			// A bare/"0" reset clears everything in reality (including any
			// background, tint or otherwise) regardless of what this
			// function did or didn't inject before it — no point injecting
			// tint only to have this escape immediately erase it.
			tintActive = false
		case !tintActive:
			b.WriteString(unfocusedTintSGR)
			tintActive = true
		}
		b.WriteString(string(runes[i : j+1]))
		reverseActive = willBeReverse
		i = j
	}
	return b.String()
}
