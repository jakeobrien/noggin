# `noggin` Spec

`noggin` is a TUI for viewing markdown files with an integrated file/folder exlorer and manager. Its primary purpose is efficiently accessing and viewing markdown files. It is explicitly not a text editor. 

While the markdown renderer is at the heart of its functionality, most of the actual feature set involve navigating, parsing and managing files, including:
- rendering file/folder trees
- basic file/folder editing (add, move, copy, delete)
- parsing metadata from files (tags)
- fuzzily searching for file names and grepping within files
- managing navigation state ("saved" folder roots, when files were last viewed, "starred" files)
- parsing wiki-links in documents and navigating to them
- parsing urls in documnts and opening them
- opening files to edit in nvim (via tmux or handoff process)

## Glossary

- "Root": A folder that serves as the root folder for file visibility. Roots can be added and removed to a session, but only one root can be "active" at a time.
- "Active Root": The Root that is active. Defines scope for File Nav and searching.
- "Browse By": The mode of the Files Nav. Defines how Files Nav behaves.
- "Active Document": The markdown document being rendered in a Document Viewer. (Can be modally replaced by a Preview).
- "Preview": A markdown document being rendered in a Document Viewer, but only modally. The Viewer reverts to the Active Document when Preview state is gone.
- "Outline": A table of contents-like list (or tree) of document headers.
- "Starred": A file can be "starred" which adds a pointer to it (relative to its root)
- "Nav Keys": Standardized keys for navigating: see Keyboard Commands below.
- "Tag": A label added to a document, used for non-hierarchical organization and relation. Tags can be nested via "/" (i.e. "projects/work").
- "Edit": Open the selected document in nvim or $EDITOR (via tmux or handoff process)

## Overview

The UI tree:
- window
    - Navigator
        - Roots Nav
        - Files Nav
    - Documents
        - Document
            - Document Viewer
            - Document Outline (optional)
        - ... (optional)
    - Footer

```
-------------------------------------------------------------------------|
| Navigator | Documents                                                  |
|-----------|------------------------------------------------------------|
| Roots Nav | Document                           | Document              |
|           |------------------------------------------------------------|
|-----------| Document Viewer | Document Outline | Document Viewer       |
| Files Nav |                 |                  |                       |
|           |                 |                  |                       |
|           |                 |                  |                       |
|------------------------------------------------------------------------|
| Footer                                                                 |
--------------------------------------------------------------------------
```

## State

### Persisted
- Saved Roots
- Per-root:
    - Per-file
        - Starred
        - Last viewed timestamp


### Ephemeral (per session)
- Active Root
- Selected Root
- Browse By mode
- Active Parsed By mode
- Active Document[]
- Document Cursor[]
- Selected file/folder
- Last search (scope and type)
- Last Results Set (excludes Tags-derived)
- Focus (Root Nav/Files Nav/Document Viewer/Outline Viewer/Modal/Dialog)
- Files opened this session (as stack for back/forward)

## UI Panes

### Navigator

The left sidebar panel of the TUI. Holds two components, the Roots Nav at the top (~15%) and the Files Nav below.

#### Roots Nav
Shows the Roots in the current session in a list. Currently active Root is highlighted. When focused, Nav Keys move selection. <Enter> makes selected root active. Entries have the name (base name of the folder) in bold, left-aligned, and the root's parent folder in non-bold, right-aligned and left-truncated to allow 3 characters between it and the name.

Changing the Active Root changes the root folder for the Files Nav, but does not change the active document.

#### Files Nav

Consists mainly of Files, but with a thin header showing Browse By mode. It also owns three modal UI panels: Browse By Picker, Tag Browser and File Search.

##### Files

Displays files based on Active Root and Browse By mode. Some modes have a tree structure, some are flat. When focused, nav keys move selection. Activation makes the selected file the active document and focuses the Document Viewer. 

Besides being Activated, files can be Previewed. Previewing also makes the targeted Document Viewer render the file, but it is ephemeral. When selection changes, the Document Viewer reverts to the Active Document. "Follow" mode can be enabled, where any change of selection auto-previews the file.

