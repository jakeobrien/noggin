# noggin never watches the filesystem

Refresh is always manual. The Staleness indicator only reports that something changed. It is fed by a cheap poll every few seconds, plus a check whenever the terminal regains focus, and never reloads anything by itself. The poll stats the directories the Tree already knows about, re-lists only those whose mtime moved, and stats open documents directly. It never re-walks the Root. Some changes go undetected by design: in-place edits to files that aren't open, and changes to git status.

## Considered Options

- **Recursive fsnotify watcher**: rejected. On macOS, kqueue holds one file descriptor per watched file. inotify has watch limits. Network and VM mounts deliver no events, so the watcher would fail silently in exactly the sandboxed and remote setups noggin runs in.
- **Native FSEvents**: needs cgo.
- **Periodic full re-walk**: too costly on a large Root.

If live reload of open documents is wanted later, the next step is to watch only their parent directories (at most two), not the whole tree.

_Formerly D49 (detection), D139._
