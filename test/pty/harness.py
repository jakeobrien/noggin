"""Drive the real noggin binary from a pty and assert on what it draws.

Bubble Tea v2 asks the terminal several questions before its first frame and
blocks until they are answered. A harness that ignores them sees only the
queries and hangs. The query set differs from v1's (no cursor-position report;
OSC 11 is BEL-terminated; Primary DA and two DECRQM mode queries are new), so
each is answered as it appears, as many times as it appears in the chunk.
"""

import fcntl
import os
import pty
import re
import select
import struct
import sys
import termios
import time

# query bytes -> reply bytes. Answered every time they appear.
REPLIES = [
    (b"\x1b]11;?\x07", b"\x1b]11;rgb:0000/0000/0000\x07"),   # OSC 11 background
    (b"\x1b]11;?\x1b\\", b"\x1b]11;rgb:0000/0000/0000\x1b\\"),  # ST-terminated form
    (b"\x1b[c", b"\x1b[?62;22c"),                            # Primary DA
    (b"\x1b[?2026$p", b"\x1b[?2026;0$y"),                    # DECRQM: unsupported
    (b"\x1b[?2027$p", b"\x1b[?2027;0$y"),                    # DECRQM: unsupported
    (b"\x1b[6n", b"\x1b[24;1R"),                             # cursor position (v1)
]

CSI = re.compile(rb"\x1b\[[0-?]*[ -/]*[@-~]")
OSC = re.compile(rb"\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)")
OTHER = re.compile(rb"\x1b[()][0-9A-Za-z]|\x1b[=><]|\x1b[PX^_].*?\x1b\\")


class Session:
    def __init__(self, argv, cols=80, rows=24, term="xterm-256color"):
        self.cols, self.rows = cols, rows
        self.out = b""
        self.pid, self.fd = pty.fork()
        if self.pid == 0:
            os.environ["TERM"] = term
            os.environ.pop("NO_COLOR", None)
            os.environ["COLORTERM"] = "truecolor"
            os.execvp(argv[0], argv)
        fcntl.ioctl(self.fd, termios.TIOCSWINSZ,
                    struct.pack("HHHH", rows, cols, 0, 0))

    def pump(self, seconds=0.6):
        """Read for a while, answering every query as it arrives."""
        end = time.time() + seconds
        while time.time() < end:
            r, _, _ = select.select([self.fd], [], [], 0.05)
            if not r:
                continue
            try:
                chunk = os.read(self.fd, 65536)
            except OSError:
                break
            if not chunk:
                break
            self.out += chunk
            for query, reply in REPLIES:
                for _ in range(chunk.count(query)):
                    os.write(self.fd, reply)
        return self.out

    def send(self, data, settle=0.5):
        os.write(self.fd, data)
        return self.pump(settle)

    def resize(self, cols, rows, settle=0.6):
        self.cols, self.rows = cols, rows
        fcntl.ioctl(self.fd, termios.TIOCSWINSZ,
                    struct.pack("HHHH", rows, cols, 0, 0))
        return self.pump(settle)

    def close(self):
        try:
            os.close(self.fd)
        except OSError:
            pass
        _, status = os.waitpid(self.pid, 0)
        return os.waitstatus_to_exitcode(status)


def strip(data):
    """Remove escape sequences, leaving the characters that occupy cells."""
    for pat in (OSC, CSI, OTHER):
        data = pat.sub(b"", data)
    return data


def rows_of(data):
    """The drawn rows: escapes stripped, split on \\r?\\n (v2 mixes both)."""
    return [r for r in re.split(rb"\r?\n", strip(data))]


def last_frame(data, height):
    rs = rows_of(data)
    while rs and rs[-1].strip() == b"":
        rs.pop()
    return rs[-height:]


