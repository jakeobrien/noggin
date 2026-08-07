"""End-to-end checks of the real binary in a pty, re-verifying M1.1-M1.3 on Charm v2."""

import os
import re
import subprocess
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
    "Find":         b"/",       # keys.Find
    "FindNext":     b"n",       # keys.FindNext
    "FindPrevious": b"N",       # keys.FindPrevious
    "Navigator":    b"\x0e",    # keys.Navigator: ctrl+n
    "Cancel":       b"\x1b",    # keys.Cancel: esc
    "Help":         b"?",       # keys.Help
    "Refresh":      b"R",       # keys.Refresh
    "FocusFiles":   b"f",       # keys.FocusFiles
    "Peek":         b"p",       # keys.Peek
    "Picker":       b" ",       # keys.Picker: space
    "PickerUp":     b"\x1b[A",  # keys.PickerUp: up arrow
    "PickerDown":   b"\x1b[B",  # keys.PickerDown: down arrow
    "PickerMode":   b"\t",      # keys.PickerMode: tab
    "PickerScope":  b"\x1b[Z",  # keys.PickerScope: shift+tab
    "ToggleHidden": b"\x1bh",   # keys.ToggleHidden: alt+h (ESC h on the wire)
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
#
# Height 30, not the usual 24: S5's two-line rows halve how many entries fit
# in the Roots pane's fixed share of the Navigator's height; 24 rows only
# leaves room for 2 of these 3 fixture roots on screen at once (an instance
# of S5's own accepted "fewer visible roots" tradeoff — real, but not what
# this fixture is testing), so this block alone uses a taller terminal to
# keep all three visible without changing what it asserts.
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

s = Session([BINARY], 80, 30, env=env)
out = s.pump(1.5)
body = "\n".join(screen_of(out, 80, 30).raw_lines())

check("M3.3 Roots pane lists all three starred roots",
      # S5's two-line rows repeat a "project" root's base name on line 1
      # and again as the leaf of its full path on line 2 (4, not 2, for
      # the two same-named roots) - "admin"/"other" below is what actually
      # pins there being two distinct entries, not this count.
      body.count("project") == 4 and "ghost" in body, repr(body[:400]))
check("M3.3 same-named roots distinguished by their parent path",
      "admin" in body and "other" in body, repr(body[:400]))
check("M3.3 a root whose path does not exist is marked broken",
      "missing" in body, repr(body[:400]))

# The first starred root (root_admin) resolved and walked cleanly, so it is
# active and focus opens on the Files pane, same as a directory launch.
footer = screen_of(s.out, 80, 30).lines()[-1]
check("M3.3 bare launch with starred roots opens focused on Files",
      footer.split(" \u00b7")[0] == "Files", repr(footer))

# Tab back to the Roots pane before its own keys act.
s.send(KEY["PrevPane"], 0.4)
footer = screen_of(s.out, 80, 30).lines()[-1]
check("M3.3 shift+tab from Files moves focus to Roots",
      footer.split(" \u00b7")[0] == "Roots", repr(footer))


def row_with(text, lines):
    for line in lines:
        if text in line:
            return line
    return ""


def row_pair_with(text, lines):
    # S5's Roots Nav entries are two physical lines each (name, path); a
    # parent-path substring used to distinguish same-named roots (e.g.
    # "admin"/"other") only matches the path line, whose own name line -
    # carrying the active/starred markers S5 keeps there - is always the
    # line immediately above it (never below: S5 never repeats a marker on
    # line 2, and a path line's own content always starts with "/" or the
    # left-truncation ellipsis, never a marker). Pairing with the line below
    # instead would risk pulling in the *next* entry's own markers (e.g. a
    # neighbouring starred root), a false positive an "is this marker
    # present" check must not risk.
    for i, line in enumerate(lines):
        if text not in line:
            continue
        content = line.strip("\u2502").strip()
        looks_like_path = content.startswith("\u2026") or content.startswith("/")
        if looks_like_path and i > 0:
            return lines[i - 1] + "\n" + line
        return line
    return ""


lines = screen_of(s.out, 80, 30).raw_lines()
admin_row = row_pair_with("admin", lines)
check("M3.3 active root is marked", "\u25cf" in admin_row, repr(admin_row))
check("M3.3 starred roots are marked", "\u2605" in admin_row, repr(admin_row))

# Star root_other (selection starts on root_admin; one Down reaches it),
# unstarring it, and check the change is written through to the store file
# (M3.2's toggleStore) as well as redrawn.
s.send(KEY["Down"], 0.4)
lines = screen_of(s.out, 80, 30).raw_lines()
other_row_before = row_pair_with("other", lines)
check("M3.3 selection reaches the second root (other/project)",
      "\u2605" in other_row_before, repr(other_row_before))

s.send(KEY["Star"], 0.5)
lines = screen_of(s.out, 80, 30).raw_lines()
other_row_after = row_pair_with("other", lines)
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
sc = screen_of(s.out, 80, 30)
footer = sc.lines()[-1]
lines = sc.raw_lines()
check("M3.3 activating a broken root is refused with a footer message",
      "ghost" in footer and "not activated" in footer, repr(footer))
admin_row = row_pair_with("admin", lines)
check("M3.3 the active root did not change after the refusal",
      "\u25cf" in admin_row, repr(admin_row))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M3.3 exits 0 after browsing and starring roots", code == 0, f"exit {code}")

# --- M3.2 (T27): a star survives a restart, across two real OS processes -

# Hardens what was, through M3.2, an ad-hoc script: star an ad-hoc root in one
# process, exit, then start a second, independent process against the same
# isolated $HOME/$XDG_CONFIG_HOME and confirm the star is still there. Two
# real OS processes, not two Sessions against one running binary, is the
# point — it is the only way to actually exercise "survives a restart".
persist_home = tempfile.mkdtemp()
persist_env, persist_store = config_env(persist_home)
persist_target = tempfile.mkdtemp()
persist_name = os.path.basename(persist_target)

