# ripgrep is required, and its ignore rules decide what Pickers see

Both Pickers depend on `ripgrep`. It supplies the candidate file list as well as the text search, and its ignore rules decide which files a Picker can see. Without `rg`, documents still read normally, but the File and Grep Pickers are unavailable. Both Pickers need the same idea of which files exist. If the File Picker and Grep Picker disagreed with each other, that would be worse than the Pickers disagreeing with the Files Nav. The Grep Picker therefore searches the File Picker's exact file list. It doesn't let `rg` walk the directories again, because `rg` doesn't apply ignore rules to files named on its command line.

## Consequences

- The Files Nav Tree uses its own small `.gitignore` matcher rather than `rg` or `git`. A Root need not be a git repository, and `git check-ignore` only works inside one. The Tree and the Pickers can therefore disagree slightly about ignored files. This is accepted.
- Filename search, the core of the Picker, doesn't work on a machine without `rg`. Fuzzy scoring happens in-process, since `rg` has no fuzzy matching.

_Formerly D28, D38; Tree rationale from D18._
