# noggin

A terminal reader for markdown notes and other documents, drawn from several Sources at once. It reads, navigates and remembers; it never authors.

## Language

### Sources

**Source**:
Anything noggin serves documents from; the entity listed in the Sources Nav. Two classes: Root and Query Source.
_Avoid_: Collection (north-star language only, not an in-app entity)

**Root**:
A folder that bounds which files are visible. Identified by exact path; several may be open, but none may nest inside another.
_Avoid_: File Source, folder source

**Query Source**:
A Source reached only by search: a remote service, exactly one of each kind, always present, never created, removed or starred. Has no Tree and no grep.
_Avoid_: Finder

**Active Source**:
The one Source the Files Nav is currently showing.

**Active Root**:
The Active Source when it is a Root. Used where only a Root can qualify (e.g. Grep).

**Ad-hoc Root**:
A Root supplied on the command line for this session only, not persisted. At most one per session.

**Broken Root**:
A Root whose path no longer exists on disk. Still listed and its starred state kept, but it cannot be activated.

**Root name**:
The editable label of a persisted Root, defaulting to its folder's base name. An Ad-hoc Root has none until starred.

**Wikipedia Source**:
The first Query Source: English Wikipedia.

### Documents and files

**Markdown file**:
A file noggin can render as markdown, recognised by a markdown extension or a conventional name such as `README` or `CHANGELOG`.
_Avoid_: Readable file (retired)

**Text file**:
Any file that is neither a Markdown file nor binary. Only ever shown as raw text. Binary files are never shown.
_Avoid_: Readable file (retired)

**File Filter**:
Which files a Root offers for discovery: `Markdown` (Markdown files only) or `All` (Text files too). One per Root.
_Avoid_: file mode, show-all

**View toggles**:
The two session-wide switches over what exists in the Tree and the File/Grep Pickers: hidden files and gitignored files.
_Avoid_: show-hidden mode

**Article**:
A document served by the Wikipedia Source, identified by its canonical URL.

**Chrome**:
Everything a fetched Wikipedia page carries that is not the article itself (navboxes, sidebars, infoboxes, citations, reference lists and the like). Stripped before conversion; content tables are not chrome.

**Frontmatter Card**:
A document's YAML frontmatter drawn as a bordered key/value box above its rendered body.

### Panes

**Navigator**:
The left sidebar: Sources Nav above, Files Nav below. Hideable.

**Sources Nav**:
The upper Navigator pane, listing the Ad-hoc Root, persisted Roots, then Query Sources.
_Avoid_: Roots Nav

**Files Nav**:
The lower Navigator pane: the Active Root's Tree, or its Recent and Starred lists. A Query Source has only the lists.

**Browse Mode**:
How the Files Nav lists documents: `Tree`, `Recent` or `Starred`.

**Browse Scope**:
What a Browse Mode draws from: `Recent` is Active Source or All Sources; `Starred` is Active Source, All Starred or Orphans. `Tree` has none. Distinct from a Picker's Scope.

**Document Viewer**:
A read-only pane showing one Active Document in its Render Mode. DV1 is always present; DV2 is an optional second viewer beside it, fully independent.

**Focus**:
The one pane currently receiving keyboard input. DV1 and DV2 are separately focusable.

**Focus Badge**:
The key that focuses a pane, shown on that pane's title row.

**Footer**:
The persistent bottom bar: contextual key hints plus transient messages.

**Help Modal**:
The modal listing the focused pane's key bindings.

**Location line**:
The de-emphasised bottom row of a Document Viewer naming where its document lives and how it is shown.

**Staleness indicator**:
A passive row above the Footer signalling that an open document or the Active Root's tree changed on disk. It never refreshes anything itself.

**Theme**:
noggin's whole colour scheme: one set of semantic colours every cell is painted from.

### Viewing

**Active Document**:
The document a Document Viewer is currently showing. Independent of the Active Source.

**Render Mode**:
How a Document Viewer shows a Markdown file: `Markdown` (rendered) or `Text` (raw source, line-numbered). One per viewer, sticky across documents.
_Avoid_: view mode, raw mode

**Activate**:
Commit to the selected document: it becomes the target viewer's Active Document. Never changes the Active Source.

**Peek**:
A transient, read-only look at the Files Nav's selected document in a viewer, without making it the Active Document or moving Focus. Distinct from the Preview pane.

**Targeting modifier**:
The modifier on Activate and Peek that sends the action to DV2 instead of DV1, creating DV2 if needed.

**Edit**:
The handoff of the Active Document to `$EDITOR`, resuming when it exits. Not in-app editing.

**Refresh**:
The on-demand re-read of the Active Root's tree and every open document, preserving selection and scroll. Re-fetches an open Article.

**Document History**:
A Document Viewer's own back/forward stack of the documents opened in it. Per viewer, session-only.

### Finding

**Find**:
Incremental search within the Active Document. Always within one document; searching across files is a Picker.
_Avoid_: search

**Committed Find**:
A Find query locked in, as distinct from one still being typed.

**Picker**:
The modal search across files or a Query Source: a Prompt, a Results list and a Preview pane. Comes in modes: File Picker, Grep Picker, Wikipedia Picker.
_Avoid_: Finder, Fuzzy Finder, File Search

**File Picker**:
The Picker whose Results are files, matched fuzzily by name.

**Grep Picker**:
The Picker whose Results are matching lines across files. Never includes Query Sources.

**Wikipedia Picker**:
The Picker mode that searches Wikipedia article titles. No Scope and no Preview.

**Prompt**:
A Picker's input field.

**Result**:
One row in a Picker: a file, a file-and-line, or an Article.

**Results list**:
The Picker pane listing its Results.

**Preview pane**:
The part of a Picker showing the selected Result's file. Never changes the Active Document.

**Scope**:
What a File or Grep Picker searches: Active Root, All Roots or Starred.

### Memory

**Document Store**:
The persisted record of every document noggin knows about: whether it is Starred and when it was last viewed. Not attached to Roots.

**Starred**:
Bookmarked and persisted across sessions. Applies to Roots and to documents, independently: un-starring a Root never touches its documents' stars.

**Recent**:
Documents ordered by when they were last viewed. Viewing means Activate or Edit; Peek does not count. Recency only, never frecency.
_Avoid_: frecency, history (that is Document History)

**Orphan**:
A Starred document that lies under no Root in the current session.