s = Session([BINARY, persist_target], 80, 24, env=persist_env)
s.pump(1.5)
s.send(KEY["PrevPane"], 0.4)  # Files -> Roots (shift+tab)
footer = screen_of(s.out, 80, 24).lines()[-1]
check("M3.2 restart fixture: focus reaches Roots to star the ad-hoc root",
      footer.split(" \u00b7")[0] == "Roots", repr(footer))

s.send(KEY["Star"], 0.5)
lines = screen_of(s.out, 80, 24).raw_lines()
row = row_with(persist_name, lines)
check("M3.2 restart fixture: star mark shows immediately", "\u2605" in row, repr(row))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M3.2 restart fixture: first process exits 0", code == 0, f"exit {code}")

with open(persist_store) as f:
    stored = f.read()
check("M3.2 star written through to the store file by the first process",
      persist_target in stored, repr(stored))

s2 = Session([BINARY], 80, 24, env=persist_env)
out2 = s2.pump(1.5)
body2 = "\n".join(screen_of(out2, 80, 24).raw_lines())
check("M3.2 star survives a restart across two separate OS processes",
      persist_name in body2 and "\u2605" in body2, repr(body2[:300]))

code2 = (s2.send(KEY["Quit"], 0.6), s2.close())[1]
check("M3.2 restart fixture: second process exits 0", code2 == 0, f"exit {code2}")

# --- M3.4 (T27): activating a different root leaves the open document -----
# unchanged, end to end (byte-diff of the viewer pane against the real binary)

rootb_base = tempfile.mkdtemp()
rootb = os.path.join(rootb_base, "root-b")
os.makedirs(rootb)
with open(os.path.join(rootb, "other.md"), "w") as f:
    f.write("# Root B\n\nDifferent content.\n")

switch_home = tempfile.mkdtemp()
switch_env, switch_store = config_env(switch_home)
os.makedirs(os.path.dirname(switch_store))
with open(switch_store, "w") as f:
    f.write(rootb + "\n")

s = Session([BINARY, TREE], 80, 24, env=switch_env)
s.pump(1.5)
for _ in range(2):
    s.send(KEY["Down"], 0.4)
s.send(KEY["Activate"], 0.6)
lines_before = screen_of(s.out, 80, 24).raw_lines()
check("M3.4 pty: README.md opens in the viewer before switching roots",
      "Root Readme" in "\n".join(lines_before), repr(lines_before[:1]))

# Byte-diff only the viewer's columns (right of the Navigator column's box
# edge), excluding the footer row, so the check is not tripped up by the
# Files Nav or footer, which ARE expected to change when the active root
# switches — M4.2 made the footer's focus-specific hint text long enough to
# reach past nav_edge (it used to end well short of that column, so slicing
# happened to land in its trailing padding regardless of focus; that is no
# longer a safe assumption, hence dropping the footer row outright).
nav_edge = lines_before[0].index("\u256e")  # box edge ╮ closing the nav column
viewer_before = [line[nav_edge + 1:] for line in lines_before[:-1]]

s.send(KEY["NextPane"], 0.4)  # Document -> Roots (tab)
footer = screen_of(s.out, 80, 24).lines()[-1]
check("M3.4 pty: tab from the Document pane reaches Roots",
      footer.split(" \u00b7")[0] == "Roots", repr(footer))

s.send(KEY["Down"], 0.4)   # selection off the ad-hoc TREE root, onto root-b
s.send(KEY["Activate"], 0.6)
lines_after = screen_of(s.out, 80, 24).raw_lines()
viewer_after = [line[nav_edge + 1:] for line in lines_after[:-1]]

check("M3.4 pty: viewer pane is byte-identical after activating a different root",
      viewer_before == viewer_after,
      f"first mismatch at row {next((i for i, (a, b) in enumerate(zip(viewer_before, viewer_after)) if a != b), None)}")
