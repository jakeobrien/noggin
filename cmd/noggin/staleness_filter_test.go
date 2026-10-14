package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// --- directory re-list filter: only changes the tree would show count ------

// staleAfter snapshots a fresh tree at root, bumps dir's mtime back so the
// change cannot land in the same mtime tick, applies change, and reports
// what stale() says.
func staleAfter(t *testing.T, root, dir string, change func()) bool {
	t.Helper()
	b := snapshotStaleness(walkFor(t, root), "")
	if b.stale() {
		t.Fatalf("setup: baseline already stale before any change")
	}
	bumpMtimeForward(t, dir)
	change()
	return b.stale()
}

// TestStaleIgnoresEntriesTheTreeDoesNotShow: an entry the tree would never
// list still bumps its directory's mtime — editor swap and probe files,
// Finder's .DS_Store, a non-Readable file — but refreshing would show no
// difference, so none of them may raise the indicator.
func TestStaleIgnoresEntriesTheTreeDoesNotShow(t *testing.T) {
	for _, name := range []string{".doc.md.swp", "4913", ".DS_Store", "image.png", "doc.md~"} {
		t.Run(name, func(t *testing.T) {
			root := mkdir(t, filepath.Join(t.TempDir(), "root"))
			sub := mkdir(t, filepath.Join(root, "sub"))
			writeFile(t, filepath.Join(sub, "doc.md"))
			if staleAfter(t, root, sub, func() { writeFile(t, filepath.Join(sub, name)) }) {
				t.Errorf("stale() = true after adding %q, which the tree does not show, want false", name)
			}
		})
	}
}

// TestStaleIgnoresAGitignoredAddition: a Readable file the tree hides via
// .gitignore is no more visible than a non-Readable one.
func TestStaleIgnoresAGitignoredAddition(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	writeFileContent(t, filepath.Join(root, ".gitignore"), "drafts.md\n")
	writeFile(t, filepath.Join(root, "doc.md"))
	if staleAfter(t, root, root, func() { writeFile(t, filepath.Join(root, "drafts.md")) }) {
		t.Error("stale() = true after adding a gitignored file, want false")
	}
}

// TestStaleIgnoresAnAtomicSaveOfANonActiveDocument: write-temp-then-rename
// (how many editors save) churns the directory but leaves the same entries.
// The document isn't open, so there is nothing the user would see change.
func TestStaleIgnoresAnAtomicSaveOfANonActiveDocument(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	doc := writeFile(t, filepath.Join(root, "doc.md"))
	if staleAfter(t, root, root, func() {
		tmp := writeFileContent(t, filepath.Join(root, ".doc.md.tmp"), "new\n")
		if err := os.Rename(tmp, doc); err != nil {
			t.Fatalf("Rename: %v", err)
		}
	}) {
		t.Error("stale() = true after an atomic save of a non-open document, want false")
	}
}

// TestStaleDetectsVisibleChanges is the other side of the filter: every
// change the tree would show after a refresh still raises the indicator.
func TestStaleDetectsVisibleChanges(t *testing.T) {
	cases := map[string]func(t *testing.T, root string){
		"readable file added":   func(t *testing.T, root string) { writeFile(t, filepath.Join(root, "new.md")) },
		"empty directory added": func(t *testing.T, root string) { mkdir(t, filepath.Join(root, "newdir")) },
		"readable file removed": func(t *testing.T, root string) {
			if err := os.Remove(filepath.Join(root, "doc.md")); err != nil {
				t.Fatal(err)
			}
		},
		"readable file renamed": func(t *testing.T, root string) {
			if err := os.Rename(filepath.Join(root, "doc.md"), filepath.Join(root, "renamed.md")); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			root := mkdir(t, filepath.Join(t.TempDir(), "root"))
			writeFile(t, filepath.Join(root, "doc.md"))
			if !staleAfter(t, root, root, func() { change(t, root) }) {
				t.Errorf("stale() = false after %s, want true", name)
			}
		})
	}
}

// TestStaleDetectsAGitignoreEditedInPlace: editing a .gitignore can change
// what the tree shows anywhere beneath it without touching any directory's
// mtime, so a changed .gitignore is itself treated as a change.
func TestStaleDetectsAGitignoreEditedInPlace(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "doc.md"))
	ignore := writeFileContent(t, filepath.Join(root, ".gitignore"), "*.log\n")

	b := snapshotStaleness(walkFor(t, root), "")
	bumpMtimeForward(t, ignore)
	if err := os.WriteFile(ignore, []byte("*.log\nsub/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !b.stale() {
		t.Error("stale() = false after a .gitignore was edited in place, want true")
	}
}

// TestStaleDetectsAGitignoreCreated: a new .gitignore is not itself a tree
// entry, so the re-listed directory can look unchanged — but its rules may
// hide things further down, so its appearance counts as a change.
func TestStaleDetectsAGitignoreCreated(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "doc.md"))
	if !staleAfter(t, root, root, func() {
		writeFileContent(t, filepath.Join(root, ".gitignore"), "nothing-here\n")
	}) {
		t.Error("stale() = false after a .gitignore was created, want true")
	}
}

// TestStaleDetectsAGitignoreRemoved: the reverse — what it hid reappears.
func TestStaleDetectsAGitignoreRemoved(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	writeFileContent(t, filepath.Join(root, ".gitignore"), "nothing-here\n")
	writeFile(t, filepath.Join(root, "doc.md"))
	if !staleAfter(t, root, root, func() {
		if err := os.Remove(filepath.Join(root, ".gitignore")); err != nil {
			t.Fatal(err)
		}
	}) {
		t.Error("stale() = false after a .gitignore was removed, want true")
	}
}

// --- terminal focus trigger -------------------------------------------------

// TestViewReportsFocus: the View must ask the terminal for focus events, or
// FocusMsg never arrives.
func TestViewReportsFocus(t *testing.T) {
	if !newModel().View().ReportFocus {
		t.Error("View().ReportFocus = false, want true")
	}
}

// TestFocusRunsAnImmediateCheck: regaining terminal focus (back from an
// editor in another window) checks right away rather than waiting up to a
// full interval, and a detected change raises the indicator.
func TestFocusRunsAnImmediateCheck(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	m := sized(t, newModel().openFolder(walkFor(t, root)), 80, 24)

	bumpMtimeForward(t, root)
	writeFile(t, filepath.Join(root, "new.md"))

	next, cmd := m.Update(tea.FocusMsg{})
	m = concrete(t, next)
	if cmd == nil {
		t.Fatal("Update(FocusMsg) returned a nil Cmd, want an immediate check")
	}
	msg, ok := cmd().(staleResultMsg)
	if !ok {
		t.Fatalf("Update(FocusMsg)'s Cmd produced %T, want staleResultMsg", cmd())
	}

	next, cmd = m.Update(msg)
	m = concrete(t, next)
	if !m.stale {
		t.Error("m.stale = false after a focus check found a change, want true")
	}
	// The periodic tick is already armed independently; a focus check must
	// not schedule a second one, or each focus would add another timer.
	if cmd != nil {
		t.Error("applying a focus check's result returned a Cmd, want nil (no extra tick)")
	}
}

// TestPeriodicCheckStillReschedules guards the other half of that split.
func TestPeriodicCheckStillReschedules(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	m := sized(t, newModel().openFolder(walkFor(t, root)), 80, 24)
	_, cmd := m.runStaleCheck(staleCheckMsg{})
	msg := cmd().(staleResultMsg)
	if _, next := m.applyStaleResult(msg); next == nil {
		t.Error("applying a periodic check's result returned nil, want the next tick")
	}
}
