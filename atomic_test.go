package bitmatrix

import (
	"sync"
	"testing"
)

func TestAtomicBitMatrix_SetHasClear(t *testing.T) {
	abm := NewAtomic(WithRows(1_000_000), WithCols(64))
	abm.Set(42, 7)
	if !abm.Has(42, 7) {
		t.Error("atomic Has false after Set")
	}
	abm.Clear(42, 7)
	if abm.Has(42, 7) {
		t.Error("atomic Has true after Clear")
	}
}

func TestAtomicBitMatrix_Ensure_WithinBounds(t *testing.T) {
	abm := NewAtomic(WithRows(1000), WithCols(4))
	abm.Ensure(500, 1)
	if !abm.Has(500, 1) {
		t.Error("Ensure within bounds: Has false after Ensure")
	}
	if abm.Rows() != 1000 {
		t.Errorf("Ensure within bounds changed row count: got %d, want 1000", abm.Rows())
	}
}

func TestAtomicBitMatrix_Ensure_OutOfBounds(t *testing.T) {
	abm := NewAtomic(WithRows(64), WithCols(4))
	abm.Set(10, 0)
	abm.Ensure(999, 2)
	if !abm.Has(999, 2) {
		t.Error("Ensure out of bounds: Has false after Ensure")
	}
	if abm.Rows() < 1000 {
		t.Errorf("Ensure did not grow matrix: rows = %d, want >= 1000", abm.Rows())
	}
	// Existing data must be preserved.
	if !abm.Has(10, 0) {
		t.Error("Ensure lost existing data after grow")
	}
}

func TestAtomicBitMatrix_Concurrent(t *testing.T) {
	abm := NewAtomic(WithRows(1_000_000), WithCols(64))
	const goroutines = 100
	const col = uint32(0)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		row := uint32(i)
		go func() {
			defer wg.Done()
			abm.Set(row, col)
		}()
	}
	wg.Wait()

	for i := 0; i < goroutines; i++ {
		if !abm.Has(uint32(i), col) {
			t.Errorf("concurrent Set: row %d not set", i)
		}
	}
}
