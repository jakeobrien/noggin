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
// right-hand label per S6 — either the covering session Root's own label
// (root.displayName, S8/M7.3 — its persisted Name if it has one, else the
// folder's base name), or, when no session Root does (an "orphan"
// document, S1's "not attached to Roots"), that document's own parent
// directory, left-truncated at render time the same way the Sources Nav
// already disambiguates two same-named Roots (rootsnav.go's leftTruncate).
type browseRow struct {
	path string
	name string
	// rootLabel is the covering Root's own label (root.displayName, S8/M7.3)
	// when orphan is false; meaningless (and left "") otherwise.
	rootLabel string
	// orphanDir is the document's own parent directory when orphan is true
	// (render truncates it, this field never is, since the budget is only
	// known at render width); meaningless (and left "") otherwise.
	orphanDir string
	orphan    bool
	// broken is S3's "broken starred document" mark (rootRow.broken's own
	// definition, pathMissing, reused verbatim): the file cannot be stat'd.
	// recentRows (C6/M6.1) and starredRows (M4.3) both never produce a
	// row for a missing file at all: recentRows omits it (S6's "silently
	// omitted"), while starredRows shows it as a *broken* row instead of
	// dropping it, since a star carries user intent a timestamp does not —
	// so this is only ever true from starredRows' own file branch.
	// Always false for an article row (below) — an article has no on-disk
	// existence to stat at all; S3 never names a "broken" state for one,
	// only a fetch that can fail at Activate/Peek time, which is a Footer
	// message (article_fetch.go), not a static row mark.
	broken bool
	// article is C6/M5.2's own addition, widened to Recent by C6/M6.1, set
	// only by buildDocBrowseRow: a Wikipedia article rather than a file
	// (docRecord.Article) — name already holds its display title
	// (buildDocBrowseRow, not filepath.Base(path), since path here is the
	// article's canonical URL, M1.5's own store key), and model.go's
	// activateFlatRowInto/togglePeekFlatRowInto read this to route through
	// fetchArticleInto (S3) instead of loadDocument, which would just fail
	// reading a URL as a file (D129's own named landmine, M4.3's precedent
	// for Refresh). Mirrors pickerFile's own article/title pair
	// (files_search.go).
	article bool
}

