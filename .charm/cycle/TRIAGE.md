# Triage — C6 sources

<!-- Per-cycle triage INBOX for actionable items surfaced out-of-band: code-review and
     validate findings, bugs / in-scope issues discovered during build that were not
     anticipated in the plan, work deliberately skipped per the escalation matrix (high,
     non-blocking, work-shaped), and the consequences of plan-invalidating discoveries.
     NOT a backlog and NOT a decision log (judgment questions go to DECISIONS.md).

     Only charm-replan parses items in full; every other role reads this inbox at summary
     level (count, severities, items referenced by blocked tasks).

     Lifecycle: consume-and-clear. Items are added `open`, then triaged by charm-replan per
     the escalation matrix — promoted to a PLAN.md task (with an acceptance criterion),
     deferred beyond this cycle into ROADMAP.md Unscheduled, recorded as a DECISIONS.md entry
     (wontfix), or dropped with a reason. Once triaged, the item is removed here (the outcome
     lives in PLAN / ROADMAP / DECISIONS). This doc stays small by construction; if it is
     growing, that means triage is overdue. It must be empty before a cycle can close.

     Keep one line per item. Ids (T1, T2…) are transient scratch labels, not durable keys —
     they are only for referring to an item until it is triaged, and may be reused after
     the inbox is cleared. Cite like findings: filepath:line. Severity is the reporter's
     estimate, not ground truth — triage filters false positives and wontfix.

     Sources: review | validate | build | docs | user. Status: open | triaged. -->

| Id | Source | Severity | Citation | Item | Status |
|----|--------|----------|----------|------|--------|
| T1 | build | low | cmd/noggin/model.go:47 (`pane.title()`, `case rootsPane: return "Roots"`) | GLOSSARY.md already renames the pane "Sources Nav" (C6), but no M1.x task's acceptance criterion covers retitling the border/footer label from "Roots" to "Sources"; M1.2 (Wikipedia row) and M1.3 (Active Root→Active Source labels) both leave it untouched by their own done-when clauses. | open |
| T2 | review | low | cmd/noggin/rootsnav.go:30, files.go:876/886, keymap.go:21/34/68/76/179/235/247, model.go:39/109/126/231/621/648/668/1308/1359/1378/1395/1454/1666/1812, footer.go:44/47/48/61, layout.go:4/19, help.go:45, roots.go:12, browse_list.go:20, style.go:6, picker_view.go:293 | Pre-existing (not introduced by M1.1-M1.5): GLOSSARY.md's "Roots Nav — Renamed Sources Nav in C6" has not propagated into code comments anywhere in the reviewed diff's surrounding files — "Roots Nav" still appears throughout, none renamed by M1.1/M1.2/M1.3 despite those tasks touching several of these same files. This is the same root issue as T1 (open) but shows the naming drift is codebase-wide in comments, not just the one pane-title string; worth folding into whichever task eventually closes T1, or a dedicated cleanup pass, rather than treating T1's scope as just the one line. | open |
| T3 | review | low | cmd/noggin/docstore.go:120 (`loadDocStore`'s `strings.SplitN(line, "\t", 5)`), docstore.go:267 (`saveDocStore`'s unescaped `b.WriteString(title)`) | An article Title containing a literal tab character corrupts the record on save/reload: SplitN's greedy final field absorbs the extra tab(s), shifting <title> and <key> apart (title truncated, key gains a stray tab-joined suffix); a Title containing a newline would split into two unparseable lines outright. No field is escaped or validated against embedded `\t`/`\n` before writing, and no test exercises a title with either. M4.1 onward will populate Title from real Wikipedia article titles (which can legally contain a mid-dash or other punctuation, though a literal tab is unlikely) — worth either rejecting/escaping such characters when a Title is set, or asserting Wikipedia titles never contain them, before M1.5's format is relied on further. | open |
