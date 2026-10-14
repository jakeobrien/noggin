# Issue tracker: minibeads

Issues and specs for this repo live in [minibeads](https://github.com/rrnewton/minibeads), a dependency-aware tracker that stores each issue as a markdown file under `.minibeads/issues/`, committed to git. Use the `mb` CLI for all operations. (The binary is `mb`; there is no `minibeads` command.)

## Conventions

- **Initialise (once)**: already done with `mb init --prefix tk`. Issue IDs are `tk-<n>`; a bare number like `14` expands to `tk-14` in `show`.
- **Create an issue**: `mb create "<title>" -t <bug|feature|task|epic|chore> -p <0-4> -l <label> --description-file -` with the body on stdin (heredoc). Always pass multi-line or markdown bodies via `--description-file`, never `-d "..."`: the shell expands backticks and `$(...)` inside inline arguments and the mangled text is stored silently. Optional fields: `--design`, `--acceptance`, `--parent <id>`, `--deps blocks:<id>`.
- **Read an issue**: `mb show <id>` (includes labels, dependencies, and comments). Add `--json` for structured output.
- **List issues**: `mb list --status open [-l <label>] [--type <type>] [--parent <id>]`. Without `--status`, closed issues are included. Add `--json` to get bodies and labels.
- **Comment on an issue**: `mb comments add <id> --body "..."`; `mb comments list <id>`.
- **Apply / remove labels**: `mb update <id> --add-label "..."` / `--remove-label "..."` (or `mb label add|remove <id> <label>`).
- **Edit text**: never hand-edit `.minibeads/*.md`. Use `mb update <id> --append "..."` to add to the end of a field, or `--search "<exact old text>" --replace "<new>"` for a targeted edit (the match must be unique). Pick the field with `--field description|design|notes|acceptance`. Avoid `--description` for rewrites.
- **Close**: `mb comments add <id> --body "..."` then `mb close <id> --reason "..."`. Re-open with `mb reopen <id>`.
- **Persist**: `.minibeads/` is tracked in git. Commit it with the code change it belongs to, but do not push unless instructed to.

Triage state is recorded as **labels** (see `triage-labels.md` for the role strings), because minibeads statuses (`open`, `in_progress`, `blocked`, `closed`) are workflow states, not triage roles. Newly created issues get `needs-triage`.

## Specs and tickets

- A **spec** is an issue of type `epic` labelled `spec`, with the spec text as its description.
- **Implementation tickets** are child issues of the spec: `mb create "<title>" --parent <spec-id> ...`. List them with `mb children <spec-id>`.
- **Blocking edges** are native: `mb dep add <blocked-id> <blocker-id>` (the first issue depends on the second; type `blocks` is the default). Check for mistakes with `mb dep cycles` and `mb dep tree <id>`.
- **Next unblocked work**: `mb ready [--parent <spec-id>] [-l ready-for-agent]`. Ready means `open` with no open blockers; claimed (`in_progress`) issues are excluded. `mb blocked` shows what is waiting and on what.

## When a skill says "publish to the issue tracker"

Run `mb create` with the body piped through `--description-file -`, applying the appropriate labels and `--parent` / `mb dep add` edges, then commit `.minibeads/`.

## When a skill says "fetch the relevant ticket"

Run `mb show <id>`. The user will normally pass the ID (or just its number).

## Wayfinding operations

Used by `/wayfinder`. The **map** is an `epic` issue with **child** issues as tickets.

- **Map**: a single issue of type `epic` labelled `wayfinder:map`, holding the Notes / Decisions-so-far / Fog body in its description. `mb create "<title>" -t epic -l wayfinder:map --description-file -`.
- **Child ticket**: `mb create "<title>" --parent <map-id> -l wayfinder:<type> --description-file -`, where `<type>` is `research`/`prototype`/`grilling`/`task`. The question goes in the body.
- **Blocking**: `mb dep add <child-id> <blocker-id>`. A ticket is unblocked when every blocker is closed.
- **Frontier query**: `mb ready --parent <map-id>`. It returns children that are open, unblocked, and unclaimed; first in the output wins. (`--sort oldest` gives creation order.)
- **Claim**: `mb claim <id>` (sets assignee and `in_progress`, 48h lease; `--for 4h` to shorten), the session's first write. Where several workers share the repo, commit and push immediately (`git commit -am "claim <id>" && git push`); a rejected push means someone else got there first, so pull and pick another ticket. Release with `mb claim <id> --release`.
- **Resolve**: `mb comments add <id> --body "<answer>"`, then `mb close <id>`, then append a context pointer (gist + link) to the map's Decisions-so-far with `mb update <map-id> --append "- <gist> (<id>)"`.