class Screen:
    """A minimal terminal emulator, enough to measure what v2 actually draws.

    Stripping escapes is not enough under Bubble Tea v2: its ultraviolet
    renderer diffs cells and edits rows in place, expressing the result as
    ECH (erase-chars), CUF (cursor-forward), DCH (delete-chars), horizontal
    tabs and SGR runs rather than as whole rows of literal characters. A
    stripped row therefore loses every gap and keeps text that DCH removed.
    Widths and row contents have to be read off a cell grid instead.

    Implemented because noggin's own frames provoke them: CUD/CUU/CUF, CUP,
    ECH, DCH, ICH, IL, DL, EL, ED, tabs (8-column stops) and \r/\n. SGR is
    parsed and dropped — colour is asserted against the raw byte stream, where
    it is unambiguous.
    """

    def __init__(self, cols, rows):
        self.cols, self.rows = cols, rows
        self.clear()

    def clear(self):
        self.grid = [[" "] * self.cols for _ in range(self.rows)]
        self.x = self.y = 0

    def _put(self, ch):
        if self.y < self.rows and self.x < self.cols:
            self.grid[self.y][self.x] = ch
        self.x += 1

    def feed(self, data):
        text = data.decode("utf-8", errors="replace")
        i, n = 0, len(text)
        while i < n:
            c = text[i]
            if c == "\x1b":
                m = re.match(r"\x1b\[([0-?]*)([ -/]*)([@-~])", text[i:])
                if m:
                    self._csi(m.group(1), m.group(3))
                    i += m.end()
                    continue
                m = re.match(r"\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)", text[i:])
                if m:
                    i += m.end()
                    continue
                m = re.match(r"\x1b[()][0-9A-Za-z]|\x1b[=><]", text[i:])
                i += m.end() if m else 1
                continue
            if c == "\r":
                self.x = 0
            elif c == "\n":
                self.y += 1
                if self.y >= self.rows:
                    self.grid.pop(0)
                    self.grid.append([" "] * self.cols)
                    self.y = self.rows - 1
            elif c == "\b":
                self.x = max(self.x - 1, 0)
            elif c == "\t":
                self.x = min((self.x // 8 + 1) * 8, self.cols - 1)
            elif c >= " ":
                self._put(c)
            i += 1

    def _csi(self, params, final):
        if params[:1] in ("?", ">", "=", "<"):
            return  # private modes and device queries: nothing to draw
        ps = [int(p) if p else 0 for p in params.split(";")] if params else [0]
        n = ps[0] or 1
        if final == "C":
            self.x = min(self.x + n, self.cols - 1)
        elif final == "D":
            self.x = max(self.x - n, 0)
        elif final == "A":
            self.y = max(self.y - n, 0)
        elif final == "B":
            self.y = min(self.y + n, self.rows - 1)
        elif final == "H":
            self.y = (ps[0] - 1 if ps and ps[0] else 0)
            self.x = (ps[1] - 1 if len(ps) > 1 and ps[1] else 0)
        elif final == "X":  # ECH: erase n cells in place
            for k in range(self.x, min(self.x + n, self.cols)):
                self.grid[self.y][k] = " "
        elif final == "P":  # DCH: delete n cells, shifting the rest left
            row = self.grid[self.y]
            del row[self.x:self.x + n]
            row.extend([" "] * (self.cols - len(row)))
        elif final == "@":  # ICH: insert n blanks, shifting the rest right
            row = self.grid[self.y]
            for _ in range(n):
                row.insert(self.x, " ")
            del row[self.cols:]
        elif final == "L":  # IL: insert n blank rows at the cursor
            for _ in range(n):
                self.grid.insert(self.y, [" "] * self.cols)
            del self.grid[self.rows:]
        elif final == "M":  # DL: delete n rows at the cursor
            del self.grid[self.y:self.y + n]
            while len(self.grid) < self.rows:
                self.grid.append([" "] * self.cols)
        elif final == "K":
            mode = ps[0]
            rng = range(self.x, self.cols) if mode == 0 else \
                  range(0, self.x + 1) if mode == 1 else range(0, self.cols)
            for k in rng:
                self.grid[self.y][k] = " "
        elif final == "J":
            mode = ps[0]
            if mode == 0:
                for k in range(self.x, self.cols):
                    self.grid[self.y][k] = " "
                for r in range(self.y + 1, self.rows):
                    self.grid[r] = [" "] * self.cols
            elif mode == 2:
                self.clear()

    def lines(self):
        return ["".join(r).rstrip() for r in self.grid]

    def raw_lines(self):
        return ["".join(r) for r in self.grid]

    def text(self):
        return "\n".join(self.lines())


def screen_of(data, cols, rows):
    s = Screen(cols, rows)
    s.feed(data)
    return s


FAILURES = []


def check(name, ok, detail=""):
    print(("PASS  " if ok else "FAIL  ") + name + (("  " + detail) if detail else ""))
    if not ok:
        FAILURES.append(name)


def done():
    print()
    if FAILURES:
        print(f"{len(FAILURES)} failure(s): " + ", ".join(FAILURES))
        sys.exit(1)
    print("all pty checks passed")
