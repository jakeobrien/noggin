package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// docStoreCap bounds the total number of records the document store may
// hold (S4). Fixed, not configurable (S1's Out-of-scope list). 1000 is
// 10x the per-Root `Recent` view cap (S6's 100), generous for a store that
// spans every Root a session has ever touched while keeping the on-disk
// file (well under 100KB at this row size) and any future in-memory scan of
// it trivially cheap (D68).
const docStoreCap = 1000

// docStoreName is the document store's file name, alongside the starred-
// Roots store under the same data directory (dataDir, M1.1) but a distinct
// file and format (M1.2): unlike the Roots store's one-path-per-line list,
// each document record carries several fields (docRecord). It has no
// attachment to Roots at all (D61) — a record is addressed purely by its
// own key, a file's absolute path or (C6/M1.5) an article's canonical URL,
// independent of which Root, if any, covers a file's path.
const docStoreName = "documents"

// docRecord is one document's persisted state (S1): whether it is starred
// and when it was last viewed. A starred document that has never been
// viewed, or one whose view timestamp M1.3 has since evicted, has a zero
// LastViewed — the two fields are independent, not "viewed implies
// timestamp present".
//
// Article and Title are C6/M1.5's own addition: a file record's key (the
// map's own key in loadDocuments/docRecord's callers) is already its
// absolute path, everything a file needs to be found or displayed, so a
// file record carries neither field — Article is false and Title is "",
// the zero value, for every record this store held before M1.5 and every
// one loadDocStore's legacy 3-field branch below still produces. An article
// record's key is instead its own canonical URL (M4.1: "the article's ...
// store key is its canonical URL"), which names the article but not its
// title, so Title carries that second half of S1's own "pointer — canonical
// URL and title" definition; Article is true so callers can tell the two
// record kinds apart without inspecting the key's own shape (a URL vs. an
// absolute path) themselves. The store's count cap (docStoreCap) and
// starred-never-evicted rule (evictDocStore) apply to every record the same
// way regardless of kind — neither function reads Article or Title at all.
type docRecord struct {
	Starred    bool
	LastViewed time.Time
	Article    bool
	Title      string
}

// docStorePath resolves the document store's path via dataDir (M1.1). An
// error here (no resolvable data directory) is treated by callers the same
// as an absent store, mirroring storePath's degrade for the starred-Roots
// store.
func docStorePath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, storeSubdir, docStoreName), nil
}

// loadDocuments resolves the document store path via docStorePath and loads
// it, folding a non-resolvable data directory into the same "no persisted
// documents" outcome as a missing or garbage file — mirroring
// loadStarredRoots for the S1 store. main.go's sole entry point for the
// model's startup load (model.withDocs, M1.4); nothing else calls it.
func loadDocuments() map[string]docRecord {
	path, err := docStorePath()
	if err != nil {
		return map[string]docRecord{}
	}
	return loadDocStore(path)
}

// loadDocStore reads the document store at path: one record per non-blank,
// non-`#`-comment line, tab-separated, in one of two shapes.
//
// The legacy, pre-M1.5 shape — `<starred>\t<lastViewed>\t<path>`, exactly
// three fields — is a file record keyed by its own absolute path; every
// store on disk before C6/M1.5 is written entirely in this shape, and this
// branch reproduces the exact same docRecord (Article/Title left at their
// zero value) it always has, so an existing store loads unchanged.
//
// The C6/M1.5 shape — `<starred>\t<lastViewed>\t<kind>\t<title>\t<key>`,
// exactly five fields — adds <kind> (`file` or `article`) and <title> ahead
// of the same trailing key field, so a file record can be told apart from
// an article one without inspecting the key's own shape. `file` behaves
// exactly like the legacy shape (an absolute-path key, <title> ignored —
// saveDocStore always writes it empty for a file record, but a hand-edited
// value here changes nothing a file record reads); `article` requires a
// non-empty <key> (S4's canonical URL) and non-empty <title> (S1's "pointer
// — canonical URL and title", both halves required) or the line is dropped,
// the same tolerance a malformed file line already gets.
//
// Either shape's <starred>/<lastViewed> fields keep their existing meaning
// (`1`/`0`; RFC3339Nano or `-` for no timestamp). A missing file, an
// unreadable one, or a line that fails to parse in either shape (wrong
// field count, unknown <kind>, bad starred flag, unparsable timestamp, a
// non-absolute file path, or a missing article key/title) is dropped rather
// than failing the whole load, the same tolerance loadStore already gives
// the Roots store. A later line for a key already seen replaces the
// earlier one (last write wins), matching a hand-edited file's most likely
// intent.
func loadDocStore(path string) map[string]docRecord {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]docRecord{}
	}
	records := map[string]docRecord{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.SplitN(line, "\t", 5)
		if len(fields) != 3 && len(fields) != 5 {
			continue
		}
		starred, lastViewed, ok := parseDocStoreCommonFields(fields[0], fields[1])
		if !ok {
			continue
		}
		switch len(fields) {
		case 3:
			docPath := fields[2]
			if !filepath.IsAbs(docPath) {
				continue
			}
			records[filepath.Clean(docPath)] = docRecord{Starred: starred, LastViewed: lastViewed}
		case 5:
			kind, title, key := fields[2], fields[3], fields[4]
			switch kind {
			case "file":
				if !filepath.IsAbs(key) {
					continue
				}
				records[filepath.Clean(key)] = docRecord{Starred: starred, LastViewed: lastViewed}
			case "article":
				if key == "" || title == "" {
					continue
				}
				records[key] = docRecord{Starred: starred, LastViewed: lastViewed, Article: true, Title: title}
			default:
				continue
			}
		default:
			continue
		}
	}
	return records
}

