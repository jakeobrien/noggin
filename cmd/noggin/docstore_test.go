package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// TestLoadDocStoreMissingFile: a store file that has never been written
// yields an empty map, not an error (M1.2, mirroring S2's loadStore
// tolerance).
func TestLoadDocStoreMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "documents")
	got := loadDocStore(path)
	if len(got) != 0 {
		t.Errorf("loadDocStore(missing) = %v, want empty", got)
	}
}

// TestLoadDocStoreGarbageFile: lines that don't parse — wrong field count, a
// bad starred flag, an unparsable timestamp, or a relative path — are
// dropped, not fatal to the whole load; a well-formed line elsewhere in the
// same file still survives.
func TestLoadDocStoreGarbageFile(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.md")
	path := filepath.Join(dir, "documents")
	content := "not a line\n" +
		"\x00\x01binary noise\n" +
		"2\t-\t" + good + "\n" + // bad starred flag
		"1\tnot-a-time\t" + good + "\n" + // unparsable timestamp
		"1\t-\trelative/path.md\n" + // relative path
		"1\t-\n" + // wrong field count
		"1\t-\t" + good + "\n" // the one good line
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadDocStore(path)
	want := map[string]docRecord{good: {Starred: true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadDocStore(garbage) = %+v, want %+v", got, want)
	}
}

// TestLoadDocStoreIgnoresCommentsAndBlanks checks blank lines and `#`
// comments are skipped, matching the Roots store's hand-editable spirit.
func TestLoadDocStoreIgnoresCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.md")
	path := filepath.Join(dir, "documents")
	content := "# documents\n\n1\t-\t" + a + "\n\n# trailing comment\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadDocStore(path)
	want := map[string]docRecord{a: {Starred: true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadDocStore = %+v, want %+v", got, want)
	}
}

// TestDocStoreRoundTrip: saving a mix of starred/unstarred and
// timestamped/untimestamped records and reloading yields byte-for-byte
// equivalent records (timestamps compared via Equal since RFC3339Nano
// round-trips exactly but time.Time zone representation may not be
// identical for a wall-clock-constructed value).
func TestDocStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	starredViewed := filepath.Join(dir, "starred-viewed.md")
	starredUnviewed := filepath.Join(dir, "starred-unviewed.md")
	unstarredViewed := filepath.Join(dir, "unstarred-viewed.md")
	path := filepath.Join(dir, "documents")

	now := time.Now().Round(time.Nanosecond)
	want := map[string]docRecord{
		starredViewed:   {Starred: true, LastViewed: now},
		starredUnviewed: {Starred: true},
		unstarredViewed: {Starred: false, LastViewed: now.Add(-time.Hour)},
	}
	if err := saveDocStore(path, want); err != nil {
		t.Fatalf("saveDocStore: %v", err)
	}

	got := loadDocStore(path)
	if len(got) != len(want) {
		t.Fatalf("loadDocStore returned %d records, want %d: %+v", len(got), len(want), got)
	}
	for p, wantRec := range want {
		gotRec, ok := got[p]
		if !ok {
			t.Errorf("loadDocStore missing record for %s", p)
			continue
		}
		if gotRec.Starred != wantRec.Starred {
			t.Errorf("%s: Starred = %v, want %v", p, gotRec.Starred, wantRec.Starred)
		}
		if !gotRec.LastViewed.Equal(wantRec.LastViewed) {
			t.Errorf("%s: LastViewed = %v, want %v", p, gotRec.LastViewed, wantRec.LastViewed)
		}
	}
}

// TestSaveDocStoreCreatesDataDir: saveDocStore makes the parent directory
// when it does not exist yet, so a first-ever star or view does not fail.
func TestSaveDocStoreCreatesDataDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "noggin", "documents")
	doc := filepath.Join(t.TempDir(), "a.md")
	if err := saveDocStore(path, map[string]docRecord{doc: {Starred: true}}); err != nil {
		t.Fatalf("saveDocStore: %v", err)
	}
	got := loadDocStore(path)
	want := map[string]docRecord{doc: {Starred: true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadDocStore after save = %+v, want %+v", got, want)
	}
}

// TestDocStoreNoAttachmentToRoots is M1.2's "no attachment to Roots" clause:
// reading a record for a path under no session Root at all succeeds exactly
// like any other path — the store has no notion of Roots to check the path
// against.
func TestDocStoreNoAttachmentToRoots(t *testing.T) {
	dir := t.TempDir()
	orphan := filepath.Join(dir, "no-root-covers-this", "orphan.md")
	path := filepath.Join(dir, "documents")
	if err := saveDocStore(path, map[string]docRecord{orphan: {Starred: true}}); err != nil {
		t.Fatalf("saveDocStore: %v", err)
	}

	got := loadDocStore(path)
	rec, ok := got[orphan]
	if !ok {
		t.Fatalf("loadDocStore missing orphan record: %+v", got)
	}
	if !rec.Starred {
		t.Errorf("orphan record Starred = false, want true")
	}
}

