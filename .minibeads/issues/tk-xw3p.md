---
title: 'Follow-up: tidy File Filter / Picker bounds after tk-ybh'
status: closed
priority: 3
issue_type: chore
depends_on:
  tk-0ta: parent-child
created_at: 2026-10-14T20:50:31.566121839+00:00
updated_at: 2026-10-15T19:08:09.745499880+00:00
closed_at: 2026-10-15T19:08:09.733661297+00:00
---

# Description

## Parent

tk-0ta (Spec: Markdown/Text Render Mode and File Filter, D140)

## Context

Came from the code review of tk-ybh (one store for the File Filter, commit 75d04e3). These are judgement-call smells the review raised and the prefactor left alone. None changes behaviour.

## Items

- **`pickerResults` keeps `hideDotfiles` and `filter` loose.** They are the "searched under" label that `resultsCurrent` compares against `pickerBounds`. They are patched by hand in three places: `toggleHiddenFiles`, `applyPickerFilter` (the Wikipedia/Starred relabel) and `applyPickerResults`. Embedding a `pickerBounds` in `pickerResults` would make `resultsCurrent` a single `==` and leave one patch site.
- **`pickerSearchCmd` and `dispatchPickerSearch` take `hideDotfiles` and the filter as two loose arguments.** They could take `pickerBounds` instead.
- **`togglePickerFilter` branches on `m.files.root == nil` itself.** It writes `lastFilter` when there is no Root, while `fileFilter()` reads it. A single `filesNav` method that sets the filter for the Active Root, or the fallback when there is no Root, would keep that rule in one place.
- **`withFilter` also sets `lastFilter` as a side effect**, which the name does not hint at.
- **Naming:** the type `pickerBounds` and the method `model.pickerBounds()` share a name. `fileFilter()`, `filterFor(path)`, `lastFilter` and `withLastFilter` are near-synonyms. `filterFor` is still called directly in the walk paths, which is correct but unclear.
- **`applyPickerFilter(next)`** now uses `next` only to patch `results.filter`. The parameter looks redundant with `fileFilter()`.
- **Test assertions:** "reopened Picker filter" and "Picker filter after the Files Nav's `a`" now check `filesNav`, not any Picker state. They could assert on `pickerBounds()` or the rendered title instead.
- **Existing duplication, not from this change:** `pickerPreviewCandidate`, `pickerGrepCandidate` and `pickerWikipediaCandidate` share one shape (check mode, check `resultsCurrent`, bounds-check `selected`).

## Acceptance criteria

- [ ] Triage which items are worth doing. Drop the rest as `wontfix`, or split them into their own tickets.
- [ ] `go test ./...` stays green, with no user-visible change.