// buildBrowseRow derives one browseRow for path against roots (S6): the
// covering session Root's own label (root.displayName, S8/M7.3), or the
// parent-directory/orphan form when no Root covers it, plus whether the
// file is currently missing on disk (S3's broken mark) — harmless to
// compute unconditionally, since recentRows only ever calls this after its
// own existence check already passed, and starredRows needs it every time.
func buildBrowseRow(path string, roots []root) browseRow {
	row := browseRow{path: path, name: filepath.Base(path), broken: pathMissing(path)}
	if r, ok := rootForPath(roots, path); ok {
		row.rootLabel = r.displayName()
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
// (S4), so it never appears regardless of scope; a file record whose file no
// longer exists is silently omitted (S6, unlike a broken star, which is
// still shown — M4.3's own concern) — an article record has no on-disk
// existence to check at all (buildDocBrowseRow's own doc comment), so it is
// never omitted this way (C6/M6.1). Every other record is bucketed by which
// session Root covers its path via rootForPath, with a "" key for the
// no-Root ("orphan") bucket — a document last viewed while covered by a
// Root no longer in the session, or never covered by one at all — except an
// article record, which buckets under the reserved wikipediaSourceKey
// sentinel instead (starredRows' own precedent, C6/M5.2): an article's path
// is a URL, which rootForPath can never match, so without its own bucket it
// would silently fall into the "" orphan bucket alongside a genuine
// Root-orphan file. Each bucket is ordered newest-first and capped at
// recentPerRootCap (S6's own per-Root cap) independently, before scope is
// applied:
//
//   - Active Source scope reads the bucket keyed by activeRootPath, or —
//     wikipedia (true only while the Files Nav's own Active Source is the
//     Wikipedia Query Source, files.go's f.wikipedia, C6/M6.1) — the article
//     bucket instead, the same activeRootPath-is-always-""-while-Wikipedia's-
//     -active distinction starredRows' own Active Source branch already
//     makes (nil when neither applies, i.e. no root is active — there is no
//     bucket for "no active root" to fall back to).
//   - All Roots scope merges every bucket's own (already-capped) list,
//     orphans and articles both included, sorted by timestamp with no
//     further cap (S6/S7) — this is where Recent's own two right-hand label
//     forms (S6) both appear, since Recent's scope set (S5) has no dedicated
//     Orphans value the way Starred's does.
func recentRows(docs map[string]docRecord, roots []root, activeRootPath string, wikipedia bool, scope browseScope) []browseRow {
	buckets := map[string][]docTimestamp{}
	for path, rec := range docs {
		if rec.LastViewed.IsZero() {
			continue
		}
		key := ""
		switch {
		case rec.Article:
			key = wikipediaSourceKey
		default:
			if pathMissing(path) {
				continue
			}
			if r, ok := rootForPath(roots, path); ok {
				key = r.path
			}
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
		switch {
		case wikipedia:
			merged = buckets[wikipediaSourceKey]
		case activeRootPath != "":
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
		rows = append(rows, buildDocBrowseRow(e.path, docs[e.path], roots))
	}
	return rows
}

// buildDocBrowseRow derives one flat-list browseRow for path (C6/M5.2,
// widened from Starred-only to also serve recentRows at C6/M6.1): rec.Article
// routes to the article's own display shape — name is the article's title
// (rec.Title, S1's own store-record title, never filepath.Base(path): path
// here is the article's canonical URL, M1.5's store key, with no filesystem
// basename worth showing), rootLabel is the fixed wikipediaLabel
// (rootsnav.go's own Wikipedia Source constant, S6's "labelled Wikipedia
// where files are labelled by Root name"), and orphan/broken both stay
// false — pathMissing/rootForPath would both give a URL a meaningless
// answer, and S3 never gives an article a static "broken" mark at all
// (browseRow.article's own doc comment). A file record delegates to
// buildBrowseRow unchanged — the same fields Recent's and Starred's own rows
// both use.
func buildDocBrowseRow(path string, rec docRecord, roots []root) browseRow {
	if rec.Article {
		return browseRow{path: path, name: rec.Title, rootLabel: wikipediaLabel, article: true}
	}
	return buildBrowseRow(path, roots)
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
//     sortByRecency's own tie-break) — an article's own name is its title
//     (buildDocBrowseRow), so it sorts alongside files by that same key.
//
// Buckets key on the same rootForPath-derived Root path, "" for a document
// under no session Root (an "orphan", S1), or — C6/M5.2 — the reserved
// wikipediaSourceKey sentinel (files.go's own per-Source memory key, reused
// here for the same "impossible as a real Root path" property) for every
// starred article: an article's path is a URL, which rootForPath can never
// match, so without its own bucket every article would silently fall into
// the "" orphan bucket alongside genuine Root-orphan files (the cosmetic
// gap M5.1's own handoff named) — wrong for two reasons S5/S6 both state:
// Orphans is a Root-only concept that must never include an article, and
// Active Source while Wikipedia is active must show only articles, never a
// stray Root-orphan file caught in the same bucket by accident. wikipedia
// (true only while the Files Nav's own Active Source is the Wikipedia Query
// Source, files.go's f.wikipedia) is what lets Active Source tell those two
// ""-shaped meanings apart: with it set, Active Source reads the article
// bucket instead of activeRootPath's (always "" while Wikipedia is active,
// filesNav.setWikipedia) — the same distinction recentRows' own Active
// Source branch now makes too (C6/M6.1). All Starred merges every bucket,
// files and articles alike (S6); Orphans (the pre-agreed cut 3 in PLAN.md/
// M4.3, carried regardless — see browse.go's modeScopes, which already
// excludes it entirely while Wikipedia is active) reads just the "" bucket
// on its own, which the article bucket's own key never collides with.
func starredRows(docs map[string]docRecord, roots []root, activeRootPath string, wikipedia bool, scope browseScope) []browseRow {
	buckets := map[string][]browseRow{}
	for path, rec := range docs {
		if !rec.Starred {
			continue
		}
		key := ""
		switch {
		case rec.Article:
			key = wikipediaSourceKey
		default:
			if r, ok := rootForPath(roots, path); ok {
				key = r.path
			}
		}
		buckets[key] = append(buckets[key], buildDocBrowseRow(path, rec, roots))
	}

	var rows []browseRow
	switch scope {
	case activeRootBrowseScope:
		if wikipedia {
			rows = buckets[wikipediaSourceKey]
		} else if activeRootPath != "" {
			rows = buckets[activeRootPath]
		}
	case orphansBrowseScope:
		rows = buckets[""]
	default: // allStarredBrowseScope
		for _, entries := range buckets {
			rows = append(rows, entries...)
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].name != rows[j].name {
			return rows[i].name < rows[j].name
		}
		return rows[i].path < rows[j].path
	})
	return rows
}
