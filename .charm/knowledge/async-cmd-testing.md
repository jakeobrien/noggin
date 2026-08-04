# Testing async tea.Cmd work without real time

As of: C2

Written at C2/M2.4 (2026-08-03), the codebase's first async command. Facts and citations.

## The shape used

- Debounce + stale-discard both reduce to one generation counter compared on message arrival
  (`cmd/noggin/picker.go`, `queuePickerSearch`/`dispatchPickerSearch`/`applyPickerResults`).
  See D41 for the decision.
- Two messages: `pickerSearchMsg{seq}` (matured `tea.Tick`) and `pickerResultsMsg{seq,
  results}` (search reporting back). `model.Update` handles both alongside `editFinishedMsg`
  (`cmd/noggin/model.go`).

## Testing it takes no sleeps and no goroutines

- Never execute the `tea.Tick` Cmd in a test: it would actually sleep the debounce. A tick's
  entire observable content is the seq it carries, so the test synthesizes
  `pickerSearchMsg{seq}` for each keystroke's recorded `picker.seq` and plays them back in
  order — exactly what the real timers deliver (`picker_test.go`, `typePickerQuery` /
  `deliverPickerTicks`).
- "A search was issued" is observable as *`Update` returning a non-nil Cmd for a
  `pickerSearchMsg`* — `dispatchPickerSearch` returns nil for every superseded timer. That is
  how M2.4's "fewer searches than keystrokes" is asserted without a counter or a test seam.
- The search Cmd itself is safe to call synchronously in a test (`cmd()`) and returns the
  results message, so a full type → debounce → search → apply cycle runs deterministically in
  one goroutine, in ~0.01s against a real rg over a `t.TempDir()` Root.
- Interleaving (results landing after further typing) is expressed by holding the Cmd,
  typing more, then calling it — no concurrency needed to test a race that Bubble Tea's own
  single-threaded `Update` resolves by message order.
