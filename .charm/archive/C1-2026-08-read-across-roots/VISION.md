# Vision — C1 read-across-roots

## Goals

Get to the point where noggin is worth opening every day, even before it does the thing it
is ultimately for.

The core use case (whole-product, not just this cycle): several consistent note locations
(personal notes, shared work notes) plus the local docs of several in-flight projects. Run
`noggin .` and instantly have the persisted sources available with the cwd as an ad-hoc
source. Find what's needed by searching, read it, compare and reconcile across sources,
hand off to nvim when editing is the point.

This cycle takes the first slice: **multiple roots, one document at a time, read it well,
leave to edit.**

- Roots: several at once, persisted across sessions, plus an ad-hoc one from the command line.
- Quick switching between roots — a keystroke, not a context reload.
- Tree browsing within the active root.
- A markdown renderer good enough to prefer over nvim's.
- Find within the open document.
- Edit: hand the document to nvim / `$EDITOR` at a well-defined boundary.
- Keyboard-centric throughout; fluent, ergonomic navigation.

## Non-goals / boundaries

- Not a text editor. Editing happens in `$EDITOR`, via an explicit handoff.
- No cross-root or filename/content search this cycle (C2).
- One document on screen; no side-by-side (C3).
- No memory beyond saved roots: no history, no starred documents, no recent (C4).
- No tags, no outline, no file/folder management (C5).

## Notes

- `NOTES.md` at the project root is the original, deliberately over-specified design dump.
  Treat its keymap and pane design as **intent, not a locked design** — the author's words.
  It is a source for the roadmap, not a contract.
- Appetite: about a week, agent-driven. Expected to be iterated on for a long time after.
- **Go with the grain of the libraries.** Feature direction balances intent against what
  trusted libraries and tools already do. If a library gives us something natively in a way
  that differs from the spec, try their way first. If getting a spec'd nicety (e.g. a preview
  pane) means authoring a subsystem from scratch, consider dropping the nicety instead. This
  governs the whole project, and the spec should be read as intent, not as a design to be
  reproduced exactly.
- **The keymap is provisional.** The fluency of keyboard-driven interaction is the goal; the
  specific bindings in `NOTES.md` were overstated and are expected to be fiddled with. Keys
  should be easy to change at build time. Not a user-facing configuration feature.
- Glamour (as used by Glow) is the rendering bar — its availability as a library is what made
  this project feel reachable.
- The MVP the author first named (roots + find + grep + in-doc find + tree + renderer + edit)
  is bigger than one week; it is C1 + C2 here.
- Honest framing agreed in brainstorm: C1 is a **dogfooding vehicle**, not the value
  proposition. The differentiator — compare across sources — arrives in C3.
