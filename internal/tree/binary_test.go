package tree

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckOpenable(t *testing.T) {
	dir := t.TempDir()
	late := make([]byte, sniffLen+10) // NUL past the sniff window: not binary
	for i := 0; i < sniffLen+5; i++ {
		late[i] = 'a'
	}
	big := make([]byte, MaxOpenSize+1)
	for i := range big {
		big[i] = 'a'
	}
	cases := []struct {
		name string
		data []byte
		want error
	}{
		{"text.txt", []byte("hello\n"), nil},
		{"empty", nil, nil},
		{"bin.dat", []byte("ab\x00cd"), ErrBinary},
		{"nulfirst", []byte{0}, ErrBinary},
		{"late", late, nil},
		{"big.txt", big, ErrTooLarge},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := filepath.Join(dir, c.name)
			write(t, p, c.data)
			if got := CheckOpenable(p); !errors.Is(got, c.want) && got != c.want {
				t.Errorf("CheckOpenable = %v, want %v", got, c.want)
			}
		})
	}
	if err := CheckOpenable(filepath.Join(dir, "missing")); !os.IsNotExist(err) {
		t.Errorf("missing file: err = %v, want not-exist", err)
	}
	if err := CheckOpenable(dir); err != nil {
		t.Errorf("directory: err = %v, want nil (read reports it)", err)
	}
}

// TestCheckOpenableSniffCache: the verdict is cached by (path, mtime, size) —
// an unchanged key is not re-read, and a changed mtime or size is.
func TestCheckOpenableSniffCache(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f")
	write(t, p, []byte("plain\n"))
	old := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(p, old, old); err != nil {
		t.Fatal(err)
	}
	if err := CheckOpenable(p); err != nil {
		t.Fatal(err)
	}
	// Same size, same mtime, different bytes: the cached verdict stands.
	write(t, p, []byte("pl\x00in\n"))
	if err := os.Chtimes(p, old, old); err != nil {
		t.Fatal(err)
	}
	if err := CheckOpenable(p); err != nil {
		t.Errorf("same (path, mtime, size): err = %v, want cached nil", err)
	}
	// A new mtime misses the cache and re-sniffs.
	now := old.Add(time.Minute)
	if err := os.Chtimes(p, now, now); err != nil {
		t.Fatal(err)
	}
	if err := CheckOpenable(p); !errors.Is(err, ErrBinary) {
		t.Errorf("new mtime: err = %v, want ErrBinary", err)
	}
	// A new size misses it too.
	write(t, p, []byte("plain text again\n"))
	if err := os.Chtimes(p, now, now); err != nil {
		t.Fatal(err)
	}
	if err := CheckOpenable(p); err != nil {
		t.Errorf("new size: err = %v, want nil", err)
	}
}
