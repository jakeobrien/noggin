---
title: 'Sidebar panes: scroll with a one-third margin instead of pinning the cursor to the edge'
status: open
priority: 2
issue_type: task
labels:
- needs-triage
created_at: 2026-10-14T15:22:58.226397454+00:00
updated_at: 2026-10-14T15:22:58.226397454+00:00
---

# Description

## Problem

In the Navigator sidebar panes (Sources Nav and Files Nav), once the list is taller than the pane, scrolling is **edge-pinned**:

- Scrolling **down** from the top: nothing moves until the cursor reaches the last visible row, then the view scrolls one row per keypress with the cursor stuck at the bottom edge.
- Scrolling **up** after reaching the bottom: the cursor immediately sits at the bottom edge, so the view starts scrolling on the very first Up keypress, and the cursor stays pinned to the bottom the whole way.

This is because `windowedViewRows` (`cmd/noggin/nav.go`) is stateless: the window offset is derived purely from `selected` (`offset = selected - perScreen + 1`). There is no remembered scroll position, so the view can't "stay put" while the cursor moves inside it.

## Desired behavior

The view keeps a persistent scroll offset and scrolls only when the cursor enters a margin zone of about **one third of the pane height** at either edge.

1. **Scrolling down from the top**: the cursor moves freely down through the first ~two thirds of the pane. Once it passes the two-thirds point (i.e. is within the bottom third), further Down presses scroll the view so the cursor stays at that two-thirds line, rather than waiting until it hits the bottom edge.
2. **Scrolling up from the bottom**: with the cursor at the bottom of the pane, pressing Up moves the cursor up *without scrolling* until it crosses into the top third (the margin line one third from the top). Past it, the view scrolls to keep the cursor at that line. The view must not keep the selection pinned to the bottom.
3. Symmetrically, scrolling down after having scrolled up behaves the same way mirrored (no scroll until the cursor reaches the bottom-third margin line).
4. Near the ends of the list the view clamps (no blank space past the first/last row; the cursor can reach the very first and very last row).
5. The cursor is always fully on screen. Jumps (e.g. jump-to-top/bottom, search/reveal, selection restored after a store write, mode or scope switch, resize) must still leave the selection visible, adjusting the offset minimally.

## Notes for implementation

- Needs a persisted scroll offset on the nav state (`filesNav`, `rootsNav`) that's updated on selection change and on resize/row-count change, then passed to the renderer. Offset must be re-clamped when rows shrink (`clampSelected` paths in `files.go`).
- Margin is a fraction of the pane's height in *entries* (`height / rowHeight`), not lines; Sources Nav rows are two lines tall (`windowedViewRows` `rowHeight`). With small panes the margin should degrade gracefully (e.g. margin = perScreen/3, floored, and never so large the two margins overlap).
- `windowedView` is also used by the Picker (`picker_view.go`: results/grep/files lists). Decide whether the Picker shares the new behavior or keeps the old one; the ticket's scope is the sidebar panes.
- Does not change the Document Viewer's own scrolling.

## Acceptance criteria

- [ ] With a list taller than the pane and the cursor at the top, pressing Down does not scroll the view until the cursor passes the two-thirds line; after that, the view scrolls and the cursor stays on that line.
- [ ] With the view scrolled to the bottom and the cursor on the last row, pressing Up moves the cursor up with the view unchanged until the cursor reaches the one-third-from-top line; then the view scrolls, keeping the cursor on that line.
- [ ] Cursor reaches the first and last rows with no blank padding rows past the list ends when the list is taller than the pane.
- [ ] Works in Files Nav (Tree, Recent, Starred) and Sources Nav (two-line rows).
- [ ] Selection stays visible after resize, row-set shrink, mode/scope switch, and jumps.
- [ ] Short lists (fit in the pane) are unaffected.
- [ ] Unit tests cover the offset-update function (down, up, clamping at both ends, tiny panes, rowHeight 2); update existing `windowedView` tests that assume edge-pinning.