// TestEvictDocStoreBoundsAndProtectsStars is M1.3's acceptance criterion in
// one scenario: with cap 3 and 5 records ordered oldest-to-newest
// starred(s1), unstarred(u1), unstarred(u2), starred(s2), unstarred(u3),
// eviction must (a) drop the oldest *unstarred* records first (u1, u2),
// (b) never remove a starred record (s1 and s2 both survive), and (c) clear
// the timestamp of a starred record that was old enough to be an eviction
// candidate (s1) while leaving a starred record reached only after the cap
// was already satisfied (s2) completely untouched.
func TestEvictDocStoreBoundsAndProtectsStars(t *testing.T) {
	base := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	records := map[string]docRecord{
		"/s1": {Starred: true, LastViewed: base},
		"/u1": {Starred: false, LastViewed: base.Add(1 * time.Minute)},
		"/u2": {Starred: false, LastViewed: base.Add(2 * time.Minute)},
		"/s2": {Starred: true, LastViewed: base.Add(3 * time.Minute)},
		"/u3": {Starred: false, LastViewed: base.Add(4 * time.Minute)},
	}

	got := evictDocStore(records, 3)

	if len(got) != 3 {
		t.Fatalf("evictDocStore left %d records, want 3: %+v", len(got), got)
	}
	if _, ok := got["/u1"]; ok {
		t.Errorf("/u1 (oldest unstarred) should have been evicted")
	}
	if _, ok := got["/u2"]; ok {
		t.Errorf("/u2 (2nd-oldest unstarred) should have been evicted")
	}
	s1, ok := got["/s1"]
	if !ok {
		t.Fatalf("/s1 (starred) must survive eviction")
	}
	if !s1.Starred {
		t.Errorf("/s1 lost its star")
	}
	if !s1.LastViewed.IsZero() {
		t.Errorf("/s1 (evicted-from-recency starred record) should have its timestamp cleared, got %v", s1.LastViewed)
	}
	s2, ok := got["/s2"]
	if !ok {
		t.Fatalf("/s2 (starred) must survive eviction")
	}
	if !s2.Starred || !s2.LastViewed.Equal(base.Add(3*time.Minute)) {
		t.Errorf("/s2 (reached only once cap already satisfied) should be untouched, got %+v", s2)
	}
	u3, ok := got["/u3"]
	if !ok || u3.Starred || !u3.LastViewed.Equal(base.Add(4*time.Minute)) {
		t.Errorf("/u3 (newest) should be untouched, got %+v, ok=%v", u3, ok)
	}
}

// TestEvictDocStoreNoopUnderCap: a store already within cap is returned
// unchanged.
func TestEvictDocStoreNoopUnderCap(t *testing.T) {
	records := map[string]docRecord{
		"/a": {Starred: false, LastViewed: time.Now()},
	}
	got := evictDocStore(records, 5)
	if !reflect.DeepEqual(got, records) {
		t.Errorf("evictDocStore under cap = %+v, want unchanged %+v", got, records)
	}
}

// TestEvictDocStoreAllStarredExceedsCap: when every record is starred and
// count already exceeds cap, eviction leaves the store over cap rather than
// ever removing a starred record.
func TestEvictDocStoreAllStarredExceedsCap(t *testing.T) {
	records := map[string]docRecord{
		"/a": {Starred: true, LastViewed: time.Now()},
		"/b": {Starred: true, LastViewed: time.Now().Add(-time.Hour)},
	}
	got := evictDocStore(records, 1)
	if len(got) != 2 {
		t.Errorf("evictDocStore(all starred) = %d records, want 2 (cap exceeded, no star removed)", len(got))
	}
}

// TestDocStorePath pins docStorePath's resolution through dataDir (M1.1),
// under its own file name distinct from the starred-Roots store.
func TestDocStorePath(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	got, err := docStorePath()
	if err != nil {
		t.Fatalf("docStorePath: %v", err)
	}
	want := filepath.Join(dataHome, storeSubdir, docStoreName)
	if got != want {
		t.Errorf("docStorePath() = %q, want %q", got, want)
	}
	if docStoreName == storeName {
		t.Errorf("docStoreName must not collide with the Roots store's storeName")
	}
}
