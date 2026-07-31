"""End-to-end checks of the real binary in a pty, re-verifying M1.1-M1.3 on Charm v2."""

import os
import sys
import tempfile

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from harness import *  # noqa: E402

# Run from the repo root (see `make check-pty`).
BINARY = "./bin/noggin"
FIXTURE = "cmd/noggin/testdata/fixture.md"
TREE = "cmd/noggin/testdata/tree"

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
    "Down":         b"j",       # keys.LineDown, second key: j
    "Up":           b"k",       # keys.LineUp, second key: k
    "Expand":       b"l",       # keys.Expand, second key: l
    "Collapse":     b"h",       # keys.Collapse, second key: h
    "Activate":     b"\r",      # keys.Activate: enter
    "Star":         b"s",       # keys.Star
    "Edit":         b"e",       # keys.Edit
}


def config_env(home_dir):
    """Env vars and the resulting starred-roots store path that make Go's
    os.UserConfigDir() resolve under home_dir, mirroring the per-OS logic in
    os/file.go: darwin ignores $XDG_CONFIG_HOME and uses $HOME/Library/
    Application Support; other unix honours $XDG_CONFIG_HOME. Setting both
    env vars works on either OS the harness might run under.
    """
    if sys.platform == "darwin":
        cfg_dir = os.path.join(home_dir, "Library", "Application Support")
    else:
        cfg_dir = os.path.join(home_dir, ".config")
    return {"HOME": home_dir, "XDG_CONFIG_HOME": cfg_dir}, os.path.join(cfg_dir, "noggin", "roots")

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

# --- M2.2: the Files Nav pane browses the active root ---------------------

# The fixture tree sorts dirs-first then by name, so the Files pane opens on
# docs, empty, README.md, notes.txt with every folder collapsed; docs/guide.md
# stays hidden until docs is opened, and the content-free empty/ still lists.
s = Session([BINARY, TREE], 80, 24)
out = s.pump(1.5)
body = "\n".join(screen_of(out, 80, 24).raw_lines())
check("M2.2 Files pane opens showing top-level entries",
      all(name in body for name in ["docs", "empty", "README.md", "notes.txt"]),
      repr(body[:200]))
check("M2.2 empty folder is listed", "empty" in body)
check("M2.2 nested file hidden while its folder is collapsed",
      "guide.md" not in body)

# Focus starts in the Files pane, so its keys act without tabbing. Selection
# opens on docs; expand reveals its children in place.
s.send(KEY["Expand"], 0.5)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M2.2 expanding a folder reveals its children in place",
      "guide.md" in body, repr(body[:200]))

# Collapsing the same folder hides them again.
s.send(KEY["Collapse"], 0.5)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M2.2 collapsing a folder hides its children again",
      "guide.md" not in body, repr(body[:200]))

# Selection moves across the visible set: re-open docs, step down onto a child,
# and the reverse-video selection bar lands on it.
s.send(KEY["Expand"], 0.4)
s.send(KEY["Down"], 0.4)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M2.2 selection moves across the expanded set (guide.md still shown)",
      "guide.md" in body)
check("M2.2 reverse-video selection bar on the wire", b"\x1b[7m" in s.out)

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M2.2 exits 0 while browsing a folder", code == 0, f"exit {code}")

# --- M2.3: activating a file opens it in the viewer -----------------------

# Selection opens on docs (a folder); step down onto README.md, a file, then
# activate it.
s = Session([BINARY, TREE], 80, 24)
s.pump(1.5)
for _ in range(2):
    s.send(KEY["Down"], 0.4)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
footer_before = screen_of(s.out, 80, 24).lines()[-1]
check("M2.3 selection reaches README.md before activating",
      "README.md" in body and footer_before.split(" ·")[0] == "Files",
      repr(footer_before))

s.send(KEY["Activate"], 0.6)
sc = screen_of(s.out, 80, 24)
footer = sc.lines()[-1]
body = "\n".join(sc.raw_lines())
check("M2.3 activating a file moves focus to the Document pane",
      footer.split(" ·")[0] == "Document", repr(footer))
check("M2.3 activated file's content is rendered",
      "Root Readme" in body, repr(body[:200]))
check("M2.3 pane title names the activated file",
      "README.md" in body)

# Scroll keys now act on the viewer, not the Files Nav: half-page-down must
# not move a Files Nav selection bar (there is none to move) and must not
# error out; the exit below is the check that the session is still healthy.
s.send(KEY["HalfPageDown"], 0.5)

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M2.3 exits 0 after activating a file", code == 0, f"exit {code}")

# --- M3.3: the Roots Nav pane lists, marks, activates and stars roots -----

# Two roots sharing a base name ("project") under different parents, plus a
# starred path that is never created (the broken root). All three are
# pre-starred so a bare launch shows them without any launch argument, per
# S1 ("noggin -> starred roots only; the first is active").
roots_base = tempfile.mkdtemp()
root_admin = os.path.join(roots_base, "work", "admin", "project")
root_other = os.path.join(roots_base, "other", "project")
root_ghost = os.path.join(roots_base, "vault", "ghost")  # deliberately not created
os.makedirs(root_admin)
os.makedirs(root_other)

home_dir = tempfile.mkdtemp()
env, store_path = config_env(home_dir)
os.makedirs(os.path.dirname(store_path))
with open(store_path, "w") as f:
    f.write(root_admin + "\n" + root_other + "\n" + root_ghost + "\n")

