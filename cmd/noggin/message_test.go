package main

import (
	"strings"
	"testing"
)

// TestMessageClearsOnNextKeypress checks S11's "clearing on the next
// keypress": a transient footer message set by a rejected action does not
// survive the keypress after it, even one that does not itself produce a new
// message (M4.2).
func TestMessageClearsOnNextKeypress(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m.message = "a rejected action's message"

	// A keypress that itself does nothing message-worthy (movement in an
	// empty pane) still clears the stale message.
	m = press(t, m, keys.LineDown.Keys()[0])

	if m.message != "" {
		t.Errorf("message after a keypress = %q, want cleared", m.message)
	}
}

// TestMessageDisplayedThenClearedNextKeypress drives the same behavior
// through the rendered frame, checking the message is visible immediately
// after the action that set it and gone after the next keypress.
func TestMessageDisplayedThenClearedNextKeypress(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m.message = "cannot save starred roots: no resolvable config directory"

	if !strings.Contains(m.frame(), "cannot save starred roots") {
		t.Fatalf("frame does not show the message before the next keypress:\n%s", m.frame())
	}

	m = press(t, m, keys.LineDown.Keys()[0])

	if strings.Contains(m.frame(), "cannot save starred roots") {
		t.Errorf("frame still shows the message after a keypress:\n%s", m.frame())
	}
}
