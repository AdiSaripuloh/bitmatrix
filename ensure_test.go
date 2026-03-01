package bitmatrix

import "testing"

func TestEnsure_WithinBounds(t *testing.T) {
	bm := New(WithRows(1000), WithCols(4))
	bm.Ensure(500, 1)
	if !bm.Has(500, 1) {
		t.Error("Ensure within bounds: Has false after Ensure")
	}
	if bm.rows != 1000 {
		t.Errorf("Ensure within bounds changed row count: got %d, want 1000", bm.rows)
	}
}

func TestEnsure_OutOfBounds(t *testing.T) {
	bm := New(WithRows(64), WithCols(4))
	bm.Ensure(999, 2)
	if !bm.Has(999, 2) {
		t.Error("Ensure out of bounds: Has false after Ensure")
	}
	if bm.rows < 1000 {
		t.Errorf("Ensure did not grow matrix: rows = %d, want >= 1000", bm.rows)
	}
}

func TestEnsure_AtBoundary(t *testing.T) {
	bm := New(WithRows(64), WithCols(4))
	bm.Ensure(64, 0) // row == bm.rows, exactly one past the end
	if !bm.Has(64, 0) {
		t.Error("Ensure at boundary: Has false after Ensure")
	}
}

func TestEnsure_RowZeroOnEmpty(t *testing.T) {
	bm := New(WithRows(MinRows), WithCols(4))
	bm.Ensure(0, 0)
	if !bm.Has(0, 0) {
		t.Error("Ensure row 0: Has false after Ensure")
	}
}

func TestEnsure_PreservesExistingData(t *testing.T) {
	bm := New(WithRows(64), WithCols(4))
	bm.Set(10, 1)
	bm.Ensure(500, 2) // triggers grow
	if !bm.Has(10, 1) {
		t.Error("Ensure lost existing data after grow")
	}
	if !bm.Has(500, 2) {
		t.Error("Ensure did not set the new bit")
	}
}

func TestEnsure_DoesNotAffectNeighbours(t *testing.T) {
	bm := New(WithRows(64), WithCols(4))
	bm.Ensure(999, 1)
	if bm.Has(998, 1) || bm.Has(1000, 1) {
		t.Error("Ensure bled into adjacent rows")
	}
	if bm.Has(999, 0) || bm.Has(999, 2) {
		t.Error("Ensure bled into adjacent columns")
	}
}