##### Browse By modes

----------------------
| Mode | Flat/Tree | Scope | Description |
|------|-----------|-------|-------------|
| Tree | Tree | Active Root |  mirrors folder structure with the root at the Active Root |
| Recent | Flat | Active Root |  list of most recently viewed files in the Active Root |
| All Recent-All | Flat | All Roots |  list of most recently viewed files across all Roots |
| Starred | Flat | Active Root | list of Starred files in th Active Root |
| Starred-All | Flat | All Roots | list of Starred files in th Active Root |
| Results | Flat | N/A | list of files manually provided by a search or Tag filter  |


##### Browse By Picker

When presented allows user to select the Browse By mode. All but `Results` are directly selectable. `Results` is indirectly selected as described in Results Sets. In addition to the directly selectable modes, the user can select two additional options:
- "Tags": which does not directly set the Browse By mode, but presents the Tag Browser
- "Last Search": displayes the last (cached) Results Set

##### Results Sets

Results Sets are sets of files sent to the Files Nav to display in Results Browse By mode.

##### Tag Browser

The Tag Browser displays all tags for the scope (Active Root or all Roots) in a tree list. Nested tags follow folder structure and so the same navigation/expand/collapse paradigms can be used. When activated, a Results Set is generated containing all files that have the given Tag. This Results Set does not override the Last Results Set. The Results Set is given to the Files Nav and Browse By mode is set to Results.

##### File Search

A modally displayed fuzzy finder across files. Uses the Fuzzy Finder UX (see below) scoped to file results. Can have the following scopes:
- Active Root
- All Roots

and the following modes:
- Match file name
- Grep (search for text in files)


### Documents

The main panel of the TUI. Displays one to two Document Viewer/Outline Viewer sets. 

#### Document Viewer

The markdown file renderer. **Not an editor**

Navigable (see Keyboard Commands) in two modes:
- Normal mode: cursor moves but does not select. Only actions besides navigation are following urls or wiki-links.
- Select mode: selects along with navigation. Supports copy after select.

#### Outline Viewer

Paired with a Document Viewer. Optionally displayed (hidden by default). Pulls the headings (H1-H6) from the document and displays as a list. When focused, navigation can be used to select, preview and jump to the selected header in Document Viewer.

#### Document Search

A modally displayed fuzzy finder for text in the Active Document. Uses the Fuzzy Finder UX (see below) scoped to the Active Document.

### Footer

A bar at the bottom of the TUI that shows contextual keyboard commands for the focused panel. If there are too many to show, will truncate and show a `?` option that will display a help modal showing all.

### Fuzzy Finder

A general purpose UX for fuzzy finding. Has three components:
- Input field (top)
- Results list (bottom left)
- Preview pane (bottom right)

Results update as a user types (possibly with a throttle ms, especially for grep). Matching files are displayed in the Results list, sorted by match strength. The selected file will be rendered in the Preview pane.

Activating a result dismisses the Fuzzy Finder and navigates the targeted Document's Active Document to the selected location.

### Modals

The following modal UIs exist (only one of which can be displayed at a time):
- Fuzzy Finder (as File Search or Document Search)
- Search Picker
- Browse By Picker
- Tag Browser

as well as dialogs (which display above modals):
- Confirm dialogs (delete file/folder)
- Input fields (move file name)
- Keyboard help


## Keyboard commands

### Global
Always availabe (except when a modal/dialog is present).

- `q`: exit/quit app
- `N`: toggle Navigator visibility
- `r`: focus Roots Nav
- `n`: focus Files Nav
- `b`: open Browse By Picker
- `d`: focus Document Viewer
- `<space>`: open Search Picker

### General
Context-dependent, but universal when applicable.

Nav keys:
- `up`/`j`: up
- `down`/`k`: down
- `left`/ `h`: left
- `right`/`l`: right
- `<shift>+`: 5x movement
- `<ctrl>+`: 15x movement
- `<alt>+`: 50x movement
- `g`: top
- `G`: bottom

### Focus-specific

