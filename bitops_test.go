package bitmatrix

import "testing"

func TestSet_Has(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(57403, 42)
	if !bm.Has(57403, 42) {
		t.Error("Has returned false after Set")
	}
}

func TestHas_UnsetBit(t *testing.T) {
	bm := newTestMatrix()
	if bm.Has(0, 0) {
		t.Error("Has returned true on zero-initialised matrix")
	}
}

func TestClear(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(100, 5)
	bm.Clear(100, 5)
	if bm.Has(100, 5) {
		t.Error("Has returned true after Clear")
	}
}

func TestClear_UnsetBit(t *testing.T) {
	bm := newTestMatrix()
	bm.Clear(0, 0) // clearing an already-clear bit must not panic or corrupt
	if bm.Has(0, 0) {
		t.Error("Has returned true after clearing an unset bit")
	}
}

func TestWordBoundaries(t *testing.T) {
	bm := newTestMatrix()
	rows := []uint32{0, 63, 64, 65, 127, 128, 129}
	col := uint32(0)
	for _, row := range rows {
		bm.Set(row, col)
		if !bm.Has(row, col) {
			t.Errorf("row %d: Has false after Set", row)
		}
		bm.Clear(row, col)
		if bm.Has(row, col) {
			t.Errorf("row %d: Has true after Clear", row)
		}
	}
}

func TestChunkBoundaries(t *testing.T) {
	bm := newTestMatrix()
	rows := []uint32{
		ChunkSize - 1,
		ChunkSize,
		ChunkSize + 1,
		2*ChunkSize - 1,
		2 * ChunkSize,
	}
	col := uint32(0)
	for _, row := range rows {
		bm.Set(row, col)
		if !bm.Has(row, col) {
			t.Errorf("row %d: Has false after Set", row)
		}
		bm.Clear(row, col)
		if bm.Has(row, col) {
			t.Errorf("row %d: Has true after Clear", row)
		}
	}
}

func TestSet_DoesNotAffectNeighbours(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(64, 0)
	if bm.Has(63, 0) || bm.Has(65, 0) {
		t.Error("Set(64) bled into adjacent rows")
	}
	if bm.Has(64, 1) {
		t.Error("Set(row=64, col=0) bled into col 1")
	}
}