// parseDocStoreCommonFields parses the <starred>/<lastViewed> pair every
// loadDocStore line shape shares (legacy and C6/M1.5 alike), reporting
// false when either fails to parse — the one piece of line-validation both
// of loadDocStore's field-count branches need, so it lives here rather than
// duplicated in each.
func parseDocStoreCommonFields(starredField, lastViewedField string) (starred bool, lastViewed time.Time, ok bool) {
	switch starredField {
	case "1":
		starred = true
	case "0":
		starred = false
	default:
		return false, time.Time{}, false
	}
	if lastViewedField != "-" {
		t, err := time.Parse(time.RFC3339Nano, lastViewedField)
		if err != nil {
			return false, time.Time{}, false
		}
		lastViewed = t
	}
	return starred, lastViewed, true
}

// evictDocStore bounds records to at most docStoreCap entries (S4),
// mutating the given map in place and returning it for chaining. Candidates
// are walked oldest-`LastViewed`-first (a zero LastViewed sorts as oldest);
// an unstarred candidate is deleted outright, while a starred candidate is
// never removed — eviction instead clears its LastViewed, leaving the star
// — and does not count toward shrinking the map below the limit. Walking
// stops as soon as the map is within limit, so a starred record encountered
// only after the limit is already satisfied is left completely untouched
// (star and timestamp both). If there are not enough unstarred records to
// reach the limit, the map is left larger than docStoreCap rather than ever
// removing a starred record — the star guarantee outranks the bound.
func evictDocStore(records map[string]docRecord, limit int) map[string]docRecord {
	if len(records) <= limit {
		return records
	}

	type entry struct {
		path string
		rec  docRecord
	}
	entries := make([]entry, 0, len(records))
	for p, r := range records {
		entries = append(entries, entry{p, r})
	}
	sort.Slice(entries, func(i, j int) bool {
		if !entries[i].rec.LastViewed.Equal(entries[j].rec.LastViewed) {
			return entries[i].rec.LastViewed.Before(entries[j].rec.LastViewed)
		}
		return entries[i].path < entries[j].path // deterministic tie-break
	})

	remaining := len(records)
	for _, e := range entries {
		if remaining <= limit {
			break
		}
		if e.rec.Starred {
			if !e.rec.LastViewed.IsZero() {
				cleared := e.rec
				cleared.LastViewed = time.Time{}
				records[e.path] = cleared
			}
			continue
		}
		delete(records, e.path)
		remaining--
	}
	return records
}

// saveDocStore writes records to the store at path, creating the data
// directory first if it does not exist yet. Every record is written in
// C6/M1.5's five-field shape (loadDocStore's own doc comment), a file
// record's <title> field simply left empty — there is no need to keep
// writing the legacy three-field shape once this binary can read both, and
// a single shape going forward is one less case for a hand-editor (or a
// future reader) to keep straight. Map iteration order is randomized by Go
// itself, but the store is read back into a map on load, so on-disk line
// order carries no meaning.
func saveDocStore(path string, records map[string]docRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for key, rec := range records {
		starredField := "0"
		if rec.Starred {
			starredField = "1"
		}
		lastViewedField := "-"
		if !rec.LastViewed.IsZero() {
			lastViewedField = rec.LastViewed.Format(time.RFC3339Nano)
		}
		kind := "file"
		title := ""
		if rec.Article {
			kind = "article"
			title = rec.Title
		}
		b.WriteString(starredField)
		b.WriteByte('\t')
		b.WriteString(lastViewedField)
		b.WriteByte('\t')
		b.WriteString(kind)
		b.WriteByte('\t')
		b.WriteString(title)
		b.WriteByte('\t')
		b.WriteString(key)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
