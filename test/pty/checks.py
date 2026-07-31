"""End-to-end checks of the real binary in a pty, re-verifying M1.1-M1.3 on Charm v2."""

import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from harness import *  # noqa: E402

# Run from the repo root (see `make check-pty`).
BINARY = "./bin/noggin"
FIXTURE = "cmd/noggin/testdata/fixture.md"

# D2 keeps key literals inside cmd/noggin/keymap.go, but a pty harness has to
# put real bytes on the wire, and it is on the far side of a language boundary
# from the keymap. This table is that carve-out, kept in one place and named
# after the binding each entry stands for, so retuning a binding is still a
# two-file edit that greps. The Go side has no key literals; these are the
# terminal's encoding of them, which is a different fact.
KEY = {
    "Quit":         b"q",       # keys.Quit
    "QuitCtrlC":    b"\x03",    # keys.Quit, second key: ctrl+c
    "NextPane":     b"\t",      # keys.NextPane: tab
    "PrevPane":     b"\x1b[Z",  # keys.PrevPane: shift+tab
    "HalfPageDown": b"\x04",    # keys.HalfPageDown: ctrl+d
    "PageUp":       b"\x02",    # keys.PageUp, second key: ctrl+b
    "Bottom":       b"G",       # keys.Bottom
}

# --- M1.1: starts, quits, restores the terminal ---------------------------

for quit_key, label in [(KEY["Quit"], "q"), (KEY["QuitCtrlC"], "ctrl+c")]:
    s = Session([BINARY], 80, 24)
    out = s.pump(1.2)
    check(f"M1.1 first frame drawn before any input ({label} run)", len(strip(out)) > 200,
          f"{len(out)} bytes")
    out = s.send(quit_key, 0.8)
    code = s.close()
    check(f"M1.1 exits 0 on {label}", code == 0, f"exit {code}")
    check(f"M1.1 cursor restored after {label}", b"\x1b[?25h" in out)
    check(f"M1.1 bracketed paste disabled after {label}", b"\x1b[?2004l" in out)

# --- M1.2: the frame, at two sizes, with focus and reflow -----------------

for cols, rows in [(80, 24), (100, 30)]:
    s = Session([BINARY], cols, rows)
    out = s.pump(1.2)
    sc = screen_of(out, cols, rows)
    lines = sc.raw_lines()

    check(f"M1.2 frame is exactly {cols}x{rows}",
          len(lines) == rows and all(len(l) == cols for l in lines),
          f"{len(lines)} rows, widths {sorted(set(len(l) for l in lines))}")

    body = "\n".join(lines)
    check(f"M1.2 all three regions present at {cols}x{rows}",
          "Roots" in body and "Files" in body and "Document" in body)
    check(f"M1.2 footer is the last row at {cols}x{rows}",
          "quit" in lines[-1], repr(lines[-1][:40]))

    # Navigator column ends where the viewer begins: two box edges on row 0.
    nav_edge = lines[0].index("╮")
    check(f"M1.2 navigator column is a distinct region at {cols}x{rows}",
          0 < nav_edge < cols - 1 and lines[0][nav_edge + 1] == "╭", f"edge at {nav_edge}")

    # Real styling: focused and unfocused border colours both on the wire.
    check(f"M1.2 focused border colour drawn at {cols}x{rows}", b"38;5;205" in out)
    check(f"M1.2 unfocused border colour drawn at {cols}x{rows}", b"38;5;240" in out)

    # Focus moves, and the footer names where it is. Reading focus off the
    # footer is what makes this a real check rather than "something redrew":
    # it also exercises bubbletea v2's own decoding of tab and shift+tab, which
    # the synthetic keyPress() helper in keymap.go cannot vouch for.
    def focused():
        return screen_of(s.out, cols, rows).lines()[-1].split(" ·")[0]

    check(f"M1.2 focus starts in the Files pane at {cols}x{rows}",
          focused() == "Files", repr(focused()))

    forward = []
    for _ in range(3):
        s.send(KEY["NextPane"], 0.5)
        forward.append(focused())
    check(f"M1.2 tab cycles focus through all three panes at {cols}x{rows}",
          forward == ["Document", "Roots", "Files"], f"{forward}")

    backward = []
    for _ in range(3):
        s.send(KEY["PrevPane"], 0.5)
        backward.append(focused())
    check(f"M1.2 shift+tab cycles focus back at {cols}x{rows}",
          backward == ["Roots", "Document", "Files"], f"{backward}")

    s.send(KEY["Quit"], 0.6)
    s.close()