body_after = "\n".join(lines_after)
check("M3.4 pty: Files Nav re-rooted to root-b's entries",
      "other.md" in body_after and "docs" not in body_after, repr(body_after[:200]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M3.4 pty: exits 0 after switching roots with a document open", code == 0, f"exit {code}")

# --- C2a/M4.1: per-file git status markers in the Files Nav --------------
# (Named M4.1 in this cycle, C2a; the pre-existing "M4.1" label just below
# belongs to an earlier cycle's edit-round-trip milestone. checks.py reuses
# M/S keys across cycles as plain descriptive labels, not durable ids.)

def run_git(cwd, *args):
    subprocess.run(["git", *args], cwd=cwd, check=True,
                    stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                    env={**os.environ,
                         "GIT_AUTHOR_NAME": "noggin-pty", "GIT_AUTHOR_EMAIL": "noggin-pty@example.com",
                         "GIT_COMMITTER_NAME": "noggin-pty", "GIT_COMMITTER_EMAIL": "noggin-pty@example.com"})

git_repo = tempfile.mkdtemp()
run_git(git_repo, "init", "-q")
with open(os.path.join(git_repo, "tracked.md"), "w") as f:
    f.write("one\n")
with open(os.path.join(git_repo, "clean.md"), "w") as f:
    f.write("stays clean\n")
run_git(git_repo, "add", "-A")
run_git(git_repo, "commit", "-q", "-m", "init")
with open(os.path.join(git_repo, "tracked.md"), "w") as f:
    f.write("one\nmodified\n")
with open(os.path.join(git_repo, "new.md"), "w") as f:
    f.write("new\n")

s = Session([BINARY, git_repo], 80, 24)
s.pump(1.5)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M4.1 pty: a modified tracked file is marked M in the Files Nav",
      "M tracked.md" in body, repr(body[:300]))
check("C2a/M4.1 pty: a new untracked file is marked A in the Files Nav",
      "A new.md" in body, repr(body[:300]))
check("C2a/M4.1 pty: an unchanged tracked file carries no marker",
      "clean.md" in body and "M clean.md" not in body and "A clean.md" not in body,
      repr(body[:300]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("C2a/M4.1 pty: exits 0 with a git-status-marked Root open", code == 0, f"exit {code}")

# A Root outside any git repository (a bare tempdir, never git-init'd) shows
# no markers and no error — the session still runs normally.
plain_root = tempfile.mkdtemp()
with open(os.path.join(plain_root, "note.md"), "w") as f:
    f.write("just a file\n")

s = Session([BINARY, plain_root], 80, 24)
s.pump(1.5)
sc = screen_of(s.out, 80, 24)
body = "\n".join(sc.raw_lines())
footer = sc.lines()[-1]
check("C2a/M4.1 pty: a non-repo Root shows its file with no git status marker",
      "note.md" in body and "M note.md" not in body and "A note.md" not in body,
      repr(body[:300]))
check("C2a/M4.1 pty: a non-repo Root raises no footer error",
      "error" not in footer.lower() and "fatal" not in footer.lower(), repr(footer))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("C2a/M4.1 pty: exits 0 with a non-repo Root open", code == 0, f"exit {code}")

# --- C2a/M5.1: manual refresh (R) ------------------------------------------
# (Named M5.1 in this cycle, C2a; a pre-existing "M5.1" label further below
# belongs to an earlier cycle's find-in-document milestone, same reused-label
# convention noted at the git-status section above.)

refresh_root = tempfile.mkdtemp()
with open(os.path.join(refresh_root, "keep.md"), "w") as f:
    f.write("kept\n")
doomed_path = os.path.join(refresh_root, "doomed.md")
with open(doomed_path, "w") as f:
    f.write("about to vanish\n")

s = Session([BINARY, refresh_root], 80, 24)
s.pump(1.2)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M5.1 pty: doomed.md listed before the refresh", "doomed.md" in body, repr(body[:300]))

# Outside noggin: doomed.md is deleted, new.md is created.
os.remove(doomed_path)
with open(os.path.join(refresh_root, "new.md"), "w") as f:
    f.write("just created\n")

s.send(KEY["Refresh"], 0.8)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M5.1 pty: new.md appears after R without relaunching", "new.md" in body, repr(body[:300]))
check("C2a/M5.1 pty: doomed.md is gone after R", "doomed.md" not in body, repr(body[:300]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("C2a/M5.1 pty: exits 0 after a refresh", code == 0, f"exit {code}")

# A document open at a mid-scroll position keeps that position across a
# refresh that only grows the file underneath it (S2's scroll-preserving
# half) — appending never shrinks past the current offset, so this checks
# preservation, not the separate clamp-to-end case go test already covers
# (viewer_test.go / refresh_test.go own that unit-level clamp check).
big_doc = os.path.join(tempfile.mkdtemp(), "big.md")
with open(big_doc, "w") as f:
    f.write("".join(f"- item {i}\n" for i in range(1, 61)))

s = Session([BINARY, big_doc], 80, 24)
s.pump(1.2)
top_frame = "\n".join(screen_of(s.out, 80, 24).raw_lines())
s.send(KEY["HalfPageDown"], 0.4)
before = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M5.1 pty: scrolled away from the top before refreshing",
      before != top_frame, repr(before[:200]))

with open(big_doc, "a") as f:
    f.write("".join(f"- item {i}\n" for i in range(61, 121)))
s.send(KEY["Refresh"], 0.8)
after = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M5.1 pty: scroll position unchanged by a refresh that only grew the document",
      after == before, f"before={before[:200]!r} after={after[:200]!r}")

s.send(KEY["Bottom"], 0.5)
bottom = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M5.1 pty: the grown content is actually reachable after refresh",
      "item 120" in bottom, repr(bottom[:300]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("C2a/M5.1 pty: exits 0 after a scroll-preserving refresh", code == 0, f"exit {code}")

# The Active Document vanishing under a refresh empties the Document Viewer
# rather than showing stale content, and the session keeps running.
vanishing_doc = os.path.join(tempfile.mkdtemp(), "vanishing.md")
with open(vanishing_doc, "w") as f:
    f.write("# Present for now\n")

s = Session([BINARY, vanishing_doc], 80, 24)
s.pump(1.2)
os.remove(vanishing_doc)
s.send(KEY["Refresh"], 0.8)
sc = screen_of(s.out, 80, 24)
body = "\n".join(sc.raw_lines())
check("C2a/M5.1 pty: the Document Viewer empties when the Active Document vanishes",
      "No document open" in body, repr(body[:300]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("C2a/M5.1 pty: exits 0 after the Active Document vanished", code == 0, f"exit {code}")

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

# --- M4.2: footer key hints, truncation, and message clear-on-keypress -----

s = Session([BINARY, TREE], 80, 24)
s.pump(1.2)
footer_files = screen_of(s.out, 80, 24).lines()[-1]
check("M4.2 Files pane footer names its own actions",
      "expand" in footer_files and "collapse" in footer_files and "open" in footer_files,
      repr(footer_files))

s.send(KEY["NextPane"], 0.4)  # Files -> Document
footer_doc = screen_of(s.out, 80, 24).lines()[-1]
check("M4.2 hints change with focus (Document differs from Files)",
      footer_doc != footer_files and "page" in footer_doc, repr(footer_doc))

s.send(KEY["NextPane"], 0.4)  # Document -> Roots
footer_roots = screen_of(s.out, 80, 24).lines()[-1]
check("M4.2 hints change with focus (Roots differs from Files and Document)",
      footer_roots not in (footer_files, footer_doc) and "star" in footer_roots,
      repr(footer_roots))
check("M4.2 quit is always in the footer",
      all("quit" in f for f in (footer_files, footer_doc, footer_roots)))
check("M4.2 the help affordance is always in the footer",
      all("help" in f for f in (footer_files, footer_doc, footer_roots)))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M4.2 exits 0 after cycling focus", code == 0, f"exit {code}")

# A narrow terminal drops the Navigator column, landing focus in the
# Document pane; its footer truncates gracefully but still names the help
# affordance and quit (S11).
s = Session([BINARY], 30, 24)
s.pump(1.2)
narrow_footer = screen_of(s.out, 30, 24).lines()[-1]
check("M4.2 narrow terminal footer still names the help affordance and quit",
      "help" in narrow_footer and "quit" in narrow_footer, repr(narrow_footer))
check("M4.2 narrow terminal footer dropped the pane-specific hints",
      "page" not in narrow_footer and "top" not in narrow_footer, repr(narrow_footer))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M4.2 exits 0 at a narrow width", code == 0, f"exit {code}")

# A rejected action's footer message clears on the next keypress.
ghost_home = tempfile.mkdtemp()
ghost_env, ghost_store = config_env(ghost_home)
os.makedirs(os.path.dirname(ghost_store))
with open(ghost_store, "w") as f:
    f.write(os.path.join(tempfile.mkdtemp(), "ghost") + "\n")  # never created: broken

s = Session([BINARY], 80, 24, env=ghost_env)
s.pump(1.2)
s.send(KEY["NextPane"], 0.4)  # Files -> Document
s.send(KEY["NextPane"], 0.4)  # Document -> Roots
s.send(KEY["Activate"], 0.6)  # refused: the starred root is broken
footer_message = screen_of(s.out, 80, 24).lines()[-1]
check("M4.2 a rejected activation shows a footer message",
      "not activated" in footer_message, repr(footer_message))

s.send(KEY["Down"], 0.4)  # any keypress, itself unrejected
footer_after = screen_of(s.out, 80, 24).lines()[-1]
check("M4.2 the message clears on the next keypress",
      "not activated" not in footer_after, repr(footer_after))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M4.2 exits 0 after the message-clear check", code == 0, f"exit {code}")

# --- M5.1: find in document -------------------------------------------------

# This is the check that matters most for M5.1: under a real pty, Glamour
# renders the fixture with real ANSI colour (no-TTY styling is a go-test-only
# fallback), and it is exactly that styling which drifts viewport.Model's own
# SetHighlights out of alignment with the document's real line numbers
# (.charm/knowledge/viewport-search-ansi-offsets.md). go test alone cannot catch that
# — it never renders in colour — so this is the one place a regression would
# actually show up.

s = Session([BINARY, FIXTURE], 80, 24)
s.pump(1.2)
sc = screen_of(s.out, 80, 24)
body_before = "\n".join(sc.raw_lines())
check("M5.1 fixture opens at the top, tail not yet on screen",
      "Reading Notes" in body_before and "final line of the fixture" not in body_before,
      repr(body_before[:80]))

query = b"final line of the fixture"
s.send(KEY["Find"], 0.3)
out = s.send(query, 0.6)
sc = screen_of(s.out, 80, 24)
body = "\n".join(sc.raw_lines())
check("M5.1 an incremental query brings a tail phrase on screen",
      "final line of the fixture" in body, repr(body[-200:]))
check("M5.1 the matched line is indicated with reverse video",
      b"\x1b[7m" in out, repr(out[-400:]))
footer_editing = sc.lines()[-1]
check("M5.1 the footer shows the query while it is being typed",
      "final line of the fixture" in footer_editing, repr(footer_editing))

s.send(KEY["Activate"], 0.4)  # commit
footer_committed = screen_of(s.out, 80, 24).lines()[-1]
check("M5.1 committing the query restores the normal footer",
      "final line of the fixture" not in footer_committed, repr(footer_committed))

scrolled_body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
s.send(KEY["Cancel"], 0.4)
after_cancel = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M5.1 dismissing the find leaves the view where it was",
      after_cancel == scrolled_body, repr(after_cancel[:80]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M5.1 exits 0 after a find round-trip", code == 0, f"exit {code}")

# Traversal: a query with several hits, wrapping past the last one. The
# fixture has 4 hits for "line", so a full cycle of 4 nexts returns to the
# same match. Comparing against the frame right after commit would be
# comparing two different arrival paths to that match (typing the query
# converges on it incrementally, one keystroke at a time, each of which can
# itself move the viewport — EnsureVisible only scrolls when a match is not
# already on screen, so which frame you land on for a given match depends on
# how you got there); comparing one next's frame against five nexts' frame
# instead holds the arrival path fixed (always "press next") on both sides.
s = Session([BINARY, FIXTURE], 80, 24)
s.pump(1.2)
s.send(KEY["Find"], 0.3)
s.send(b"line", 0.4)
s.send(KEY["Activate"], 0.3)

# Two warm-up cycles settle the sequence of viewport positions into a fixed,
# repeating one (EnsureVisible only scrolls when a match is not already on
# screen, so the very first visit to each match can still be settling in);
# from there, one more full cycle must reproduce the same frame.
for _ in range(2 * 4):
    s.send(KEY["FindNext"], 0.2)
ref = "\n".join(screen_of(s.out, 80, 24).raw_lines())
for _ in range(4):
    s.send(KEY["FindNext"], 0.2)
wrapped = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M5.1 a full cycle of next-match traversal returns to the same match",
      wrapped == ref, f"ref[:80]={ref[:80]!r} wrapped[:80]={wrapped[:80]!r}")

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M5.1 exits 0 after a traversal round-trip", code == 0, f"exit {code}")

# Typing a character that is also a binding elsewhere (q, quit) must not
# quit while the query has the keyboard.
s = Session([BINARY, FIXTURE], 80, 24)
s.pump(1.2)
s.send(KEY["Find"], 0.3)
s.send(b"q", 0.4)
footer_q = screen_of(s.out, 80, 24).lines()[-1]
check("M5.1 q types into the query instead of quitting",
      "/q" in footer_q, repr(footer_q))

code = (s.send(KEY["Cancel"], 0.4), s.send(KEY["Quit"], 0.6), s.close())[2]
check("M5.1 exits 0 after the quit-while-typing check", code == 0, f"exit {code}")

# --- M5.2: navigator toggle -------------------------------------------------

s = Session([BINARY, FIXTURE], 80, 24)
s.pump(1.2)
before = screen_of(s.out, 80, 24)
before_lines = before.raw_lines()
before_body = "\n".join(before_lines)
check("M5.2 the sidebar is present before toggling",
      "Roots" in before_body and "Files" in before_body,
      repr(before_lines[1]))
nav_edge_before = before_lines[0].index("╮")

out = s.send(KEY["Navigator"], 0.6)
hidden = screen_of(s.out, 80, 24)
hidden_lines = hidden.raw_lines()
check("M5.2 toggling hides the whole sidebar",
      "Roots" not in "\n".join(hidden_lines) and "Files" not in "\n".join(hidden_lines),
      repr(hidden_lines[0]))
check("M5.2 the document viewer's border now spans the full width",
      hidden_lines[0][0] == "╭" and hidden_lines[0][-1] == "╮",
      repr(hidden_lines[0]))
check("M5.2 hiding the sidebar re-wraps the document to the width it gained",
      hidden_lines[0].index("╮") > nav_edge_before, f"{hidden_lines[0].index('╮')} vs {nav_edge_before}")

s.send(KEY["Navigator"], 0.6)
restored = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M5.2 toggling again restores the exact previous layout",
      restored == "\n".join(before_lines), repr(restored[:80]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M5.2 exits 0 after a navigator round-trip", code == 0, f"exit {code}")

# The Files Nav's own selection must survive a hide/restore cycle. Toggling
# moves focus off the Files pane (it is hidden, so it cannot hold focus, the
# same rule a narrow-terminal resize follows); S9 says nothing about
# restoring focus, only the sidebar's own selection, so PrevPane brings focus
# back to Files deliberately before checking it. The selection survives iff
# activating from it still opens README.md (row 2 of docs, empty, README.md,
# notes.txt), not docs (row 0, the reset-to-top default).
s = Session([BINARY, TREE], 80, 24)
s.pump(1.2)
for _ in range(2):
    s.send(KEY["Down"], 0.3)  # docs -> empty -> README.md
s.send(KEY["Navigator"], 0.5)
s.send(KEY["Navigator"], 0.5)
s.send(KEY["PrevPane"], 0.4)  # deliberately back to Files; S9 owes no focus
footer_focus = screen_of(s.out, 80, 24).lines()[-1]
check("M5.2 focus is back on Files before activating",
      footer_focus.startswith("Files"), repr(footer_focus))

s.send(KEY["Activate"], 0.5)
opened = screen_of(s.out, 80, 24)
check("M5.2 the selection survived the hide/restore cycle: README.md, not docs, opens",
      "Document: README.md" in "\n".join(opened.raw_lines()),
      repr(opened.raw_lines()[0]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M5.2 exits 0 after the selection-intact check", code == 0, f"exit {code}")

# --- M5.3: help modal --------------------------------------------------------

# Reachable from the footer affordance (S12), and it lists the focused pane's
# own keys: opened from Files, it must name Files' bindings (expand/collapse)
# and not another pane's (star is Roots-only).
# Height 30, not 24 (C2a/M9.1 grew filesPane's helpBindings list to 21
# entries — a 25-row box — past what a 24-row terminal fits without
# clamping; see the go-test-only TestHelpOverlayNeverExceedsTheTerminal for
# the clamped case, which this check is not exercising).
s = Session([BINARY, TREE], 80, 30)
s.pump(1.2)
before = screen_of(s.out, 80, 30)
before_lines = before.raw_lines()
top_edge_before = before_lines[0]
footer_before = before_lines[-1]
check("M5.3 focus starts on Files",
      footer_before.startswith("Files"), repr(footer_before))

s.send(KEY["Help"], 0.5)
opened = screen_of(s.out, 80, 30)
opened_body = "\n".join(opened.raw_lines())
check("M5.3 the modal names the focused pane's own context",
      "Help: Files" in opened_body, repr(opened_body[:200]))
check("M5.3 the modal lists a Files-only binding (expand)",
      "expand" in opened_body, repr(opened_body[:200]))
check("M5.3 the modal does not list a Roots-only binding (star)",
      "star" not in opened_body, repr(opened_body[:200]))

# It opens over the pane, not instead of it: the frame's own top border and
# footer, outside the centered modal box, are still exactly what they were.
check("M5.3 the top border is unchanged outside the modal",
      opened.raw_lines()[0] == top_edge_before, repr(opened.raw_lines()[0]))
check("M5.3 the footer is unchanged outside the modal",
      opened.raw_lines()[-1] == footer_before, repr(opened.raw_lines()[-1]))

# The modal owns the keyboard while open: a movement key must not reach the
# Files Nav underneath (selection unaffected once the modal is dismissed).
s.send(KEY["Down"], 0.3)
s.send(KEY["Down"], 0.3)
s.send(KEY["Cancel"], 0.5)
restored = "\n".join(screen_of(s.out, 80, 30).raw_lines())
check("M5.3 esc dismisses the modal, restoring the exact previous screen",
      restored == "\n".join(before_lines), repr(restored[:80]))

# The Help key itself dismisses it too (pressing the affordance again).
s.send(KEY["Help"], 0.5)
s.send(KEY["Help"], 0.5)
restored_again = "\n".join(screen_of(s.out, 80, 30).raw_lines())
check("M5.3 pressing Help again also dismisses the modal",
      restored_again == "\n".join(before_lines), repr(restored_again[:80]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M5.3 exits 0 after a help round-trip", code == 0, f"exit {code}")

# Opens over the Document Viewer too, with the viewer's own context (search
# keys, not Files/Roots-only ones), proving "any pane" rather than one.
s = Session([BINARY, FIXTURE], 80, 24)
s.pump(1.2)
s.send(KEY["Help"], 0.5)
viewer_help = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M5.3 opened from the viewer, the modal names the Document context",
      "Help: Document" in viewer_help, repr(viewer_help[:200]))
check("M5.3 opened from the viewer, next/prev match are listed",
      "next match" in viewer_help and "prev match" in viewer_help,
      repr(viewer_help[:200]))
check("M5.3 opened from the viewer, a Files/Roots-only binding is not listed",
      "expand" not in viewer_help and "star" not in viewer_help,
      repr(viewer_help[:200]))

code = (s.send(KEY["Cancel"], 0.4), s.send(KEY["Quit"], 0.6), s.close())[2]
check("M5.3 exits 0 after the viewer-context help check", code == 0, f"exit {code}")

# --- M7.1: the Picker modal (open, type, results, preview, activate, ---------
# --- dismiss, and the mode/Scope indicators) ---------------------------------

# Round trip first: dismissing without picking anything must restore the
# exact frame that was on screen before the Picker opened (M7.1's own
# criterion) — the same "esc restores exactly" shape M5.3's help-modal check
# already gives, applied to the Picker instead.
s = Session([BINARY, TREE], 80, 24)
s.pump(1.5)
before_lines = screen_of(s.out, 80, 24).raw_lines()

s.send(KEY["Picker"], 0.5)
opened_body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M7.1 space opens the Picker modal",
      "Results" in opened_body and "Preview" in opened_body, repr(opened_body[:200]))
check("M7.1 the Picker opens in File mode, Active Root Scope (S1's fixed default)",
      "File · Active Root" in opened_body, repr(opened_body[:200]))
check("M7.1 an empty Prompt reports the query threshold",
      "type at least 3 characters" in opened_body, repr(opened_body[:200]))

s.send(KEY["Cancel"], 0.5)
after_lines = screen_of(s.out, 80, 24).raw_lines()
check("M7.1 esc dismisses the Picker, restoring the exact frame before it opened",
      after_lines == before_lines, repr(after_lines[:1]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M7.1 exits 0 after a Picker open/dismiss round trip", code == 0, f"exit {code}")

# File mode flow: type a query, see Results and Preview populate, then
# activate the previewed Result. File is the mode the Picker always opens in
# (S1), so this never switches mode/Scope itself — kept separate from the
# mode/Scope flow below deliberately: switching mode/Scope *after* an async
# search has already populated Results/Preview raced, intermittently, with
# the cell-diffing renderer's own cursor-relative redraws (observed directly:
# identical trailing bytes replayed to two different screens depending on how
# many async-search redraws preceded them in the stream) — a replay-fragility
# finding about compounding an async search with an immediately-following
# mode/Scope switch, not a product defect (see TRIAGE). Ordering each flow to
# never do both at once sidesteps it.
s = Session([BINARY, TREE], 80, 24)
s.pump(1.5)
s.send(KEY["Picker"], 0.5)

query = b"guide"  # matches only docs/guide.md by fuzzy file-name subsequence
s.send(query, 0.9)  # clears the debounce (120ms) plus an rg run
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M7.1 a query above threshold populates the Results list",
      "docs/guide.md" in body, repr(body[:300]))
check("M7.1 the Preview pane loads the selected Result's raw content",
      "Nested guide." in body, repr(body[:300]))

s.send(KEY["Activate"], 0.6)
activated = screen_of(s.out, 80, 24)
activated_body = "\n".join(activated.raw_lines())
footer_activated = activated.lines()[-1]
check("M7.1 activating a File Result dismisses the Picker",
      "Results" not in activated_body and "Preview" not in activated_body,
      repr(activated_body[:200]))
check("M7.1 activating a File Result opens it as the Active Document, focused",
      footer_activated.split(" \u00b7")[0] == "Document", repr(footer_activated))
check("M7.1 the activated Result's pane title names the file",
      "guide.md" in activated_body, repr(activated_body[:200]))
check("M7.1 the activated Result's content is rendered",
      "Nested guide." in activated_body, repr(activated_body[:200]))

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("M7.1 exits 0 after the File mode search/preview/activate flow", code == 0, f"exit {code}")

# Mode/Scope indicator flow: both switches while the Prompt is still empty
# (no async search racing the switch, per the note above), then type once
# already in Grep mode so the query, mode and Scope all settle together in a
# single search — covering Grep's own Results/Preview (rg's smart case makes
# the lowercase query match the fixture's capitalised heading) without ever
# switching mode/Scope after a search has landed.
s = Session([BINARY, TREE], 80, 24)
s.pump(1.5)
s.send(KEY["Picker"], 0.5)

s.send(KEY["PickerMode"], 0.4)
indicator_grep = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M7.1 tab switches the mode indicator to Grep",
      "Grep · Active Root" in indicator_grep, repr(indicator_grep[:200]))

s.send(KEY["PickerScope"], 0.4)
indicator_allroots = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M7.1 shift+tab switches the Scope indicator to All Roots",
      "Grep · All Roots" in indicator_allroots, repr(indicator_allroots[:200]))

s.send(KEY["PickerScope"], 0.4)
indicator_activeroot = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M7.1 shift+tab switches the Scope indicator back to Active Root",
      "Grep · Active Root" in indicator_activeroot, repr(indicator_activeroot[:200]))

s.send(query, 1.0)  # first query typed, already in Grep mode/Active Root
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("M7.1 a Grep query populates the Results list with a matching line",
      "Grep · Active Root" in body and "docs/guide.md" in body, repr(body[:300]))
# D58: a bare "reverse video appears somewhere on the wire" check would pass
# even with Preview highlighting removed, because the selected Results row
# (pickerSelectedStyle) is itself reverse-video regardless of the Preview.
# Scope the assertion to reverse video (SGR param 7, wherever it falls in an
# SGR's semicolon list) immediately preceding the Preview's own untruncated
# rendering of the matched line ("# Guide") -- the Results row only ever
# shows that line truncated with an ellipsis ("# Gui\u2026") -- so only a
# genuine Preview highlight can satisfy it.
reverse_then_matched_line = re.compile(rb"\x1b\[(?:\d+;)*7(?:;\d+)*m[^\x1b]{0,20}# Guide")
check("M7.1 the Preview highlights the matched line (reverse video on the wire)",
      reverse_then_matched_line.search(s.out) is not None, repr(s.out[-400:]))

code = (s.send(KEY["Cancel"], 0.4), s.send(KEY["Quit"], 0.6), s.close())[2]
check("M7.1 exits 0 after the mode/Scope indicator flow", code == 0, f"exit {code}")

# --- C2a/M6.2: alt+h dotfile toggle in the Picker --------------------------
# File mode: a dotfile is a Result by default (M6.1); alt+h hides it, a
# second alt+h restores it, and the hidden state survives Cancel + reopening
# the Picker in the same session (S4's "reset only on relaunch").
#
# Every alt+h in this section is sent on an empty Prompt, before that
# Picker-open's first search — never immediately after one has landed. This
# is the same ordering rule the mode/Scope flow above already follows, for
# the same reason (picker-pty-async-ordering.md, D58): toggling right after a
# landed search races the cell-diffing emulator's cursor-relative replay
# (confirmed empirically here too — alt+h immediately after a landed search
# intermittently misread the title on ~1-in-3 runs; reordering to
# toggle-before-type made 5 consecutive isolated trials and repeated full
# `make check-pty` runs clean). Reopening the Picker (Cancel then Picker)
# always lands on a fresh empty Prompt (S8's "no memory"), which is what
# makes "toggle, then close and reopen, then toggle again" a sequence of
# three empty-Prompt toggles rather than ever a toggle-after-a-search.
hidden_root = tempfile.mkdtemp()
with open(os.path.join(hidden_root, ".secret.md"), "w") as f:
    f.write("needlecontent\n")
with open(os.path.join(hidden_root, "public.md"), "w") as f:
    f.write("public text\n")

s = Session([BINARY, hidden_root], 80, 24)
s.pump(1.5)
s.send(KEY["Picker"], 0.4)
s.send(KEY["ToggleHidden"], 0.4)  # empty Prompt: no search in flight to race
s.send(b"secret", 0.9)            # single search, already hidden
hidden = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M6.2 pty: alt+h hides the dotfile from the File Results",
      ".secret.md" not in hidden, repr(hidden[:300]))
check("C2a/M6.2 pty: alt+h marks the Prompt title Hidden",
      "File \u00b7 Active Root \u00b7 Hidden" in hidden, repr(hidden[:300]))

s.send(KEY["Cancel"], 0.3)
s.send(KEY["Picker"], 0.3)   # reopen: fresh empty Prompt, toggle untouched
s.send(b"secret", 0.9)       # no toggle this cycle -- just re-search
reopened = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M6.2 pty: the hidden state survives closing and reopening the Picker",
      ".secret.md" not in reopened and "Hidden" in reopened, repr(reopened[:300]))

s.send(KEY["Cancel"], 0.3)
s.send(KEY["Picker"], 0.3)
s.send(KEY["ToggleHidden"], 0.4)  # empty Prompt again: toggle back off
s.send(b"secret", 0.9)
restored = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M6.2 pty: a second alt+h restores the dotfile",
      ".secret.md" in restored, repr(restored[:300]))
check("C2a/M6.2 pty: a second alt+h drops the Hidden title suffix",
      "Hidden" not in restored, repr(restored[:300]))

code = (s.send(KEY["Cancel"], 0.4), s.send(KEY["Quit"], 0.6), s.close())[2]
check("C2a/M6.2 pty: exits 0 after the File mode alt+h flow", code == 0, f"exit {code}")

# Grep mode: the same empty-Prompt-only toggle rule, plus the mode/Scope
# flow's own rule of switching mode only on an empty Prompt too. Reopening
# the Picker resets mode back to File (S1's fixed opening state) but leaves
# the alt+h toggle untouched (model.hideDotfiles, not picker-owned), so each
# reopen below re-switches to Grep before typing.
grep_root = tempfile.mkdtemp()
with open(os.path.join(grep_root, ".secret.md"), "w") as f:
    f.write("needlecontent\n")
with open(os.path.join(grep_root, "public.md"), "w") as f:
    f.write("public text\n")

s = Session([BINARY, grep_root], 80, 24)
s.pump(1.5)
s.send(KEY["Picker"], 0.4)
s.send(KEY["PickerMode"], 0.3)    # empty Prompt: mode switch
s.send(KEY["ToggleHidden"], 0.4)  # empty Prompt: dotfile toggle
s.send(b"needlecontent", 1.0)     # single search, already Grep + hidden
hidden_grep = "\n".join(screen_of(s.out, 80, 24).raw_lines())
# Not "needlecontent": that's the query text itself, which the Prompt line
# always echoes back regardless of whether the search found anything -
# checking for its absence is a check that can never fail, hidden or not.
# ".secret.md" is the Grep Result row's own filename, present only when an
# actual match is listed.
check("C2a/M6.2 pty: alt+h hides a dotfile's match from the Grep Results too",
      ".secret.md" not in hidden_grep, repr(hidden_grep[:300]))
check("C2a/M6.2 pty: alt+h marks the Grep Prompt title Hidden",
      "Grep \u00b7 Active Root \u00b7 Hidden" in hidden_grep, repr(hidden_grep[:300]))

s.send(KEY["Cancel"], 0.3)
s.send(KEY["Picker"], 0.3)
s.send(KEY["PickerMode"], 0.3)    # re-enter Grep after the reopen reset it
s.send(KEY["ToggleHidden"], 0.4)  # toggle back off, still on an empty Prompt
s.send(b"needlecontent", 1.0)
restored_grep = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M6.2 pty: a second alt+h restores the Grep Result",
      ".secret.md" in restored_grep, repr(restored_grep[:300]))

code = (s.send(KEY["Cancel"], 0.4), s.send(KEY["Quit"], 0.6), s.close())[2]
check("C2a/M6.2 pty: exits 0 after the Grep mode alt+h flow", code == 0, f"exit {code}")

# --- C2a/M7.1: Peek (`p` in the Files Nav) ---------------------------------
# The inverted header (colour, unlike the plain-text checks go test already
# covers) is exactly the kind of rendering ambiguity M4.1's git markers
# needed a pty check for (.charm/knowledge/pty-verification.md) — go test
# has no TTY and so never exercises the SGR reverse-video code at all.

peek_root = tempfile.mkdtemp()
with open(os.path.join(peek_root, "active.md"), "w") as f:
    f.write("# Active\n")
with open(os.path.join(peek_root, "peeked.md"), "w") as f:
    f.write("# Peeked\n")

s = Session([BINARY, peek_root], 80, 24)
s.pump(1.2)
body = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M7.1 pty: both files listed before Peeking",
      "active.md" in body and "peeked.md" in body, repr(body[:300]))

s.send(KEY["Activate"], 0.5)  # active.md becomes the Active Document; focus -> viewer
s.send(KEY["FocusFiles"], 0.4)  # back to the Files Nav (S6)
s.send(KEY["Down"], 0.4)  # select peeked.md

mark = len(s.out)
s.send(KEY["Peek"], 0.6)
sc = screen_of(s.out, 80, 24)
body = "\n".join(sc.raw_lines())
check("C2a/M7.1 pty: Peek shows the non-Active file's content",
      "Peeked" in body, repr(body[:300]))
check("C2a/M7.1 pty: Peek header names the Peeked document",
      "Peek: peeked.md" in body, repr(body[:300]))
check("C2a/M7.1 pty: the Active Document's own content is hidden while Peeking",
      "# Active" not in body, repr(body[:300]))
# lipgloss combines Bold+Reverse into one SGR run (e.g. \x1b[1;7m), not the
# bare \x1b[7m M2.2's own reverse-video check gets away with checking for
# directly — titleStyle is Bold, and peekTitleStyle adds Reverse on top of
# it, so the "7" parameter has to be matched inside whatever run it lands in.
reverse_sgr = re.compile(rb"\x1b\[(?:\d+;)*7(?:;\d+)*m")
check("C2a/M7.1 pty: the inverted Peek header reaches the wire as reverse video",
      reverse_sgr.search(s.out[mark:]) is not None, repr(s.out[mark:][:200]))

frame_before_scroll = "\n".join(screen_of(s.out, 80, 24).raw_lines())
s.send(KEY["HalfPageDown"], 0.4)  # a scroll-only binding, never a Files Nav one
frame_after_scroll = "\n".join(screen_of(s.out, 80, 24).raw_lines())
check("C2a/M7.1 pty: a Peeked document cannot be scrolled",
      frame_after_scroll == frame_before_scroll,
      f"before={frame_before_scroll[:200]!r} after={frame_after_scroll[:200]!r}")

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("C2a/M7.1 pty: exits 0 after Peeking", code == 0, f"exit {code}")

# --- C2a/M8.1: two-line Roots Nav rows --------------------------------------

# Two same-named roots (the shape M3.3's own fixture used) pins S5's row
# shape: name and path must reach the wire as two separate physical lines,
# not one. The exact bold/unbold and reverse-video SGR *combinations* per
# line are already pinned precisely at the go-test level
# (TestRootsNavRendersTwoLinesPerEntry, TestRootsNavSelectionHighlightsBothLines) —
# Screen.raw_lines() parses and drops SGR entirely (pty-verification.md), so
# it cannot re-check *which* line carries which attribute. What the pty can
# still add: that bold and reverse video genuinely reach the wire (not just
# a `go test`, no-TTY artifact) during this pane's very first paint, and
# that browsing with Down actually moves between the real, distinct roots
# on screen — a functional, not stylistic, check.
rows_base = tempfile.mkdtemp()
row_a = os.path.join(rows_base, "work", "admin", "project")
row_b = os.path.join(rows_base, "other", "project")
os.makedirs(row_a)
os.makedirs(row_b)

home_dir = tempfile.mkdtemp()
env, store_path = config_env(home_dir)
os.makedirs(os.path.dirname(store_path))
with open(store_path, "w") as f:
    f.write(row_a + "\n" + row_b + "\n")

s = Session([BINARY], 80, 24, env=env)
initial = s.pump(1.5)  # the pane's first paint: nothing to diff against yet,
# so every visible cell's SGR is emitted in full here, unlike a later
# incremental redraw that may only emit what actually changed.

lines = screen_of(initial, 80, 24).raw_lines()
name_line_idx = next(i for i, l in enumerate(lines) if "project" in l)
path_line_idx = next(i for i, l in enumerate(lines) if "admin" in l)
check("C2a/M8.1 pty: the name and its path land on two separate lines",
      name_line_idx != path_line_idx, f"name={name_line_idx} path={path_line_idx}")
check("C2a/M8.1 pty: the path line does not repeat the active/starred markers",
      "\u25cf" not in lines[path_line_idx] and "\u2605" not in lines[path_line_idx],
      repr(lines[path_line_idx]))

bold_sgr = re.compile(rb"\x1b\[(?:\d+;)*1(?:;\d+)*m")
reverse_sgr = re.compile(rb"\x1b\[(?:\d+;)*7(?:;\d+)*m")
check("C2a/M8.1 pty: a bold name line reaches the wire on the Roots pane's first paint",
      bold_sgr.search(initial) is not None, repr(initial[:200]))
check("C2a/M8.1 pty: a reverse-video selected row reaches the wire too",
      reverse_sgr.search(initial) is not None, repr(initial[:200]))

# One Down steps a whole entry (two lines): activating afterwards must make
# root_b (other/project) active, not root_a — the functional proof that
# selection tracks whole entries, not the individual lines S5 added.
s.send(KEY["PrevPane"], 0.4)  # Files -> Roots
s.send(KEY["Down"], 0.4)
s.send(KEY["Activate"], 0.6)
lines = screen_of(s.out, 80, 24).raw_lines()
admin_after = row_pair_with("admin", lines)
other_after = row_pair_with("other", lines)
check("C2a/M8.1 pty: one Down then Activate makes the second entry active, not the first",
      "\u25cf" in other_after and "\u25cf" not in admin_after,
      f"admin={admin_after!r} other={other_after!r}")

code = (s.send(KEY["Quit"], 0.6), s.close())[1]
check("C2a/M8.1 pty: exits 0 after browsing two-line Roots Nav rows", code == 0, f"exit {code}")

done()
