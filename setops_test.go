package bitmatrix

import "testing"

func TestColAnd(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(1, 0)
	bm.Set(1, 1)
	bm.Set(2, 0) // only col 0

	result := bm.ColAnd(0, 1)
	if result.Count() != 1 {
		t.Errorf("ColAnd count = %d, want 1", result.Count())
	}
	if !result.Has(1) {
		t.Error("ColAnd missing entity 1")
	}
	if result.Has(2) {
		t.Error("ColAnd contains entity 2 (only has col 0)")
	}
}

func TestColOr(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(1, 0)
	bm.Set(2, 1)

	result := bm.ColOr(0, 1)
	if result.Count() != 2 {
		t.Errorf("ColOr count = %d, want 2", result.Count())
	}
	if !result.Has(1) || !result.Has(2) {
		t.Error("ColOr missing expected entities")
	}
}

func TestColAndNot(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(1, 0)
	bm.Set(1, 1)
	bm.Set(2, 0)

	result := bm.ColAndNot(0, 1) // col0 but NOT col1
	if result.Count() != 1 {
		t.Errorf("ColAndNot count = %d, want 1", result.Count())
	}
	if !result.Has(2) {
		t.Error("ColAndNot missing entity 2")
	}
	if result.Has(1) {
		t.Error("ColAndNot contains entity 1 (has both cols)")
	}
}

func TestBitmap_Count_RespectsRows(t *testing.T) {
	bm := New(WithRows(100), WithCols(4))
	bm.SetColChunk(0, 0)
	bm.SetColChunk(1, 0)
	result := bm.ColAnd(0, 1)
	// Count must respect the 100-row boundary, not count the full chunk padding.
	if got := result.Count(); got != 100 {
		t.Errorf("Bitmap.Count = %d, want 100 (rows)", got)
	}
}

func TestBitmap_Count_MatchesForEach(t *testing.T) {
	bm := New(WithRows(100), WithCols(4))
	bm.SetColChunk(0, 0)
	result := bm.ColOr(0, 0)

	count := 0
	result.ForEach(func(row uint32) { count++ })
	if got := result.Count(); got != count {
		t.Errorf("Count() = %d, ForEach count = %d — should match", got, count)
	}
}

func TestBitmap_ForEach(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(0, 0)
	bm.Set(63, 0)
	bm.Set(64, 0)
	bm.Set(999_999, 0)

	result := bm.ColOr(0, 0)
	got := make(map[uint32]bool)
	result.ForEach(func(row uint32) { got[row] = true })

	for _, row := range []uint32{0, 63, 64, 999_999} {
		if !got[row] {
			t.Errorf("ForEach missed row %d", row)
		}
	}
	if len(got) != 4 {
		t.Errorf("ForEach visited %d rows, want 4", len(got))
	}
}
