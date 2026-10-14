package tree

import (
	"bytes"
	"errors"
	"io"
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
	// sniffCacheMax bounds the sniff cache; it is simply reset when full.
	sniffCacheMax = 4096
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
	key := sniffKey{path, info.ModTime(), info.Size()}
	sniffCache.Lock()
	binary, hit := sniffCache.m[key]
	sniffCache.Unlock()
	if !hit {
		if binary, err = sniffBinary(path); err != nil {
			return err
		}
		sniffCache.Lock()
		if len(sniffCache.m) >= sniffCacheMax {
			clear(sniffCache.m)
		}
		sniffCache.m[key] = binary
		sniffCache.Unlock()
	}
	if binary {
		return ErrBinary
	}
	return nil
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
