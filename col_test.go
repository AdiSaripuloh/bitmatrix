package bitmatrix

import "testing"

func TestCountInCol(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(0, 5)
	bm.Set(100, 5)
	bm.Set(999_999, 5)
	if got := bm.CountInCol(5); got != 3 {
		t.Errorf("CountInCol = %d, want 3", got)
	}
}

func TestClearCol(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(0, 3)
	bm.Set(1000, 3)
	bm.Set(999_999, 3)
	bm.ClearCol(3)
	if got := bm.CountInCol(3); got != 0 {
		t.Errorf("CountInCol after ClearCol = %d, want 0", got)
	}
}

func TestSetColChunk(t *testing.T) {
	bm := newTestMatrix()
	bm.SetColChunk(0, 0)

	want := int(bm.chunkSize)
	if got := bm.CountInCol(0); got != want {
		t.Errorf("CountInCol after SetColChunk = %d, want %d", got, want)
	}
	// Chunk 1 must be untouched.
	if bm.Has(bm.chunkSize, 0) {
		t.Errorf("SetColChunk(chunk=0) bled into chunk 1")
	}
}

func TestCountInCol_RespectsRows(t *testing.T) {
	bm := New(WithRows(100), WithCols(4))
	bm.SetColChunk(0, 0) // sets ChunkSize entities in underlying storage
	// CountInCol must only count the 100 valid rows, not the full chunk.
	if got := bm.CountInCol(0); got != 100 {
		t.Errorf("CountInCol = %d, want 100 (rows)", got)
	}
}

func TestCountInCol_WordAlignedRows(t *testing.T) {
	bm := New(WithRows(128), WithCols(4))
	bm.SetColChunk(0, 0)
	if got := bm.CountInCol(0); got != 128 {
		t.Errorf("CountInCol = %d, want 128", got)
	}
}
