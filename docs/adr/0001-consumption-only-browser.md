# noggin reads documents; authoring always hands off

noggin is a browser for documents from several Sources: local notes, and remote reference such as Wikipedia. It reads, navigates and remembers. It never authors. Editing suspends noggin and hands the terminal to `$EDITOR`, then re-renders when the editor exits. It does not run the editor in a tmux pane or edit in-app. LLM chat was turned down for the same reason: it creates documents rather than serving them, and it would bring in an input field that owns the keyboard, live-changing content, and create/model-selection flows. The way in is to hand off to a dedicated tool and read its saved output back as an ordinary Root. Taking on authoring would turn a reading product into an editor, and good editors and chat TUIs already exist.

_Formerly D3, D92, D93._
