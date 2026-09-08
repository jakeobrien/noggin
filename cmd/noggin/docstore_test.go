package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

// TestDocStoreMixedFileArticleRoundTrip is C6/M1.5's own acceptance
// wording: article records (a canonical-URL key plus Starred/LastViewed and
// a Title) round-trip alongside file records in the same store, each kind
// keeping exactly its own fields — an article's Title survives, a file
// record's Article/Title both stay at the zero value.
func TestDocStoreMixedFileArticleRoundTrip(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notes.md")
	const articleURL = "https://en.wikipedia.org/wiki/Go_(programming_language)"
	path := filepath.Join(dir, "documents")

	now := time.Now().Round(time.Nanosecond)
	want := map[string]docRecord{
		filePath:   {Starred: true, LastViewed: now},
		articleURL: {Starred: true, LastViewed: now.Add(-time.Hour), Article: true, Title: "Go (programming language)"},
	}
	if err := saveDocStore(path, want); err != nil {
		t.Fatalf("saveDocStore: %v", err)
	}

	got := loadDocStore(path)
	if len(got) != len(want) {
		t.Fatalf("loadDocStore returned %d records, want %d: %+v", len(got), len(want), got)
	}

	gotFile, ok := got[filePath]
	if !ok {
		t.Fatalf("loadDocStore missing the file record for %s: %+v", filePath, got)
	}
	if gotFile.Article || gotFile.Title != "" {
		t.Errorf("file record = %+v, want Article false and Title \"\"", gotFile)
	}
	if !gotFile.Starred || !gotFile.LastViewed.Equal(now) {
		t.Errorf("file record = %+v, want Starred true, LastViewed %v", gotFile, now)
	}

	gotArticle, ok := got[articleURL]
	if !ok {
		t.Fatalf("loadDocStore missing the article record for %s: %+v", articleURL, got)
	}
	if !gotArticle.Article || gotArticle.Title != "Go (programming language)" {
		t.Errorf("article record = %+v, want Article true, Title %q", gotArticle, "Go (programming language)")
	}
	if !gotArticle.Starred || !gotArticle.LastViewed.Equal(now.Add(-time.Hour)) {
		t.Errorf("article record = %+v, want Starred true, LastViewed %v", gotArticle, now.Add(-time.Hour))
	}
}

// TestDocStoreArticleTitleTabAndNewlineRoundTrip is M1.7's acceptance
// criterion (TRIAGE T3): an article title carrying an embedded tab, and one
// carrying an embedded newline, both round-trip through save/load byte-for-
// byte, without corrupting the record itself (the title read back must
// still match) or the adjacent record sharing the file — a raw tab or
// newline in an unescaped title would otherwise shift loadDocStore's fixed
// tab-separated field positions, or split one line into two.
func TestDocStoreArticleTitleTabAndNewlineRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "documents")
	const (
		tabURL     = "https://en.wikipedia.org/wiki/Tab_Title"
		newlineURL = "https://en.wikipedia.org/wiki/Newline_Title"
		plainURL   = "https://en.wikipedia.org/wiki/Plain_Title"
	)
	tabTitle := "Embedded\tTab"
	newlineTitle := "Embedded\nNewline"
	want := map[string]docRecord{
		tabURL:     {Article: true, Title: tabTitle},
		newlineURL: {Article: true, Title: newlineTitle},
		plainURL:   {Article: true, Title: "Plain Title", Starred: true},
	}
	if err := saveDocStore(path, want); err != nil {
		t.Fatalf("saveDocStore: %v", err)
	}

	// The escaped on-disk bytes must still be exactly one line per record
	// (no raw newline split a title across lines) and exactly 5 tab-
	// separated fields per line (no raw tab shifted a field boundary) —
	// checked directly against the file, not just inferred from a
	// successful load, so a subtly-broken line/field count that happened
	// to still parse could not slip past this test unnoticed.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading store file: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != len(want) {
		t.Fatalf("on-disk store has %d lines, want %d (a raw tab/newline likely escaped incorrectly): %q", len(lines), len(want), raw)
	}
	for _, line := range lines {
		if n := strings.Count(line, "\t"); n != 4 {
			t.Errorf("line %q has %d tab-separated field boundaries, want 4 (5 fields)", line, n)
		}
	}
	if !strings.Contains(string(raw), `Embedded\tTab`) {
		t.Errorf("on-disk store does not contain the escaped tab title %q: %q", `Embedded\tTab`, raw)
	}
	if !strings.Contains(string(raw), `Embedded\nNewline`) {
		t.Errorf("on-disk store does not contain the escaped newline title %q: %q", `Embedded\nNewline`, raw)
	}

	got := loadDocStore(path)
	if len(got) != len(want) {
		t.Fatalf("loadDocStore returned %d records, want %d: %+v", len(got), len(want), got)
	}
	for key, wantRec := range want {
		gotRec, ok := got[key]
		if !ok {
			t.Fatalf("loadDocStore missing record for %s: %+v", key, got)
		}
		if gotRec.Title != wantRec.Title || gotRec.Article != wantRec.Article || gotRec.Starred != wantRec.Starred {
			t.Errorf("record for %s = %+v, want %+v", key, gotRec, wantRec)
		}
	}
}

