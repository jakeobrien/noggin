//go:build unix

package tree

import (
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"
)

// A FIFO has no content to sniff and opening it blocks: All must hide it
// rather than hang the walk.
func TestFilterAllHidesAFifoWithoutBlocking(t *testing.T) {
	root := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(root, "pipe"), 0o644); err != nil {
		t.Skip("mkfifo unavailable:", err)
	}
	type result struct {
		node *Node
		err  error
	}
	done := make(chan result, 1)
	go func() {
		node, err := Walk(root, FilterAll)
		done <- result{node, err}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			t.Fatal(r.err)
		}
		if slices.Contains(childNames(r.node), "pipe") {
			t.Errorf("All lists a FIFO: %v", childNames(r.node))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Walk under All blocked on a FIFO")
	}
}