#### Navigator

##### Roots Nav

- `a`: add new root
- `X`: remove root (w/ confirm)
- `o`: open root folder in finder
- `t`: open tmux pane in root folder
- `s`: toggle whether root is "starred" (persisted to data, will be auto-added on each new session)
- `c`: copy root path

##### Files Nav

Selection-independent: 
- `f`: toggle "follow" (auto-previews any selected file)
Files or folder selected:
- `c`: copy file/folder path
- `e`: Edit file
- `a`: add file (prompt for name and create new file (child if folder is selected, sibling if file is selected)
File selected:
- `<enter>`: set as the active document in the targeted Document Viewer and focus it
- `p`: peek the document (Document Viewer will show the document until item is unselected)
- `m`: move file (show prompt with current path, after user edits, move file to new location) (error on collision)
- `D`: duplicate file (create copy of file in same folder
- `X`: delete file (w/ confirm)
Folder selected:
- `X`: delete folder (w/ confirm) (error if folder non-empty)

##### Browse By Picker

- `d`: Tree
- `r`: Recent
- `R`: Recent-All
- `s`: Starred
- `S`: Starred-All
- `t`: Tags

#### Document

##### Document Viewer

- `v`: Enter Select mode
- `<esc>`: Enter Normal mode
- `[`: go to previous document in history
- `]`: go to next document in history
- `^`, `$`, `w`, `e`, `b`: corresponding vim motions

###### Normal Mode

- `J`: (if cursor is within) open URL in default browser/open wiki-link locally
- `)`: jump to next URL/wiki-link
- `(`: jump to previous URL/wiki-link
- `o`: focus outline
- `O`: toggle Outline visibility
- `E`: Edit
- `F`: open Document Search

###### Select Mode

- `c`, `y`: copy selected

##### Outline Viewer

- `<enter>`: jump to selected header in Document Viewer and focus
- `o`: unfocus outline (focus Document)
- `O`: unfocus outline (focus Document) and hide outline

#### Search Picker

- `s`: open file name search scoped to Active Root
- `S`: open file name search scoped to All Roots
- `,`: open file name search scoped to Documents opened this session
- `/`: open grep search scoped to Active Root
- `\`: open grep search scoped to All Roots
- `.`: open most recent search

#### Fuzzy Finder

- `<esc>`: dismiss
- Nav Keys: navigate results
- `<enter>`: open in Document Viewer
- `<ctrl>+d`: scroll preview down
- `<ctrl>+u`: scroll preview up
- `<ctrl>+o`: send Results to File Nav
- `<ctrl>+h`: toggle hidden
- `<ctrl>+i`: toggle ignored

### Targetting

Since we can have two Document Viewers, the following modifiers apply to determine which we are targeting:
- No mod: Document Viewer 1
- `<ctrl>+`: Document Viewer 2

These apply to:
- Genera Nav:
    - `d` - focus document viewer
- Files Nav:
    - `<enter>` - load
    - `p` - peek
    - `f` - follow
- Fuzzy Finder:
    - `<enter>` - load
- Document Viewer:
    - `J` (if wiki-link)

Ambiguity handling:
- If Document Viewer does not exist, silent no-op (if we develop a need for quiet notifications, we can add this)
    - Example: `<ctrl>+p` is entered when only 2 DVs are shown, no-op
- Only one Document Viewer can follow at a time, so if `f` toggle follow for DV1, `<shift>+f` will disable follow of DV1 and enable it for DV2. 

## Tech stack

Golang using BubbleTree. Glamour gets us a beautiful markdown renderer for free.

## Open Questions

### Tags notation

Options:
- Obsidian-style (frontmatter)
- Free text 
- Metadata

Considerations:
- Parsing/cacheing strategy/efficiency

## Future/Out of scope
- Search query history
- Backlinks parsing (Outline Viewer mode)
- Create daily note
- Create zettelkasten note
- Simple editing
- Outline re-ordering
- Audio input
- Encryption
- Sync (git?)
- wikipedia pane
- agent chat pane
- scrap notes

