package main

import (
	"fmt"
	"strings"
	"testing"

	"noggin/internal/tree"
)

// TestScrollOffset covers the sidebar panes' margin-scrolling rule at its
// seam: given the remembered offset and the cursor, where does the window
// start? Expected values are worked by hand from the one-third-margin rule
// (margin = perScreen/3, capped so the two margins never overlap).
func TestScrollOffset(t *testing.T) {
	const count, per = 30, 9 // margin 3: cursor lines at window rows 3 and 5
	tests := []struct {
		name                    string
		offset, selected, count int
		perScreen, want         int
	}{
		{"top, cursor free in first two thirds", 0, 5, count, per, 0},
		{"down past two-thirds line scrolls", 0, 6, count, per, 1},
		{"keeps cursor on the line while scrolling down", 4, 10, count, per, 5},
		{"up from last row does not scroll", 21, 28, count, per, 21},
		{"up from bottom still free above the top margin", 21, 24, count, per, 21},
		{"up crossing the top margin scrolls", 21, 23, count, per, 20},
		{"keeps cursor on the top line while scrolling up", 10, 12, count, per, 9},
		{"down after scrolling up is free until bottom margin", 9, 14, count, per, 9},
		{"down after scrolling up scrolls past bottom margin", 9, 15, count, per, 10},
		{"clamps at the top: cursor reaches row 0", 2, 0, count, per, 0},
		{"clamps at the bottom: cursor reaches last row", 20, 29, count, per, 21},
		{"selection far past the window: minimal jump", 0, 29, count, per, 21},
		{"selection far before the window: minimal jump", 21, 0, count, per, 0},
		{"stale offset past the end is re-clamped", 100, 29, count, per, 21},
		{"rows shrank below the offset", 21, 5, 12, per, 2},
		{"negative offset is re-clamped", -4, 0, count, per, 0},
		{"short list fits: no scroll", 3, 2, 5, per, 0},
		{"list exactly the pane height", 0, 8, 9, per, 0},
		{"pane of zero entries", 0, 4, count, 0, 0},
		{"tiny pane of 1 has no margin", 0, 3, count, 1, 3},
		{"tiny pane of 2 has no margin", 0, 2, count, 2, 1},
		{"pane of 3 uses margin 1", 0, 2, count, 3, 1},
		{"pane of 4 margin 1: cursor line is row 2", 0, 3, count, 4, 1},
		{"two-line rows: 10 lines is 5 entries, margin 1", 0, 4, count, 10 / rootRowHeight, 1},
		{"empty list", 0, 0, 0, per, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scrollOffset(tc.offset, tc.selected, tc.count, tc.perScreen); got != tc.want {
				t.Errorf("scrollOffset(%d, %d, %d, %d) = %d, want %d",
					tc.offset, tc.selected, tc.count, tc.perScreen, got, tc.want)
			}
		})
	}
}

func TestScrollMargin(t *testing.T) {
	for per, want := range map[int]int{0: 0, 1: 0, 2: 0, 3: 1, 5: 1, 6: 2, 9: 3, 10: 3, 30: 10} {
		if got := scrollMargin(per); got != want {
			t.Errorf("scrollMargin(%d) = %d, want %d", per, got, want)
		}
	}
}

func manyFilesNav(n, height int) filesNav {
	kids := make([]*tree.Node, n)
	for i := range kids {
		name := fmt.Sprintf("f%02d.md", i)
		kids[i] = &tree.Node{Name: name, Path: "/root/" + name}
	}
	root := &tree.Node{Name: "root", Path: "/root", IsDir: true, Children: kids}
	return newFiles(root).setSize(30, height)
}

// TestFilesNavScrollsWithMargin drives the Files Nav through the ticket's
// two scenarios: down from the top, then back up from the bottom.
func TestFilesNavScrollsWithMargin(t *testing.T) {
	f := manyFilesNav(30, 10) // 9-row list: margin 3
	step := func(delta int) {
		f = f.move(delta).syncScroll()
	}
	for i := 0; i < 5; i++ { // cursor 5: still inside the free zone
		step(1)
	}
	if f.offset != 0 {
		t.Fatalf("offset after 5 downs = %d, want 0", f.offset)
	}
	step(1) // cursor 6 passes the two-thirds line
	if f.offset != 1 {
		t.Fatalf("offset after 6 downs = %d, want 1", f.offset)
	}
	for i := 0; i < 30; i++ {
		step(1)
	}
	if f.selected != 29 || f.offset != 21 {
		t.Fatalf("at bottom: selected=%d offset=%d, want 29/21", f.selected, f.offset)
	}
	step(-1)
	step(-1)
	step(-1)
	step(-1) // cursor 25: still below the top margin line (24)
	if f.offset != 21 {
		t.Fatalf("offset after 4 ups = %d, want 21 (view must not move yet)", f.offset)
	}
	step(-1) // cursor 24: on the margin line
	if f.offset != 21 {
		t.Fatalf("offset with cursor on margin line = %d, want 21", f.offset)
	}
	step(-1) // cursor 23 crosses it
	if f.offset != 20 {
		t.Fatalf("offset after crossing top margin = %d, want 20", f.offset)
	}
	if got := f.view(); !strings.Contains(got, "f20.md") || !strings.Contains(got, "f28.md") || strings.Contains(got, "f29.md") {
		t.Errorf("view should show f20..f28:\n%s", got)
	}
}

