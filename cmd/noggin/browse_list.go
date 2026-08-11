package main

import (
	"path/filepath"
	"sort"
	"time"
)

// recentPerRootCap bounds how many documents Recent's own per-Root list
// carries (S6's 100) before All Roots ever merges it with any other Root's
// own list — the cap docstore.go's docStoreCap doc comment already sizes
// itself against ("10x the per-Root Recent view cap").
const recentPerRootCap = 100

// browseRow is one rendered row in a Files Nav flat list (Recent, M4.2;
// Starred, M4.3): a document's own path, its display name, and its
// right-hand label per S6 — either the base name of the session Root that
// covers it, or, when no session Root does (an "orphan" document, S1's "not
// attached to Roots"), that document's own parent directory, left-truncated
// at render time the same way the Roots Nav already disambiguates two
// same-named Roots (rootsnav.go's leftTruncate).
type browseRow struct {
	path string
	name string
	// rootLabel is the covering Root's base name when orphan is false;
	// meaningless (and left "") otherwise.
	rootLabel string
	// orphanDir is the document's own parent directory when orphan is true
	// (render truncates it, this field never is, since the budget is only
	// known at render width); meaningless (and left "") otherwise.
	orphanDir string
	orphan    bool
	// broken is S3's "broken starred document" mark (rootRow.broken's own
	// definition, pathMissing, reused verbatim): the file cannot be stat'd.
	// recentRows never produces a row for a missing file at all (S6's
	// "silently omitted"), so this is always false there; starredRows
	// (M4.3) is the one caller that shows a broken row rather than
	// dropping it, since a star carries user intent a timestamp does not.
	broken bool
}

// buildBrowseRow derives one browseRow for path against roots (S6): the
// covering session Root's base name, or the parent-directory/orphan form
// when no Root covers it, plus whether the file is currently missing on
// disk (S3's broken mark) — harmless to compute unconditionally, since
// recentRows only ever calls this after its own existence check already
// passed, and starredRows needs it every time.
func buildBrowseRow(path string, roots []root) browseRow {
	row := browseRow{path: path, name: filepath.Base(path), broken: pathMissing(path)}
	if r, ok := rootForPath(roots, path); ok {
		row.rootLabel = filepath.Base(r.path)
	} else {
		row.orphan = true
		row.orphanDir = filepath.Dir(path)
	}
	return row
}

// docTimestamp pairs a document's path with the timestamp it is being
// ordered by — recentRows' own working type, never exposed past this file.
type docTimestamp struct {
	path string
	ts   time.Time
}

// sortByRecency orders entries newest-first, breaking an exact timestamp
// tie by path for a deterministic result (docstore.go's evictDocStore uses
// the same tie-break, oldest-first, for the same reason: map iteration
// order is otherwise unspecified).
func sortByRecency(entries []docTimestamp) {
	sort.Slice(entries, func(i, j int) bool {
		if !entries[i].ts.Equal(entries[j].ts) {
			return entries[i].ts.After(entries[j].ts)
		}
		return entries[i].path < entries[j].path
	})
}