s = Session([BINARY], 80, 24, env=env)
out = s.pump(1.5)
body = "\n".join(screen_of(out, 80, 24).raw_lines())

check("M3.3 Roots pane lists all three starred roots",
      body.count("project") == 2 and "ghost" in body, repr(body[:400]))
check("M3.3 same-named roots distinguished by their parent path",
      "admin" in body and "other" in body, repr(body[:400]))
check("M3.3 a root whose path does not exist is marked broken",
      "missing" in body, repr(body[:400]))

# The first starred root (root_admin) resolved and walked cleanly, so it is
# active and focus opens on the Files pane, same as a directory launch.
footer = screen_of(s.out, 80, 24).lines()[-1]
check("M3.3 bare launch with starred roots opens focused on Files",
      footer.split(" \u00b7")[0] == "Files", repr(footer))

# Tab back to the Roots pane before its own keys act.
s.send(KEY["PrevPane"], 0.4)
footer = screen_of(s.out, 80, 24).lines()[-1]
check("M3.3 shift+tab from Files moves focus to Roots",
      footer.split(" \u00b7")[0] == "Roots", repr(footer))


def row_with(text, lines):
    for line in lines:
        if text in line:
            return line
    return ""


lines = screen_of(s.out, 80, 24).raw_lines()
admin_row = row_with("admin", lines)
check("M3.3 active root is marked", "\u25cf" in admin_row, repr(admin_row))
check("M3.3 starred roots are marked", "\u2605" in admin_row, repr(admin_row))

# Star root_other (selection starts on root_admin; one Down reaches it),
# unstarring it, and check the change is written through to the store file
# (M3.2's toggleStore) as well as redrawn.
s.send(KEY["Down"], 0.4)
lines = screen_of(s.out, 80, 24).raw_lines()
other_row_before = row_with("other", lines)
check("M3.3 selection reaches the second root (other/project)",
      "\u2605" in other_row_before, repr(other_row_before))

s.send(KEY["Star"], 0.5)
lines = screen_of(s.out, 80, 24).raw_lines()
other_row_after = row_with("other", lines)
check("M3.3 star key un-marks the selected root on screen",
      "\u2605" not in other_row_after, repr(other_row_after))

with open(store_path) as f:
    stored = f.read()
check("M3.3 unstarring persists: root_other dropped from the store file",
      root_other not in stored and root_admin in stored and root_ghost in stored,
      repr(stored))

# Move onto the broken root and try to activate it: refused, with a footer
# message naming the missing path, and the active root and Files pane both
# stay as they were.
s.send(KEY["Down"], 0.4)
s.send(KEY["Activate"], 0.6)
sc = screen_of(s.out, 80, 24)
footer = sc.lines()[-1]
lines = sc.raw_lines()
check("M3.3 activating a broken root is refused with a footer message",
      "ghost" in footer and "not activated" in footer, repr(footer))
admin_row = row_with("admin", lines)
check("M3.3 the active root did not change after the refusal",
      "\u25cf" in admin_row, repr(admin_row))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M3.3 exits 0 after browsing and starring roots", code == 0, f"exit {code}")

# --- M4.1: edit hands the active document to $EDITOR and resumes -----------

# A fake $EDITOR that appends a marker line to whatever file it is given,
# standing in for a real editor session without needing one interactively.
edit_dir = tempfile.mkdtemp()
fake_editor = os.path.join(edit_dir, "fake-editor.sh")
with open(fake_editor, "w") as f:
    f.write('#!/bin/sh\necho "Appended by fake editor" >> "$1"\n')
os.chmod(fake_editor, 0o755)

# Short on purpose: the appended line must land within the viewport without
# needing a scroll, since re-opening a document starts at the top.
edit_doc = os.path.join(edit_dir, "doc.md")
with open(edit_doc, "w") as f:
    f.write("# Edit Handoff Fixture\n\nOriginal line.\n")

s = Session([BINARY, edit_doc], 80, 24, env={"EDITOR": fake_editor})
out = s.pump(1.5)
body = "\n".join(screen_of(out, 80, 24).raw_lines())
check("M4.1 document opens before editing",
      "doc.md" in body and "Edit Handoff Fixture" in body, repr(body[:200]))

s.send(KEY["Edit"], 1.5)
sc = screen_of(s.out, 80, 24)
lines = sc.raw_lines()
body = "\n".join(lines)
check("M4.1 noggin resumes with an intact frame after the editor exits",
      len(lines) == 24 and all(len(l) == 80 for l in lines) and
      lines[0].startswith("╭") and lines[-2].startswith("╰"),
      f"{len(lines)} rows")
check("M4.1 the appended line is visible on return, re-rendered from disk",
      "Appended by fake editor" in body, repr(body[-200:]))

with open(edit_doc) as f:
    check("M4.1 the editor's write reached the real file on disk",
          "Appended by fake editor" in f.read())

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M4.1 exits 0 after an edit round-trip", code == 0, f"exit {code}")

# With no document open, the key reports that instead of doing nothing.
s = Session([BINARY], 80, 24, env={"EDITOR": fake_editor})
s.pump(1.2)
s.send(KEY["Edit"], 0.6)
footer = screen_of(s.out, 80, 24).lines()[-1]
check("M4.1 editing with no document open reports it in the footer",
      "no document open" in footer.lower(), repr(footer))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M4.1 exits 0 after the no-document-open report", code == 0, f"exit {code}")

done()
