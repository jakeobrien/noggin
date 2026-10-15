# Testing async `tea.Cmd` work without real time

The Picker's search is the reference shape for async work (`cmd/noggin/picker.go`): read it before adding the next async command.

## The shape

Debounce and stale-result discard both reduce to one generation counter, `picker.seq`, compared when a message arrives:

- every change to what would be searched bumps `seq` and schedules a `tea.Tick` carrying it (`pickerSearchMsg{seq}`);
- a matured tick dispatches the search only if its `seq` is still current (`dispatchPickerSearch`);
- the search reports back as `pickerResultsMsg{seq, results}`, applied only if `seq` is still current (`applyPickerResults`).

The counter is monotonic for the model's lifetime: dismissing and reopening the Picker bumps it rather than resetting it, so a search in flight across a dismissal can't be accepted by the next Picker. Bubble Tea can't cancel a started Cmd, so discarding has to happen on arrival. The search Cmd (`pickerSearchCmd`) is a pure function of copied-in values and touches no model state, since it runs concurrently with `Update`.

## Testing it: no sleeps, no goroutines, no test seams

- **Never execute the `tea.Tick` Cmd**: it would really sleep. A tick's whole observable content is the `seq` it carries, so tests synthesise `pickerSearchMsg{seq}` for each keystroke and play them back in order (`typePickerQuery` / `deliverPickerTicks` in `picker_test.go`).
- **"A search was issued" = `Update` returned a non-nil Cmd** for a `pickerSearchMsg`; superseded ticks return nil. That's how "fewer searches than keystrokes" is asserted without a counter.
- **The search Cmd is safe to call synchronously** (`cmd()`), returning the results message. A full type → debounce → search → apply cycle runs deterministically on one goroutine against real `rg` over a `t.TempDir()` Root.
- **Races are expressed by ordering**: hold the Cmd, type more, then call it. Bubble Tea's single-threaded `Update` resolves races by message order, so no concurrency is needed to test them.

For pty coverage of async work, see the race note in `pty-harness.md`.