// recentRows builds S5/S6's Recent-mode flat list for scope from the S1
// document store's LastViewed timestamps and the session's own Roots list
// (roots, in rootsNav.rootList order; activeRootPath is rootsNav.activePath,
// "" when no root is active).
//
// A record with a zero LastViewed has never been (Activate/$EDITOR-)viewed
// (S4), so it never appears regardless of scope; a record whose file no
// longer exists is silently omitted (S6, unlike a broken star, which is
// still shown — M4.3's own concern). Every other record is bucketed by
// which session Root covers its path via rootForPath, with a "" key for the
// no-Root ("orphan") bucket — a document last viewed while covered by a
// Root no longer in the session, or never covered by one at all. Each
// bucket is ordered newest-first and capped at recentPerRootCap (S6's own
// per-Root cap) independently, before scope is applied:
//
//   - Active Root scope reads only the bucket keyed by activeRootPath (nil
//     when activeRootPath is "", i.e. no root is active — there is no
//     bucket for "no active root" to fall back to).
//   - All Roots scope merges every bucket's own (already-capped) list,
//     orphans included, sorted by timestamp with no further cap (S6) —
//     this is where Recent's own two right-hand label forms (S6) both
//     appear, since Recent's scope set (S5) has no dedicated Orphans value
//     the way Starred's does.
func recentRows(docs map[string]docRecord, roots []root, activeRootPath string, scope browseScope) []browseRow {
	buckets := map[string][]docTimestamp{}
	for path, rec := range docs {
		if rec.LastViewed.IsZero() || pathMissing(path) {
			continue
		}
		key := ""
		if r, ok := rootForPath(roots, path); ok {
			key = r.path
		}
		buckets[key] = append(buckets[key], docTimestamp{path: path, ts: rec.LastViewed})
	}
	for key, entries := range buckets {
		sortByRecency(entries)
		if len(entries) > recentPerRootCap {
			entries = entries[:recentPerRootCap]
		}
		buckets[key] = entries
	}

	var merged []docTimestamp
	if scope == activeRootBrowseScope {
		if activeRootPath != "" {
			merged = buckets[activeRootPath]
		}
	} else {
		for _, entries := range buckets {
			merged = append(merged, entries...)
		}
		sortByRecency(merged)
	}

	rows := make([]browseRow, 0, len(merged))
	for _, e := range merged {
		rows = append(rows, buildBrowseRow(e.path, roots))
	}
	return rows
}

// starredRows builds S5/S6's Starred-mode flat list for scope from the S1
// document store's Starred paths and the session's own Roots list —
// recentRows' shape (M4.2), with the differences S3/S5/S6 call for a
// Starred list rather than a Recent one:
//
//   - every Starred record appears regardless of LastViewed (S1: Starred and
//     LastViewed are independent fields; a starred-but-never-viewed document
//     still belongs in every Starred scope);
//   - a record whose file no longer exists is marked broken (buildBrowseRow
//     already computes this) rather than omitted — S3's "broken starred
//     document ... is still listed and still marked, but cannot be
//     activated", the opposite of Recent's own S6 omission rule;
//   - there is no per-Root cap (recentPerRootCap is Recent's own S6 bound,
//     named nowhere in S5's Starred scope table) and no timestamp to order
//     by, so rows sort by name (tie-broken by path for determinism, mirroring
//     sortByRecency's own tie-break).
//
// Buckets key on the same rootForPath-derived Root path, or "" for a
// document under no session Root (an "orphan", S1) — recentRows' identical
// bucketing, just never capped. Active Root reads only its own bucket
// (activeRootPath == "" short-circuits to none, same as recentRows); All
// Starred merges every bucket, orphans included; Orphans (the pre-agreed cut
// 3 in PLAN.md/M4.3, carried regardless — see browse.go's modeScopes) reads
// just the "" bucket on its own.
func starredRows(docs map[string]docRecord, roots []root, activeRootPath string, scope browseScope) []browseRow {
	buckets := map[string][]string{}
	for path, rec := range docs {
		if !rec.Starred {
			continue
		}
		key := ""
		if r, ok := rootForPath(roots, path); ok {
			key = r.path
		}
		buckets[key] = append(buckets[key], path)
	}

	var paths []string
	switch scope {
	case activeRootBrowseScope:
		if activeRootPath != "" {
			paths = buckets[activeRootPath]
		}
	case orphansBrowseScope:
		paths = buckets[""]
	default: // allStarredBrowseScope
		for _, entries := range buckets {
			paths = append(paths, entries...)
		}
	}

	sort.Slice(paths, func(i, j int) bool {
		ni, nj := filepath.Base(paths[i]), filepath.Base(paths[j])
		if ni != nj {
			return ni < nj
		}
		return paths[i] < paths[j]
	})

	rows := make([]browseRow, 0, len(paths))
	for _, p := range paths {
		rows = append(rows, buildBrowseRow(p, roots))
	}
	return rows
}
