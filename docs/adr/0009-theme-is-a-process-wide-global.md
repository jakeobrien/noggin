# The Theme is a process-wide global, written from Update

The active Theme is a package-level `theme` variable. `model.cycleTheme` writes to it, and every Theme-coloured style is a function that reads it when called. This deliberately breaks the Bubble Tea pattern of keeping all state in the model, because passing a Theme through every renderer would touch most files and change no behaviour. Open viewers re-render in place on a switch, which keeps scroll and Find state.

## Consequences

- A Theme-coloured style written as a package-level `var` captures the colours at startup and silently misses switches. `TestThemeKeyRepaintsTheWholeFrame` catches this for anything visible in its fixture frame.
- Tests that switch Themes must restore the global (`keepTheme`).

_Formerly D142._
