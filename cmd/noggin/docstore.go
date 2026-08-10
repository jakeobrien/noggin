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
// each document record carries two fields. It has no attachment to Roots at
// all (D61) — a record is addressed purely by its document's absolute path,
// independent of which Root, if any, currently covers it.
const docStoreName = "documents"

// docRecord is one document's persisted state (S1): whether it is starred
// and when it was last viewed. A starred document that has never been
// viewed, or one whose view timestamp M1.3 has since evicted, has a zero
// LastViewed — the two fields are independent, not "viewed implies
// timestamp present".
type docRecord struct {
	Starred    bool
	LastViewed time.Time
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

// loadDocStore reads the document store at path: one record per non-blank,
// non-`#`-comment line, tab-separated as `<starred>\t<lastViewed>\t<path>`,
// where <starred> is `1` or `0` and <lastViewed> is RFC3339Nano or `-` for
// no timestamp. A missing file, an unreadable one, or a line that fails to
// parse (wrong field count, bad starred flag, unparsable timestamp, or a
// non-absolute path) is dropped rather than failing the whole load, the same
// tolerance loadStore already gives the Roots store. A later line for a
// path already seen replaces the earlier one (last write wins), matching a
// hand-edited file's most likely intent.
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
		fields := strings.SplitN(line, "\t", 3)
		if len(fields) != 3 {
			continue
		}
		starredField, lastViewedField, docPath := fields[0], fields[1], fields[2]
		if !filepath.IsAbs(docPath) {
			continue
		}
		var starred bool
		switch starredField {
		case "1":
			starred = true
		case "0":
			starred = false
		default:
			continue
		}
		var lastViewed time.Time
		if lastViewedField != "-" {
			t, err := time.Parse(time.RFC3339Nano, lastViewedField)
			if err != nil {
				continue
			}
			lastViewed = t
		}
		records[filepath.Clean(docPath)] = docRecord{Starred: starred, LastViewed: lastViewed}
	}
	return records
}

// evictDocStore bounds records to at most docStoreCap entries (S4),
// mutating the given map in place and returning it for chaining. Candidates
// are walked oldest-`LastViewed`-first (a zero LastViewed sorts as oldest);
// an unstarred candidate is deleted outright, while a starred candidate is
// never removed — eviction instead clears its LastViewed, leaving the star
// — and does not count toward shrinking the map below the cap. Walking
// stops as soon as the map is within cap, so a starred record encountered
// only after the cap is already satisfied is left completely untouched
// (star and timestamp both). If there are not enough unstarred records to
// reach the cap, the map is left larger than docStoreCap rather than ever
// removing a starred record — the star guarantee outranks the bound.
func evictDocStore(records map[string]docRecord, cap int) map[string]docRecord {
	if len(records) <= cap {
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
		if remaining <= cap {
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
// directory first if it does not exist yet. Map iteration order is
// randomized by Go itself, but the store is read back into a map on load,
// so on-disk line order carries no meaning.
func saveDocStore(path string, records map[string]docRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for docPath, rec := range records {
		starredField := "0"
		if rec.Starred {
			starredField = "1"
		}
		lastViewedField := "-"
		if !rec.LastViewed.IsZero() {
			lastViewedField = rec.LastViewed.Format(time.RFC3339Nano)
		}
		b.WriteString(starredField)
		b.WriteByte('\t')
		b.WriteString(lastViewedField)
		b.WriteByte('\t')
		b.WriteString(docPath)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