func TestFilesNavScrollKeepsSelectionVisibleAfterJumpAndShrink(t *testing.T) {
	f := manyFilesNav(30, 10)
	f.selected = 29
	f = f.syncScroll()
	if f.offset != 21 {
		t.Fatalf("jump to last: offset = %d, want 21", f.offset)
	}
	f = f.setSize(30, 6).syncScroll() // 5-row list, margin 1
	if f.selected < f.offset || f.selected >= f.offset+5 {
		t.Fatalf("after resize selection %d outside window [%d,%d)", f.selected, f.offset, f.offset+5)
	}
	f.selected = 3
	f = f.syncScroll()
	if f.selected < f.offset || f.selected >= f.offset+5 {
		t.Fatalf("after jump selection %d outside window [%d,%d)", f.selected, f.offset, f.offset+5)
	}
}

func TestFilesNavShortListUnaffected(t *testing.T) {
	f := manyFilesNav(4, 10)
	f.selected = 3
	f = f.syncScroll()
	if f.offset != 0 {
		t.Errorf("offset = %d, want 0 for a list that fits", f.offset)
	}
	if got := f.view(); !strings.Contains(got, "f00.md") {
		t.Errorf("first row missing from view:\n%s", got)
	}
}

// TestRootsNavScrollsWithMarginInTwoLineRows checks the Sources Nav, whose
// window counts two-line entries: 10 lines is 5 entries, margin 1.
func TestRootsNavScrollsWithMarginInTwoLineRows(t *testing.T) {
	roots := make([]root, 19) // +1 Wikipedia row = 20 entries
	for i := range roots {
		roots[i] = root{path: fmt.Sprintf("/r/r%02d", i)}
	}
	r := newRootsNav().setRoots(roots, -1).setSize(30, 10)
	for i := 0; i < 3; i++ {
		r = r.move(1).syncScroll()
	}
	if r.offset != 0 {
		t.Fatalf("offset with cursor on row 3 = %d, want 0", r.offset)
	}
	r = r.move(1).syncScroll() // row 4 passes the bottom margin line (3)
	if r.offset != 1 {
		t.Fatalf("offset with cursor on row 4 = %d, want 1", r.offset)
	}
	r.selected = 19
	r = r.syncScroll()
	if r.offset != 15 {
		t.Fatalf("offset at last row = %d, want 15", r.offset)
	}
	r = r.move(-1).move(-1).syncScroll() // cursor 17: still below the top margin line (16)
	if r.offset != 15 {
		t.Fatalf("offset after 2 ups = %d, want 15", r.offset)
	}
	r = r.move(-1).move(-1).syncScroll() // cursor 15 < 16 → scroll
	if r.offset != 14 {
		t.Fatalf("offset after crossing top margin = %d, want 14", r.offset)
	}
}

// TestModelUpdatePersistsFilesNavOffset checks the wiring: key presses sent
// through Update leave the persisted offset current, so a later Up does not
// re-pin the cursor to the bottom edge.
func TestModelUpdatePersistsFilesNavOffset(t *testing.T) {
	kids := make([]*tree.Node, 40)
	for i := range kids {
		name := fmt.Sprintf("f%02d.md", i)
		kids[i] = &tree.Node{Name: name, Path: "/root/" + name}
	}
	root := &tree.Node{Name: "root", Path: "/root", IsDir: true, Children: kids}
	m := sized(t, newModel().openFolder(root), 80, 24)
	if m.focus != filesPane {
		t.Fatalf("test setup: focus = %v", m.focus)
	}
	for i := 0; i < 39; i++ {
		m = press(t, m, keys.LineDown.Keys()[0])
	}
	bottom := m.files.offset
	if bottom == 0 {
		t.Fatal("offset did not advance while scrolling down")
	}
	m = press(t, m, keys.LineUp.Keys()[0])
	if m.files.offset != bottom {
		t.Errorf("offset after one Up = %d, want unchanged %d", m.files.offset, bottom)
	}
}

// TestPickerListsKeepEdgePinnedScrolling pins the decision that only the
// sidebar panes margin-scroll: the Picker's lists scroll the selection just
// onto the last row.
func TestPickerListsKeepEdgePinnedScrolling(t *testing.T) {
	render := func(selected int) string {
		return windowedView(10, 4, selected, 20, func(idx int) string { return fmt.Sprintf("row%02d", idx) })
	}
	if got := strings.SplitN(render(3), "\n", 2)[0]; got != "row00" {
		t.Errorf("selection on last visible row: first row = %q, want row00", got)
	}
	if got := strings.SplitN(render(4), "\n", 2)[0]; got != "row01" {
		t.Errorf("selection one past the window: first row = %q, want row01", got)
	}
}
