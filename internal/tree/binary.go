package tree

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"
)

// Open refusals (D140). A file noggin will not open is reported as one of
// these, so every open route can show the same Footer message.
var (
	// ErrBinary marks a file with a NUL byte in its first sniffLen bytes —
	// ripgrep's and git's test for binary content.
	ErrBinary = errors.New("binary file")
	// ErrTooLarge marks a file over MaxOpenSize bytes.
	ErrTooLarge = errors.New("too large")
)

const (
	// sniffLen is how much of a file's head is checked for a NUL byte.
	sniffLen = 8 << 10
	// MaxOpenSize is the largest file noggin will open (~10MB).
	MaxOpenSize = 10 << 20
	// sniffCacheMax bounds the sniff cache; it is simply reset when full. It
	// is sized for a File Filter `All` walk, which sniffs every non-Markdown
	// file in a Root (a few MB at this cap), not for one-off opens.
	sniffCacheMax = 1 << 16
)

type sniffKey struct {
	path  string
	mtime time.Time
	size  int64
}

var sniffCache = struct {
	sync.Mutex
	m map[sniffKey]bool
}{m: map[sniffKey]bool{}}

// CheckOpenable reports whether noggin will open the file at path: nil, or
// ErrTooLarge / ErrBinary. Any other error is the underlying stat/read
// failure. The binary sniff reads the head of the file once and is cached by
// (path, mtime, size), so repeated opens of an unchanged file cost one stat.
func CheckOpenable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil // let the read report the directory error as it always has
	}
	if info.Size() > MaxOpenSize {
		return ErrTooLarge
	}
	binary, err := cachedBinary(path, info)
	if err != nil {
		return err
	}
	if binary {
		return ErrBinary
	}
	return nil
}

// cachedBinary is the binary sniff of path, cached by (path, mtime, size) of
// the info the caller just took, so an unchanged file is sniffed once.
func cachedBinary(path string, info os.FileInfo) (bool, error) {
	key := sniffKey{path, info.ModTime(), info.Size()}
	sniffCache.Lock()
	binary, hit := sniffCache.m[key]
	sniffCache.Unlock()
	if hit {
		return binary, nil
	}
	binary, err := sniffBinary(path)
	if err != nil {
		return false, err
	}
	sniffCache.Lock()
	if len(sniffCache.m) >= sniffCacheMax {
		clear(sniffCache.m)
	}
	sniffCache.m[key] = binary
	sniffCache.Unlock()
	return binary, nil
}

// IsNotBinary reports whether the file at path is a non-binary regular file:
// what a File Filter `All` listing and the Picker's Starred scope keep. It is
// a regular file (never a FIFO, socket or a symlink to a directory, which a
// sniff could block on or fail) whose head has no NUL byte. A file that cannot
// be sniffed is kept, so opening it reports the real error; one that no longer
// exists (a dangling symlink, a file deleted mid-walk) is dropped. Size is not
// considered: an oversize file is listed and refused when opened
// (CheckOpenable).
func IsNotBinary(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return !errors.Is(err, fs.ErrNotExist)
	}
	if !info.Mode().IsRegular() {
		return false
	}
	binary, err := cachedBinary(path, info)
	return err != nil || !binary
}

// sniffBinary reports whether the first sniffLen bytes of path contain a NUL.
func sniffBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, err
	}
	return bytes.IndexByte(buf[:n], 0) >= 0, nil
}
