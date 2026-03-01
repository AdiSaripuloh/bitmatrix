package bitmatrix

import (
	"sync"
	"testing"
)

func TestGrow(t *testing.T) {
	bm := New(WithRows(8192), WithCols(4))
	bm.Set(8191, 0) // last row of first chunk

	bm.Grow(16384)

	if bm.rows != 16384 {
		t.Errorf("rows after Grow = %d, want 16384", bm.rows)
	}
	// Existing data must be preserved.
	if !bm.Has(8191, 0) {
		t.Error("Grow lost existing data at row 8191")
	}
	// New rows must be zero.
	if bm.Has(8192, 0) {
		t.Error("Grow produced non-zero bit in new rows")
	}
}

func TestGrow_NoOp(t *testing.T) {
	bm := New(WithRows(8192), WithCols(4))
	bm.Set(0, 0)
	bm.Grow(4096) // smaller than current rows
	if !bm.Has(0, 0) {
		t.Error("Grow with smaller value corrupted existing data")
	}
	if bm.rows != 8192 {
		t.Errorf("rows changed after no-op Grow: got %d, want 8192", bm.rows)
	}
}

// TestGrow_ConcurrentWithLock verifies that Grow is safe when protected by an
// external mutex, as documented. Run with -race to detect data races.
func TestGrow_ConcurrentWithLock(t *testing.T) {
	bm := New(WithRows(64), WithCols(4))

	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		newRows := uint32((i + 2) * 64)
		go func() {
			defer wg.Done()
			mu.Lock()
			bm.Grow(newRows)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if bm.rows < 10*64 {
		t.Errorf("rows after concurrent Grow = %d, want >= %d", bm.rows, 10*64)
	}
	// Verify the matrix is functional after concurrent grows.
	bm.Set(bm.rows-1, 0)
	if !bm.Has(bm.rows-1, 0) {
		t.Error("Set/Has failed on last row after concurrent Grow")
	}
}
