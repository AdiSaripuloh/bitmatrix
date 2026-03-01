package bitmatrix

import "testing"

func FuzzSetHasRoundTrip(f *testing.F) {
	f.Add(uint32(0), uint32(0))
	f.Add(uint32(63), uint32(15))
	f.Add(uint32(999), uint32(63))

	bm := New(WithRows(1024), WithCols(64))

	f.Fuzz(func(t *testing.T, row, col uint32) {
		row %= 1024
		col %= 64
		bm.Set(row, col)
		if !bm.Has(row, col) {
			t.Errorf("Has(%d, %d) = false after Set", row, col)
		}
	})
}

func FuzzClearRoundTrip(f *testing.F) {
	f.Add(uint32(0), uint32(0))
	f.Add(uint32(63), uint32(15))
	f.Add(uint32(999), uint32(63))

	bm := New(WithRows(1024), WithCols(64))

	f.Fuzz(func(t *testing.T, row, col uint32) {
		row %= 1024
		col %= 64
		bm.Set(row, col)
		bm.Clear(row, col)
		if bm.Has(row, col) {
			t.Errorf("Has(%d, %d) = true after Clear", row, col)
		}
	})
}
