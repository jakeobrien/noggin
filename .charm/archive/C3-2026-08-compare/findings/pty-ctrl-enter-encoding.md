# Putting `ctrl+enter` on the wire in a pty check

Written at C3/M1.7 (TRIAGE T6). Facts and citations.

## The problem

`keys.ActivateDV2` (`cmd/noggin/keymap.go`) binds two keys — `ctrl+enter` (primary) and
`ctrl+l` (fallback, D55) — precisely because `ctrl+enter` has no reliable raw-byte
encoding: without the Kitty keyboard protocol, `CSI u` (fixterms), or xterm's
`modifyOtherKeys` negotiated, a real terminal sends it byte-identical to plain `enter`
(keymap.go's own comment at the `ActivateDV2` field). `test/pty/harness.py`'s startup
handshake answers neither query (`harness.py`'s `REPLIES`), so a literal "send what a
plain terminal sends for ctrl+enter" byte string is indistinguishable from `Activate`'s
own `\r` — a pty check driving that would silently test `Activate`, not `ActivateDV2`.

## The fix: send the CSI u form, which decodes unconditionally

`github.com/charmbracelet/ultraviolet` (bubbletea v2's input decoder,
`decoder.go`'s `EventDecoder.parseCsi`) parses `CSI ... u` sequences as Kitty-keyboard/
`fixterms` input *unconditionally* — the `case 'u':` branch (`decoder.go:515`) is not
gated on whatever the `\x1b[?u` flags query or the DECRQM `2026`/`2027` queries
answered; it fires for any well-formed `u`-terminated CSI sequence regardless of
negotiation state. `parseKittyKeyboard` (`decoder.go:1389`) maps the first parameter
through `kittyKeyMap`, where `ansi.CR` (13) → `KeyEnter` (`decoder.go:1203`), and the
second parameter through `fromKittyMod(mod-1)`, where `5` (`1 + ctrl's bit, 4`) → Ctrl
only.

So `\x1b[13;5u` reaches noggin's model exactly as a `KeyPressEvent{Code: KeyEnter,
Mod: ModCtrl}` — the same event a Kitty/CSI-u-capable terminal would send for a real
`ctrl+enter` keypress — without either side negotiating anything. `test/pty/checks.py`'s
`KEY["ActivateDV2CtrlEnter"]` uses this; verified by regression injection (temporarily
dropping `ctrl+enter` from `keys.ActivateDV2`'s key list makes exactly the two checks
built on that KEY entry fail, and only those).

## Scope

General to any future chord whose only wire-distinguishable form is a modifier the
terminal cannot express as a raw C0/C1 byte (arrow+ctrl, function keys, etc.) — reread
before adding pty coverage for one. Chords that *do* have a raw control byte (any
`ctrl+<letter>`) don't need this; `ord(letter) & 0x1f` is what the rest of `checks.py`'s
`KEY` table already uses.