# reflow on resize
s = Session([BINARY], 80, 24)
s.pump(1.2)
mark = len(s.out)
s.resize(100, 30, 1.0)
sc = screen_of(s.out, 100, 30)
lines = sc.raw_lines()
check("M1.2 reflows to 100x30 on resize",
      len(lines) == 30 and all(len(l) == 100 for l in lines),
      f"{len(lines)} rows, widths {sorted(set(len(l) for l in lines))}")
check("M1.2 resize produced a redraw", len(s.out) > mark)
s.send(KEY["Quit"], 0.6)
s.close()

# --- M1.3: the document renders ------------------------------------------

FRONTMATTER = ["title: Fixture Document", "tags: [noggin, fixture]", "draft: true"]

for cols, rows in [(80, 24), (100, 30)]:
    s = Session([BINARY, FIXTURE], cols, rows)
    out = s.pump(1.5)
    sc = screen_of(out, cols, rows)
    lines = sc.raw_lines()

    check(f"M1.3 frame is exactly {cols}x{rows} with a document open",
          len(lines) == rows and all(len(l) == cols for l in lines),
          f"{len(lines)} rows, widths {sorted(set(len(l) for l in lines))}")
    check(f"M1.3 pane title names the open file at {cols}x{rows}",
          "fixture.md" in "\n".join(lines))
    check(f"M1.3 document heading on screen at {cols}x{rows}",
          "Reading Notes" in "\n".join(lines))
    for fm in FRONTMATTER:
        check(f"M1.3 frontmatter {fm!r} absent at {cols}x{rows}",
              fm not in "\n".join(lines))
    check(f"M1.3 no literal tab reaches the screen at {cols}x{rows}",
          "\t" not in "".join(lines))
    check(f"M1.3 markdown is styled (SGR colour on the wire) at {cols}x{rows}",
          b"\x1b[38;" in out or b"\x1b[1m" in out)

    # ctrl+d (half page) repeatedly must reach the code fence, every row intact.
    reached_fence, reached_table = False, False
    for _ in range(30):
        s.send(KEY["HalfPageDown"], 0.35)
        sc = screen_of(s.out, cols, rows)
        lines = sc.raw_lines()
        bad = [(i, len(l)) for i, l in enumerate(lines) if len(l) != cols]
        if bad:
            check(f"M1.3 rows stay {cols} cells while scrolling at {cols}x{rows}",
                  False, f"bad rows {bad[:3]}")
            break
        # A frame taller than the terminal scrolls the screen, taking the top
        # border with it — the visible symptom of content that measures short.
        if not (lines[0].startswith("╭") and lines[0].endswith("╮")
                and lines[-2].startswith("╰") and lines[-2].endswith("╯")):
            check(f"M1.3 frame border intact while scrolling at {cols}x{rows}",
                  False, f"top {lines[0][:3]!r} bottom {lines[-2][:3]!r}{lines[-2][-3:]!r}")
            break
        body = "\n".join(lines)
        if "os.ReadFile" in body:
            reached_fence = True
        if "─┼─" in body or "├" in body or "Shown in the pane title" in body:
            reached_table = True
    else:
        check(f"M1.3 rows stay {cols} cells while scrolling at {cols}x{rows}", True)
        check(f"M1.3 frame border intact while scrolling at {cols}x{rows}", True)
    check(f"M1.3 half-page scrolling reaches the code fence at {cols}x{rows}", reached_fence)
    check(f"M1.3 box-drawing table rendered at {cols}x{rows}", reached_table)

    # G must reach the document's last line, and ctrl+b must come back off it.
    s.send(KEY["Bottom"], 0.6)
    sc = screen_of(s.out, cols, rows)
    tail = "The final line of the fixture document."
    check(f"M1.3 G reaches the last line at {cols}x{rows}",
          tail in "\n".join(sc.raw_lines()))
    s.send(KEY["PageUp"], 0.6)
    sc = screen_of(s.out, cols, rows)
    check(f"M1.3 ctrl+b pages back off the last line at {cols}x{rows}",
          tail not in "\n".join(sc.raw_lines()))

    code = (s.send(KEY["Quit"], 0.6), s.close())[1]
    check(f"M1.3 exits 0 with a document open at {cols}x{rows}", code == 0, f"exit {code}")

# empty state on bare launch
s = Session([BINARY], 80, 24)
out = s.pump(1.2)
sc = screen_of(out, 80, 24)
check("M1.3 empty state on bare launch",
      "No document open" in "\n".join(sc.raw_lines()))
s.send(KEY["Quit"], 0.6)
s.close()

done()
