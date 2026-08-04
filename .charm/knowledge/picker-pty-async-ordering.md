# pty-replaying the Picker's async search: an ordering hazard, not a product bug

As of: C2

Written at C2/M7.1 (2026-08-04). Facts and citations.

## The symptom

Adding `test/pty/checks.py` coverage for the Picker's mode/Scope switch (`PickerMode`/
`PickerScope`, `cmd/noggin/keymap.go`) intermittently misread the mode/Scope indicator
title (`pickerModeScope`, `picker_view.go`) when the switch followed a keystroke that had
already landed a completed async search (`picker.go`'s `queuePickerSearch`/
`dispatchPickerSearch`/`applyPickerResults`, D41). Same trailing bytes on the wire, replayed
through `test/pty/harness.py`'s `Screen` emulator, landed the title text on the right row in
~75% of runs and the wrong one in the rest — confirmed by diffing the exact byte tail across
passing and failing runs: identical tail, divergent result, so the divergence traces to
*earlier* bytes in the same session (the async search's own redraw, whose timing relative to
the next keypress is a genuine race) shifting the emulator's cumulative cursor-position state
before the tail's cursor-relative moves (`\x1b[NA`/`\x1b[ND` etc.) are applied.

Isolated by bisecting the action sequence, not by reading `Screen`'s CSI handling — the same
CSI final-byte set (`A C D J P X`) appears in passing and failing runs alike, so this is a
timing/ordering hazard in what gets replayed, not a missing escape code.

## The fix: never switch mode/Scope immediately after a landed search

Reordering pty actions to never do both at once made 5 consecutive full `make check-pty` runs
100% clean (previously ~1-in-4 checks failed per run):

- Switch mode/Scope while the Prompt is still empty (`queuePickerSearch`'s own
  `!m.searchable()` early return means an empty-Prompt switch dispatches no search at all —
  provably zero async redraws in flight) — verified clean across 10 isolated trials.
- Only then type the query, once already in the target mode/Scope, so exactly one search
  settles and nothing switches mode/Scope afterward.
- Keep any post-search action (activate) in the *same* mode/Scope the search ran in.

`test/pty/checks.py`'s M7.1 section follows this split: a File-mode flow (type → results →
preview → activate, no mode/Scope switch at all) and a separate mode/Scope flow (switch →
switch → switch, all on an empty Prompt, then one Grep-mode search).

## Scope of the finding

Specific to the Picker's pty coverage — the codebase's first genuinely async `tea.Cmd`
(`async-cmd-testing.md`) meeting the codebase's first cell-diffing pty replay
(`.charm/knowledge/pty-verification.md`). No other pty-tested feature combines an
in-flight background search with an immediately-following action that changes the frame
again, so this has not been seen (and would not be expected) anywhere else in
`checks.py`. Worth rereading before adding pty coverage for any *future* async command, or
before adding a pty check that both types into the Picker and switches its mode/Scope in the
same breath.
