package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// starredInRoot lists the starred files under rootPath in folder order, the
// order the Files Nav Tree shows them (folderOrderLess). Articles and
// starred files missing on disk (broken stars) are left out: neither can be
// loaded from the Root. The File Filter and View toggles are ignored, as
// they are for the Starred Browse Mode.
func starredInRoot(docs map[string]docRecord, rootPath string) []string {
	var paths []string
	for path, rec := range docs {
		if !rec.Starred || rec.Article || !strictlyInside(rootPath, path) || pathMissing(path) {
			continue
		}
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool { return folderOrderLess(rootPath, paths[i], paths[j]) })
	return paths
}

// folderOrderLess orders two paths under rootPath the way tree.Walk orders
// the Tree, flattened depth-first: within a folder, subfolders come before
// files, then each by name.
func folderOrderLess(rootPath, a, b string) bool {
	ra, errA := filepath.Rel(rootPath, a)
	rb, errB := filepath.Rel(rootPath, b)
	if errA != nil || errB != nil {
		return a < b
	}
	pa := strings.Split(ra, string(filepath.Separator))
	pb := strings.Split(rb, string(filepath.Separator))
	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] == pb[i] {
			continue
		}
		aDir, bDir := i < len(pa)-1, i < len(pb)-1
		if aDir != bDir {
			return aDir
		}
		return pa[i] < pb[i]
	}
	return len(pa) < len(pb)
}

// nextStarred picks the document to cycle to from current through paths
// (already in folder order), stepping dir (+1 or -1) and wrapping. From a
// document not in paths, +1 lands on the first and -1 on the last. ok is
// false when paths is empty.
func nextStarred(paths []string, current string, dir int) (string, bool) {
	n := len(paths)
	if n == 0 {
		return "", false
	}
	for i, p := range paths {
		if p == current {
			return paths[((i+dir)%n+n)%n], true
		}
	}
	if dir < 0 {
		return paths[n-1], true
	}
	return paths[0], true
}

// cycleStarred is keys.StarredNext/keys.StarredPrevious's action: it loads
// the Active Root's next (dir +1) or previous (dir -1) starred document into
// the focused Document Viewer. It is a navigation like Activate: it records
// recency and pushes onto that viewer's Document History. A silent no-op
// while Peeking, and when the only starred document is the one already
// shown. With Wikipedia as the Active Source, no Active Root, or no starred
// document in it, the Footer says so.
func (m model) cycleStarred(dir int) model {
	p := m.focusedViewerPane()
	v := m.viewerFor(p)
	if v.peeking() {
		return m
	}
	if m.roots.activeIsWikipedia() {
		m.message = "starred documents cycle within a Root; Wikipedia is the Active Source"
		return m
	}
	rootPath := m.roots.activePath()
	if rootPath == "" {
		m.message = "no Active Root to cycle starred documents in"
		return m
	}
	current := ""
	if v.doc.isOpen() {
		current = v.doc.path
	}
	target, ok := nextStarred(starredInRoot(m.docs, rootPath), current, dir)
	if !ok {
		m.message = fmt.Sprintf("no starred documents in %s", m.activeRootLabel(rootPath))
		return m
	}
	if target == current {
		return m
	}
	d, err := loadDocument(target)
	if err != nil {
		m.message = cannotOpen(target, err)
		return m
	}
	m = m.recordViewed(target)
	return m.navigate(p, d)
}

// activeRootLabel is the Active Root's display name, for Footer messages.
func (m model) activeRootLabel(rootPath string) string {
	for _, r := range m.roots.rootList() {
		if r.path == rootPath {
			return r.displayName()
		}
	}
	return filepath.Base(rootPath)
}
