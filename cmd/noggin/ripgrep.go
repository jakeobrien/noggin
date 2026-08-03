package main

import "os/exec"

// ripgrepBinary is the executable M1's search layer shells out to. Detection
// looks it up by name on PATH rather than assuming a fixed location, the
// same portability tradeoff exec.LookPath makes for any PATH-resolved tool.
const ripgrepBinary = "rg"

// ripgrepAvailability is the outcome of detecting ripgrep once at session
// startup (S14): whether the Pickers have a search layer to run on, and,
// when they do not, why — so a Picker reached for later can say (M2.5).
type ripgrepAvailability struct {
	available bool
	reason    string // why unavailable; "" when available
}

// detectRipgrep looks up the ripgrep binary on PATH. It is a pure query — it
// never affects the launch path (S14: "the session still launches fully
// usable for reading" regardless of the outcome); main.go folds the result
// into session state once, and nothing consults it until the Picker exists.
func detectRipgrep() ripgrepAvailability {
	if _, err := exec.LookPath(ripgrepBinary); err != nil {
		return ripgrepAvailability{reason: "ripgrep (rg) not found on PATH"}
	}
	return ripgrepAvailability{available: true}
}