// TestLoadDocStoreArticleRequiresTitle checks S1's "pointer — canonical URL
// and title", both halves required: an article line with an empty <title>
// is dropped like any other malformed line, never kept with a blank title.
func TestLoadDocStoreArticleRequiresTitle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "documents")
	const articleURL = "https://en.wikipedia.org/wiki/Orphan"
	content := "1\t-\tarticle\t\t" + articleURL + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadDocStore(path)
	if len(got) != 0 {
		t.Errorf("loadDocStore(article, no title) = %+v, want empty (dropped)", got)
	}
}

// TestLoadDocStoreLegacyFileFormatByteForByteEquivalent is C6/M1.5's own
// "existing on-disk stores load unchanged" acceptance clause: a store
// written entirely in the pre-M1.5, three-field-per-line shape loads into
// exactly the same docRecord values it always has — Starred and LastViewed
// as before, Article/Title at their new zero value, nothing about a legacy
// file's own records changed by M1.5 existing.
func TestLoadDocStoreLegacyFileFormatByteForByteEquivalent(t *testing.T) {
	dir := t.TempDir()
	starredViewed := filepath.Join(dir, "starred-viewed.md")
	unstarredUnviewed := filepath.Join(dir, "unstarred-unviewed.md")
	path := filepath.Join(dir, "documents")

	viewedAt := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	// Exactly the legacy, pre-M1.5 <starred>\t<lastViewed>\t<path> shape —
	// no <kind>/<title> fields at all.
	content := "1\t" + viewedAt.Format(time.RFC3339Nano) + "\t" + starredViewed + "\n" +
		"0\t-\t" + unstarredUnviewed + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got := loadDocStore(path)
	want := map[string]docRecord{
		starredViewed:     {Starred: true, LastViewed: viewedAt},
		unstarredUnviewed: {Starred: false},
	}
	if len(got) != len(want) {
		t.Fatalf("loadDocStore returned %d records, want %d: %+v", len(got), len(want), got)
	}
	for p, wantRec := range want {
		gotRec, ok := got[p]
		if !ok {
			t.Errorf("loadDocStore missing record for %s", p)
			continue
		}
		if !reflect.DeepEqual(gotRec, wantRec) {
			t.Errorf("%s: loadDocStore = %+v, want byte-for-byte equivalent %+v", p, gotRec, wantRec)
		}
	}
}

// TestEvictDocStoreMixedFileArticleStore is C6/M1.5's own eviction clause:
// the cap and the starred-never-evicted rule apply across file and article
// records together, as one pool ordered by LastViewed — not per kind. With
// cap 2 and, oldest to newest, an unstarred article, an unstarred file and
// a starred article, the two unstarred records compete for eviction by
// recency regardless of kind, and the starred article survives regardless
// of kind too.
func TestEvictDocStoreMixedFileArticleStore(t *testing.T) {
	base := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	const oldArticle = "https://en.wikipedia.org/wiki/Old"
	const starredArticle = "https://en.wikipedia.org/wiki/Starred"
	records := map[string]docRecord{
		oldArticle:     {LastViewed: base, Article: true, Title: "Old"},
		"/file.md":     {LastViewed: base.Add(time.Minute)},
		starredArticle: {Starred: true, LastViewed: base.Add(2 * time.Minute), Article: true, Title: "Starred"},
	}

	got := evictDocStore(records, 2)

	if len(got) != 2 {
		t.Fatalf("evictDocStore left %d records, want 2: %+v", len(got), got)
	}
	if _, ok := got[oldArticle]; ok {
		t.Errorf("%s (oldest, unstarred) should have been evicted", oldArticle)
	}
	fileRec, ok := got["/file.md"]
	if !ok || fileRec.Article {
		t.Errorf("/file.md should survive untouched (newer than the evicted article), got %+v, ok=%v", fileRec, ok)
	}
	articleRec, ok := got[starredArticle]
	if !ok {
		t.Fatalf("%s (starred) must survive eviction", starredArticle)
	}
	if !articleRec.Starred || !articleRec.Article || articleRec.Title != "Starred" {
		t.Errorf("%s survived but lost its own fields: %+v", starredArticle, articleRec)
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
