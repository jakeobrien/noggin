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
	done := make(chan *Node, 1)
	go func() {
		node, _ := Walk(root, FilterAll)
		done <- node
	}()
	select {
	case node := <-done:
		if slices.Contains(childNames(node), "pipe") {
			t.Errorf("All lists a FIFO: %v", childNames(node))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Walk under All blocked on a FIFO")
	}
}
